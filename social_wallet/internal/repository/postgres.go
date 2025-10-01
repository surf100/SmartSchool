package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"social-wallet/internal/entity"
	service "social-wallet/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) SaveTransaction(ctx context.Context, iin, schoolBin string, now time.Time, setSocPay string) error {
	socialPayment := false
	if setSocPay == "1" {
		socialPayment = true
	}

	query := `
		INSERT INTO external_susn_data (iin, school_bin, social_payment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (iin)
		DO UPDATE SET social_payment = $3, updated_at = $5
	`

	_, err := r.db.Exec(ctx, query, iin, schoolBin, socialPayment, now, now)
	return err
}

func (r *PostgresRepo) SyncSusnStatuses(ctx context.Context) error {
	oldPersons, err := r.GetAllPersons(ctx)
	if err != nil {
		return fmt.Errorf("failed to get old persons: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        UPDATE persons
        SET susn = true
        WHERE iin IN (
            SELECT iin FROM external_susn_data WHERE social_payment = '1'
        )
    `)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        UPDATE persons
        SET susn = false
        WHERE iin IN (
            SELECT iin FROM external_susn_data WHERE social_payment = '0'
        )
    `)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	newPersons, err := r.GetAllPersons(ctx)
	if err != nil {
		return fmt.Errorf("failed to get updated persons: %w", err)
	}

	oldMap := make(map[string]bool)
	for _, person := range oldPersons {
		oldMap[person.Pin] = person.Susn
	}

	for _, person := range newPersons {
		prevSusn, existed := oldMap[person.Pin]
		currSusn := person.Susn

		if !existed || prevSusn != currSusn {
			log.Printf("SUSN изменился: PIN=%s, Был=%v → Стал=%v", person.Pin, prevSusn, currSusn)

			if err := service.AssignAccessLevelsToPerson(ctx, person.Pin, currSusn); err != nil {
				log.Printf("Ошибка назначения уровней доступа для %s: %v", person.Pin, err)
			}
		}
	}

	return nil
}

func (r *PostgresRepo) FindPersonsByIIN(ctx context.Context, iin, schoolBin string) ([]*entity.Person, error) {
	query := `
		SELECT  
			pin, iin, school_bin, susn, 
			COALESCE(card_number, '') AS card_number
		FROM persons
		WHERE iin = $1 AND school_bin = $2
	`
	rows, err := r.db.Query(ctx, query, iin, schoolBin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []*entity.Person
	for rows.Next() {
		var p entity.Person
		err := rows.Scan(
			&p.Pin,
			&p.IIN,
			&p.SchoolBin,
			&p.Susn,
			&p.CardNumber,
		)
		if err != nil {
			return nil, err
		}
		persons = append(persons, &p)
	}
	return persons, nil
}
func (r *PostgresRepo) FindExternalSusnDataByIIN(ctx context.Context, iin, schoolBin string) ([]*entity.ExternalSusnData, error) {
	query := `
		SELECT id, iin, school_bin, social_payment, created_at, updated_at
		FROM external_susn_data
		WHERE iin = $1 AND school_bin = $2
	`
	rows, err := r.db.Query(ctx, query, iin, schoolBin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.ExternalSusnData
	for rows.Next() {
		var data entity.ExternalSusnData
		err := rows.Scan(&data.ID, &data.IIN, &data.SchoolBin, &data.SocialPayment, &data.CreatedAt, &data.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, &data)
	}
	return result, nil
}
func (r *PostgresRepo) FindExternalSusnDataByIINOnly(ctx context.Context, iin string) ([]*entity.ExternalSusnData, error) {
	query := `
        SELECT id, iin, school_bin, social_payment, created_at, updated_at
        FROM external_susn_data
        WHERE iin = $1
    `
	rows, err := r.db.Query(ctx, query, iin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.ExternalSusnData
	for rows.Next() {
		var data entity.ExternalSusnData
		err := rows.Scan(&data.ID, &data.IIN, &data.SchoolBin, &data.SocialPayment, &data.CreatedAt, &data.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, &data)
	}
	return result, nil
}

func (r *PostgresRepo) UpdateSchoolBinByIIN(ctx context.Context, iin, newSchoolBin string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE external_susn_data SET school_bin = $1, updated_at = NOW() WHERE iin = $2`,
		newSchoolBin, iin,
	)
	return err
}

func (r *PostgresRepo) GetPersonsWithSusn(ctx context.Context) ([]*entity.Person, error) {
	var persons []*entity.Person

	rows, err := r.db.Query(ctx, `
		SELECT pin, iin, school_bin, susn, card_number
		FROM persons
		WHERE susn = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p entity.Person
		if err := rows.Scan(&p.Pin, &p.IIN, &p.SchoolBin, &p.Susn, &p.CardNumber); err != nil {
			return nil, err
		}
		persons = append(persons, &p)
	}

	return persons, nil
}
func (r *PostgresRepo) GetPersonByPin(ctx context.Context, pin string) (*entity.Person, error) {
	query := `
		WITH cand AS (
			SELECT
				pin,
				-- cert_number теперь хранит ИИН
				COALESCE(NULLIF(cert_number, ''), '') AS iin,
				COALESCE(school_bin, '')  AS school_bin,
				COALESCE(susn, false)     AS susn,
				COALESCE(card_number, '') AS card_number,
				CASE
					WHEN TRIM(pin) = TRIM($1) THEN 0
					WHEN TRIM(COALESCE(cert_number, '')) = TRIM($1) THEN 1
					ELSE 2
				END AS rk
			FROM persons
			WHERE TRIM(pin) = TRIM($1)
			   OR TRIM(COALESCE(cert_number, '')) = TRIM($1)
		)
		SELECT pin, iin, school_bin, susn, card_number
		FROM cand
		ORDER BY rk
		LIMIT 1;
	`
	row := r.db.QueryRow(ctx, query, pin)

	var p entity.Person
	if err := row.Scan(&p.Pin, &p.IIN, &p.SchoolBin, &p.Susn, &p.CardNumber); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepo) GetAllPersons(ctx context.Context) ([]*entity.Person, error) {
	query := `
		SELECT  
			pin, iin, school_bin, susn, 
			COALESCE(card_number, '') AS card_number
		FROM persons
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []*entity.Person
	for rows.Next() {
		var p entity.Person
		err := rows.Scan(
			&p.Pin,
			&p.IIN,
			&p.SchoolBin,
			&p.Susn,
			&p.CardNumber,
		)
		if err != nil {
			return nil, err
		}
		persons = append(persons, &p)
	}
	return persons, nil
}

func (r *PostgresRepo) AssignCardNumberByIIN(ctx context.Context, iin, schoolBin, card string) error {
	// апсерт записи по ИИН+BIN; если запись есть — задаём карту ТОЛЬКО если её ещё нет
	const q = `
        INSERT INTO persons(iin, school_bin, card_number)
        VALUES ($1, $2, $3)
        ON CONFLICT (iin, school_bin) DO UPDATE
        SET card_number = COALESCE(persons.card_number, EXCLUDED.card_number)
    `
	_, err := r.db.Exec(ctx, q, iin, schoolBin, card)
	return err
}
