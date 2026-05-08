package domain

import (
	"context"
	"time"
)

type School struct {
	ID       int64
	BIN      string
	Schema   string
	DeviceIP string
	Port     int
	Enabled  bool
}

type SchoolRepo interface {
	ListEnabled() ([]School, error)
}

type SyncStats struct {
	Total, Inserted, Updated, Failed int
}

type RunLogger interface {
	BeginRun(runID string, schoolID int64) error
	FinishRun(runID string, schoolID int64, st SyncStats, ok bool) error
}

type DeviceUser struct {
	ExternalID string
	FullName   string
	CardNo     string
	Department string
	DeptCode   string
	Role       string
	Phone      string
}

type DeviceClient interface {
	Ping(ip string, port int) error
	GetUsers(ip string, port int) ([]DeviceUser, error)
}

type Tenant struct {
	BIN    string
	Schema string
}


type PersonClient interface {
	Upsert(ctx context.Context, t Tenant, src string, u DeviceUser) (created bool, out *Person, err error)
}


type Person struct {
	ID         int64
	ExternalID string
	FullName   string
	CardNo     string
}

type SyncRunView struct {
	RunID                            string
	SchoolID                         int64
	StartedAt                        time.Time
	FinishedAt                       time.Time
	Total, Inserted, Updated, Failed int
	OK                               bool
}

type RunReader interface {
	GetSyncRuns(ctx context.Context, schoolID *int64, limit int) ([]SyncRunView, error)
}

type SchoolLookup interface {
	SchoolByID(id int64) (School, error)
	SchoolByBIN(bin string) (School, error)
}
