package repository

import (
	"context"
	"time"

	"social-wallet/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) GetMappingByReaderName(ctx context.Context, readerName string) (*entity.AccessMapping, error) {
	var pin string
	err := r.db.QueryRow(ctx, "SELECT pin FROM users WHERE reader_name=$1", readerName).Scan(&pin)
	if err != nil {
		return nil, err
	}
	return &entity.AccessMapping{ReaderName: readerName, Pin: pin}, nil
}

func (r *PostgresRepo) GetStudentByPin(ctx context.Context, pin string) (*entity.Student, error) {
	var s entity.Student
	err := r.db.QueryRow(ctx, `
        SELECT pin, iin, school_bin, set_socpay
        FROM users WHERE pin=$1
    `, pin).Scan(&s.Pin, &s.IIN, &s.SchoolBin, &s.SetSocPay)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepo) SaveAccessEvent(ctx context.Context, event *entity.AccessEvent) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO access_events (pin, reader_name, event_time, raw_json)
        VALUES ($1, $2, $3, $4)
    `, event.Pin, event.ReaderName, event.EventTime, event.RawJSON)
	return err
}

func (r *PostgresRepo) SaveTransaction(ctx context.Context, iin string, schoolBin string, date time.Time, set_socpay string, reset_socpay string) error {
	query := `
        INSERT INTO external_susn_data (iin, school_bin, date, set_socpay, reset_socpay)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.Exec(ctx, query, iin, schoolBin, date, set_socpay, reset_socpay)
	return err
}
func (r *PostgresRepo) SyncSusnStatuses(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        UPDATE users
        SET susn = true
        WHERE iin IN (
            SELECT iin FROM external_susn_data WHERE set_socpay = '1'
        )
    `)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        UPDATE users
        SET susn = false
        WHERE iin IN (
            SELECT iin FROM external_susn_data WHERE set_socpay = '0' AND reset_socpay = '1'
        )
    `)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
