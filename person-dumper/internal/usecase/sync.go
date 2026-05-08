package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"person-dumper/internal/domain"
	"person-dumper/internal/zkbio"
)

type SyncUC struct {
	Schools domain.SchoolRepo
	Lookup  domain.SchoolLookup
	Runs    domain.RunLogger
	Reads   domain.RunReader
	Dev     domain.DeviceClient
	Persons domain.PersonClient 
}

func (u *SyncUC) TriggerAll(ctx context.Context, dry bool) (runID string, scheduled int, err error) {
	runID = genRunID()

	schools, err := u.Schools.ListEnabled()
	if err != nil {
		return "", 0, err
	}
	scheduled = len(schools)

	for _, s := range schools {
		_ = u.Runs.BeginRun(runID, s.ID)

		targetSchema := strings.TrimSpace(s.Schema)
		if targetSchema == "" {
			prefix := os.Getenv("SCHEMA_PREFIX")
			if prefix == "" {
				prefix = "school_"
			}
			targetSchema = fmt.Sprintf("%s%s", prefix, s.BIN) 
		}
		slog.Info("school→schema", "school_id", s.ID, "bin", s.BIN, "schema", targetSchema)
		tenant := domain.Tenant{BIN: s.BIN, Schema: targetSchema}


		st := domain.SyncStats{}

		// ---- ZKBio API ----
		if os.Getenv("ZK_MODE") == "api" || os.Getenv("ZK_BASE_URL") != "" {
			base := os.Getenv("ZK_BASE_URL")
			token := os.Getenv("ZK_ACCESS_TOKEN")
			if base == "" || token == "" {
				slog.Error("zkbio env missing", "base", base != "", "token", token != "")
				_ = u.Runs.FinishRun(runID, s.ID, st, false)
				continue
			}

			dcs := deptCodesForBIN(s.BIN) 
			slog.Info("zkbio mode", "base", base, "deptCodes", strings.Join(dcs, ","), "tls_insecure", zkInsecure())

			cli := zkbio.New(base, token, zkInsecure(), zkPageSize())
			users, err := cli.ListUsers(ctx, dcs)
			if err != nil {
				slog.Error("zkbio list failed", "bin", s.BIN, "err", err)
				st.Failed++
				_ = u.Runs.FinishRun(runID, s.ID, st, false)
				continue
			}
			st.Total = len(users)

			wantedPins := make(map[string]struct{}, len(users))
			if !dry {
				for _, du := range users {
					if du.ExternalID == "" {
						st.Failed++
						continue
					}
					wantedPins[du.ExternalID] = struct{}{}

					created, out, err := u.Persons.Upsert(ctx, tenant, "zkbio", du)
					if err != nil {
						st.Failed++
						continue
					}
					if created {
						st.Inserted++
					} else if out != nil {
						st.Updated++
					}
				}
				
				u.reconcileDeletesDB(ctx, tenant.Schema, wantedPins, &st)

			}

			ok := (st.Failed == 0)
			_ = u.Runs.FinishRun(runID, s.ID, st, ok)
			slog.Info("UC: run FINISH", "runId", runID, "stats", st, "ok", ok)
			continue
		}
		// ---- /ZKBio API ----

		mode := "real"
		if os.Getenv("ZK_DEMO") == "1" {
			mode = "demo"
		}
		if p := os.Getenv("ZK_USERS_FILE"); p != "" {
			mode = "file:" + p
		}
		slog.Info("device source", "mode", mode, "ip", s.DeviceIP, "port", s.Port)

		// ping
		if err := u.Dev.Ping(s.DeviceIP, s.Port); err != nil {
			slog.Error("device ping failed", "school_id", s.ID, "bin", s.BIN, "ip", s.DeviceIP, "port", s.Port, "err", err)
			st.Failed++
			_ = u.Runs.FinishRun(runID, s.ID, st, false)
			continue
		}

		// get users
		users, e := u.Dev.GetUsers(s.DeviceIP, s.Port)
		if e != nil {
			slog.Error("get users failed", "school_id", s.ID, "bin", s.BIN, "err", e)
			_ = u.Runs.FinishRun(runID, s.ID, st, false)
			continue
		}
		slog.Info("got users", "school_id", s.ID, "count", len(users))
		st.Total = len(users)

		wantedPins := make(map[string]struct{}, len(users))
		if !dry {
			for _, du := range users {
				if du.ExternalID != "" {
					wantedPins[du.ExternalID] = struct{}{}
				}
				created, out, e := u.Persons.Upsert(ctx, tenant, "zkteco", du)
				if e != nil {
					slog.Error("upsert failed", "school_id", s.ID, "bin", s.BIN, "schema", targetSchema, "err", e)
					st.Failed++
					continue
				}
				if created {
					st.Inserted++
				} else if out != nil {
					st.Updated++
				}
			}
			// удаление отсутствующих (reconcile)
			u.reconcileDeletesDB(ctx, tenant.Schema, wantedPins, &st)

		}

		ok := (st.Failed == 0)
		_ = u.Runs.FinishRun(runID, s.ID, st, ok)
		slog.Info("UC: run FINISH", "runId", runID, "stats", st, "ok", ok)
	}
	return runID, scheduled, nil
}

func (u *SyncUC) TriggerSchoolSync(ctx context.Context, schoolID *int64, bin *string, dry bool) (string, error) {
	slog.Info("UC: TriggerSchoolSync ENTER", "schoolId", schoolID, "bin", bin, "dry", dry)

	// 0) найдём школу
	schools, err := u.Schools.ListEnabled()
	if err != nil {
		return "", err
	}
	var s domain.School
	found := false
	for _, it := range schools {
		if schoolID != nil && it.ID == *schoolID {
			s = it
			found = true
			break
		}
		if bin != nil && *bin != "" && strings.EqualFold(it.BIN, *bin) {
			s = it
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("school not found")
	}
	slog.Info("UC: school RESOLVED", "id", s.ID, "bin", s.BIN, "schema", s.Schema)

	// --- выбор схемы by-school ---
	targetSchema := strings.TrimSpace(s.Schema)
	if targetSchema == "" {
		prefix := os.Getenv("SCHEMA_PREFIX")
		if prefix == "" {
			prefix = "school_"
		}
		targetSchema = fmt.Sprintf("%s%s", prefix, s.BIN)
	}
	slog.Info("school→schema", "school_id", s.ID, "bin", s.BIN, "schema", targetSchema)
	tenant := domain.Tenant{BIN: s.BIN, Schema: targetSchema}

	// 1) старт прогона
	runID := genRunID()
	if err := u.Runs.BeginRun(runID, s.ID); err != nil {
		slog.Error("UC: begin run failed", "err", err)
		return "", err
	}
	slog.Info("UC: run STARTED", "runId", runID, "schoolId", s.ID)

	st := domain.SyncStats{}

	// ---- ZKBio API ----
	if os.Getenv("ZK_MODE") == "api" || os.Getenv("ZK_BASE_URL") != "" {
		base := os.Getenv("ZK_BASE_URL")
		token := os.Getenv("ZK_ACCESS_TOKEN")
		if base == "" || token == "" {
			slog.Error("zkbio env missing", "base_set", base != "", "token_set", token != "")
			_ = u.Runs.FinishRun(runID, s.ID, st, false)
			return runID, fmt.Errorf("zkbio: env not set")
		}
		dcs := deptCodesForBIN(s.BIN)
		slog.Info("zkbio mode", "base", base, "deptCodes", strings.Join(dcs, ","), "tls_insecure", zkInsecure())

		cli := zkbio.New(base, token, zkInsecure(), zkPageSize())
		users, err := cli.ListUsers(ctx, dcs)
		if err != nil {
			slog.Error("zkbio list failed", "bin", s.BIN, "err", err)
			st.Failed++
			_ = u.Runs.FinishRun(runID, s.ID, st, false)
			return runID, nil
		}
		st.Total = len(users)

		wantedPins := make(map[string]struct{}, len(users))
		if !dry {
			for _, du := range users {
				if du.ExternalID == "" {
					st.Failed++
					continue
				}
				wantedPins[du.ExternalID] = struct{}{}

				created, out, e := u.Persons.Upsert(ctx, tenant, "zkbio", du)
				if e != nil {
					st.Failed++
					continue
				}
				if created {
					st.Inserted++
				} else if out != nil {
					st.Updated++
				}
			}
			u.reconcileDeletesDB(ctx, tenant.Schema, wantedPins, &st)

		}
		ok := (st.Failed == 0)
		_ = u.Runs.FinishRun(runID, s.ID, st, ok)
		slog.Info("UC: run FINISH", "runId", runID, "stats", st, "ok", ok)
		return runID, nil
	}
	// ---- /ZKBio API ----

	// ---- device-ветка ----
	mode := "real"
	if os.Getenv("ZK_DEMO") == "1" {
		mode = "demo"
	}
	if p := os.Getenv("ZK_USERS_FILE"); p != "" {
		mode = "file:" + p
	}
	slog.Info("device source", "mode", mode, "ip", s.DeviceIP, "port", s.Port)

	if err := u.Dev.Ping(s.DeviceIP, s.Port); err != nil {
		slog.Error("device ping failed", "school_id", s.ID, "bin", s.BIN, "ip", s.DeviceIP, "port", s.Port, "err", err)
		st.Failed++
		_ = u.Runs.FinishRun(runID, s.ID, st, false)
		return runID, nil
	}

	users, e := u.Dev.GetUsers(s.DeviceIP, s.Port)
	if e != nil {
		slog.Error("get users failed", "school_id", s.ID, "bin", s.BIN, "err", e)
		_ = u.Runs.FinishRun(runID, s.ID, st, false)
		return runID, nil
	}
	slog.Info("got users", "school_id", s.ID, "count", len(users))
	st.Total = len(users)

	wantedPins := make(map[string]struct{}, len(users))
	if !dry {
		for _, du := range users {
			if du.ExternalID != "" {
				wantedPins[du.ExternalID] = struct{}{}
			}
			created, out, e := u.Persons.Upsert(ctx, tenant, "zkteco", du)
			if e != nil {
				st.Failed++
				continue
			}
			if created {
				st.Inserted++
			} else if out != nil {
				st.Updated++
			}
		}
		u.reconcileDeletesDB(ctx, tenant.Schema, wantedPins, &st)

	}
	ok := (st.Failed == 0)
	_ = u.Runs.FinishRun(runID, s.ID, st, ok)
	slog.Info("UC: run FINISH", "runId", runID, "stats", st, "ok", ok)
	return runID, nil
}

func (u *SyncUC) GetRuns(ctx context.Context, schoolID *int64, limit int) ([]domain.SyncRunView, error) {
	return u.Reads.GetSyncRuns(ctx, schoolID, limit)
}

func genRunID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// ---- reconcile helpers ----

type personListerDeleter interface {
	ListPins(ctx context.Context, tenant domain.Tenant) (map[string]int64, error)
	DeleteByID(ctx context.Context, tenant domain.Tenant, id int64) error
}

func (u *SyncUC) reconcileDeletes(ctx context.Context, tenant domain.Tenant, wantedPins map[string]struct{}, st *domain.SyncStats) {
	ld, ok := u.Persons.(personListerDeleter)
	if !ok {
		slog.Warn("reconcile skipped: PersonClient doesn't support ListPins/DeleteByID")
		return
	}

	dbPins, err := ld.ListPins(ctx, tenant)
	if err != nil {
		slog.Error("reconcile: list pins failed", "schema", tenant.Schema, "err", err)
		return
	}

	deleted := 0
	for pin, id := range dbPins {
		if _, keep := wantedPins[pin]; !keep {
			if err := ld.DeleteByID(ctx, tenant, id); err != nil {
				slog.Error("reconcile: delete failed", "schema", tenant.Schema, "pin", pin, "id", id, "err", err)
				st.Failed++ 
				continue
			}
			deleted++
		}
	}
	if deleted > 0 {
		slog.Info("reconcile: deleted", "schema", tenant.Schema, "count", deleted)
	}
}

//  утилитные хелперы
func deptCodesForBIN(bin string) []string {
	raw := os.Getenv("ZK_SCHOOL_DEPTCODES")
	if raw == "" {
		return nil 
	}
	type m map[string][]string
	var mm m
	if err := json.Unmarshal([]byte(raw), &mm); err != nil {
		slog.Warn("bad ZK_SCHOOL_DEPTCODES JSON", "err", err)
		return nil
	}
	return mm[bin]
}

func zkInsecure() bool {
	v := strings.ToLower(os.Getenv("ZK_TLS_INSECURE"))
	return v == "1" || v == "true" || v == "yes"
}
func zkPageSize() int { return 200 }

// reconcileDeletesDB удаляет из <schema>.persons всех, чьих PIN нет в wantedPins.
func (u *SyncUC) reconcileDeletesDB(ctx context.Context, schema string, wantedPins map[string]struct{}, st *domain.SyncStats) {
	pins := make([]string, 0, len(wantedPins))
	for p := range wantedPins {
		if p != "" {
			pins = append(pins, p)
		}
	}
	if len(pins) == 0 {
		slog.Warn("reconcile skipped: empty wantedPins (source returned no users)")
		return
	}

	type deleter interface {
		DeletePinsNotIn(ctx context.Context, schema string, pins []string) (int64, error)
	}
	d, ok := any(u.Schools).(deleter)
	if !ok {
		slog.Warn("reconcile skipped: repo does not implement DeletePinsNotIn")
		return
	}

	n, err := d.DeletePinsNotIn(ctx, schema, pins)
	if err != nil {
		slog.Error("reconcile delete failed", "schema", schema, "err", err)
		return
	}
	if n > 0 {
		slog.Info("reconcile: deleted", "schema", schema, "count", n)
	}
}
