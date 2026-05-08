package postgres

import (
	"context"

	"person-dumper/internal/domain"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ DB *pgxpool.Pool }

func New(dsn string) (*Repo, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &Repo{DB: pool}, nil
}


func (r *Repo) ListEnabled() ([]domain.School, error) {
	rows, err := r.DB.Query(context.Background(),
		`select id, bin, db_schema, device_ip::text, device_port, is_enabled
		   from public.schools
		  where is_enabled = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.School
	for rows.Next() {
		var s domain.School
		if err := rows.Scan(&s.ID, &s.BIN, &s.Schema, &s.DeviceIP, &s.Port, &s.Enabled); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repo) SchoolByID(id int64) (domain.School, error) {
	row := r.DB.QueryRow(context.Background(),
		`select id, bin, db_schema, device_ip::text, device_port, is_enabled
		   from public.schools where id=$1`, id)
	var s domain.School
	if err := row.Scan(&s.ID, &s.BIN, &s.Schema, &s.DeviceIP, &s.Port, &s.Enabled); err != nil {
		return domain.School{}, err
	}
	return s, nil
}

func (r *Repo) SchoolByBIN(bin string) (domain.School, error) {
	row := r.DB.QueryRow(context.Background(),
		`select id, bin, db_schema, device_ip::text, device_port, is_enabled
		   from public.schools where bin=$1`, bin)
	var s domain.School
	if err := row.Scan(&s.ID, &s.BIN, &s.Schema, &s.DeviceIP, &s.Port, &s.Enabled); err != nil {
		return domain.School{}, err
	}
	return s, nil
}



func (r *Repo) GetSyncRuns(ctx context.Context, schoolID *int64, limit int) ([]domain.SyncRunView, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows pgx.Rows
	var err error
	if schoolID != nil {
		rows, err = r.DB.Query(ctx, `
		  select run_id, school_id, started_at, finished_at, total, inserted, updated, failed, coalesce(ok,false)
		    from public.sync_runs
		   where school_id=$1
		   order by id desc
		   limit $2`, *schoolID, limit)
	} else {
		rows, err = r.DB.Query(ctx, `
		  select run_id, school_id, started_at, finished_at, total, inserted, updated, failed, coalesce(ok,false)
		    from public.sync_runs
		   order by id desc
		   limit $1`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SyncRunView
	for rows.Next() {
		var v domain.SyncRunView
		if err := rows.Scan(&v.RunID, &v.SchoolID, &v.StartedAt, &v.FinishedAt, &v.Total, &v.Inserted, &v.Updated, &v.Failed, &v.OK); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *Repo) BeginRun(runID string, schoolID int64) error {
	_, err := r.DB.Exec(context.Background(),
		`insert into public.sync_runs (run_id, school_id) values ($1,$2)`, runID, schoolID)
	return err
}

func (r *Repo) FinishRun(runID string, schoolID int64, st domain.SyncStats, ok bool) error {
	_, err := r.DB.Exec(context.Background(),
		`update public.sync_runs
		     set finished_at = now(),
		         total      = $3,
		         inserted   = $4,
		         updated    = $5,
		         failed     = $6,
		         ok         = $7
		   where run_id = $1 and school_id = $2`,
		runID, schoolID, st.Total, st.Inserted, st.Updated, st.Failed, ok)
	return err
}

func (r *Repo) DeletePinsNotIn(ctx context.Context, schema string, pins []string) (int64, error) {
	if len(pins) == 0 {
		return 0, nil
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()


	if _, err = tx.Exec(ctx, `select set_config('search_path', $1, true)`, schema); err != nil {
		return 0, err
	}


	tag, err := tx.Exec(ctx, `delete from persons where pin is not null and not (pin = any($1))`, pins)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
