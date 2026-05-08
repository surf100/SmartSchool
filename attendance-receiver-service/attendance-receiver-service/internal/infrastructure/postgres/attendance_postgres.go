package postgres

import (
	"attendance-api/internal/domain"
	"attendance-api/internal/domain/repository"
	"context"
	"database/sql"
	"errors"
	"log"
)

type attendanceRepository struct {
	db *sql.DB
}

func NewAttendanceRepository(db *sql.DB) repository.AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) CreateAttendanceEvent(ctx context.Context, event *domain.AttendanceEvent) (*domain.AttendanceEvent, error) {
	row := r.db.QueryRowContext(ctx, `
  INSERT INTO attendancelog (
    pin, name, last_name, dept_code, dept_name, reader_name, door_name,
    device_sn, capture_photo_path, verify_mode_name, event_name,
    event_time, raw_payload, unique_key, log_id, created_at
  )
  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
  ON CONFLICT (unique_key) DO NOTHING
  RETURNING id
`, event.Pin, event.Name, event.LastName, event.DeptCode, event.DeptName,
		event.ReaderName, event.DoorName, event.DeviceSN, event.CapturePhotoPath,
		event.VerifyModeName, event.EventName, event.EventTime, event.RawPayload,
		event.UniqueKey, event.LogID, event.CreatedAt)

	if err := row.Scan(&event.ID); err != nil {
		if err == sql.ErrNoRows {
			// дубликат — запись уже есть
			log.Printf("CreateAttendanceEvent: duplicate unique_key=%s", event.UniqueKey)
			return event, nil
		}
		return event, err
	}
	return event, nil

}

func (r *attendanceRepository) GetAttendanceEvent(ctx context.Context, id int) (*domain.AttendanceEvent, error) {
	var event domain.AttendanceEvent
	row := r.db.QueryRowContext(ctx,
		`SELECT id, pin, name, last_name, dept_code, dept_name, reader_name, door_name, device_sn, capture_photo_path, verify_mode_name, event_name, event_time, raw_payload, unique_key, log_id, created_at 
		 FROM attendancelog WHERE id=$1`, id)
	err := row.Scan(&event.ID, &event.Pin, &event.Name, &event.LastName, &event.DeptCode, &event.DeptName, &event.ReaderName, &event.DoorName, &event.DeviceSN, &event.CapturePhotoPath, &event.VerifyModeName, &event.EventName, &event.EventTime, &event.RawPayload, &event.UniqueKey, &event.LogID, &event.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("attendance event not found")
	}
	return &event, err
}

func (r *attendanceRepository) GetEventsByPin(ctx context.Context, pin string, limit, offset int) ([]*domain.AttendanceEvent, error) {
	query := `SELECT id, pin, name, last_name, dept_code, dept_name, reader_name, door_name, device_sn, capture_photo_path, verify_mode_name, event_name, event_time, raw_payload, unique_key, log_id, created_at 
		 FROM attendancelog WHERE pin=$1 ORDER BY event_time DESC LIMIT $2 OFFSET $3`

	log.Printf("GetEventsByPin: query=%s, pin=%s, limit=%d, offset=%d", query, pin, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, pin, limit, offset)
	if err != nil {
		log.Printf("GetEventsByPin: database error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var events []*domain.AttendanceEvent
	for rows.Next() {
		var event domain.AttendanceEvent
		if err := rows.Scan(&event.ID, &event.Pin, &event.Name, &event.LastName, &event.DeptCode, &event.DeptName, &event.ReaderName, &event.DoorName, &event.DeviceSN, &event.CapturePhotoPath, &event.VerifyModeName, &event.EventName, &event.EventTime, &event.RawPayload, &event.UniqueKey, &event.LogID, &event.CreatedAt); err != nil {
			log.Printf("GetEventsByPin: scan error: %v", err)
			return nil, err
		}
		events = append(events, &event)
	}

	log.Printf("GetEventsByPin: found %d events for pin %s", len(events), pin)
	return events, nil
}

func (r *attendanceRepository) GetEventsByDeviceSN(ctx context.Context, deviceSN string, limit, offset int) ([]*domain.AttendanceEvent, error) {
	query := `SELECT id, pin, name, last_name, dept_code, dept_name, reader_name, door_name, device_sn, capture_photo_path, verify_mode_name, event_name, event_time, raw_payload, unique_key, log_id, created_at 
		 FROM attendancelog WHERE device_sn=$1 ORDER BY event_time DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, deviceSN, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.AttendanceEvent
	for rows.Next() {
		var event domain.AttendanceEvent
		if err := rows.Scan(&event.ID, &event.Pin, &event.Name, &event.LastName, &event.DeptCode, &event.DeptName, &event.ReaderName, &event.DoorName, &event.DeviceSN, &event.CapturePhotoPath, &event.VerifyModeName, &event.EventName, &event.EventTime, &event.RawPayload, &event.UniqueKey, &event.LogID, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, nil
}

func (r *attendanceRepository) ListAttendanceEvents(ctx context.Context, limit, offset int) ([]*domain.AttendanceEvent, error) {
	query := `SELECT id, pin, name, last_name, dept_code, dept_name, reader_name, door_name, device_sn, capture_photo_path, verify_mode_name, event_name, event_time, raw_payload, unique_key, log_id, created_at 
		 FROM attendancelog ORDER BY event_time DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.AttendanceEvent
	for rows.Next() {
		var event domain.AttendanceEvent
		if err := rows.Scan(&event.ID, &event.Pin, &event.Name, &event.LastName, &event.DeptCode, &event.DeptName, &event.ReaderName, &event.DoorName, &event.DeviceSN, &event.CapturePhotoPath, &event.VerifyModeName, &event.EventName, &event.EventTime, &event.RawPayload, &event.UniqueKey, &event.LogID, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, nil
}

func (r *attendanceRepository) GetEventByUniqueKey(ctx context.Context, uniqueKey string) (*domain.AttendanceEvent, error) {
	var event domain.AttendanceEvent
	row := r.db.QueryRowContext(ctx,
		`SELECT id, pin, name, last_name, dept_code, dept_name, reader_name, door_name, device_sn, capture_photo_path, verify_mode_name, event_name, event_time, raw_payload, unique_key, log_id, created_at 
		 FROM attendancelog WHERE unique_key=$1`, uniqueKey)
	err := row.Scan(&event.ID, &event.Pin, &event.Name, &event.LastName, &event.DeptCode, &event.DeptName, &event.ReaderName, &event.DoorName, &event.DeviceSN, &event.CapturePhotoPath, &event.VerifyModeName, &event.EventName, &event.EventTime, &event.RawPayload, &event.UniqueKey, &event.LogID, &event.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &event, err
}

func (r *attendanceRepository) GetPersonByPin(ctx context.Context, pin string) (*domain.Person, error) {
	const q = `
		SELECT pin, iin, school_bin, susn
		FROM persons
		WHERE pin = $1
		LIMIT 1
	`
	p := &domain.Person{}
	err := r.db.QueryRowContext(ctx, q, pin).Scan(&p.Pin, &p.IIN, &p.SchoolBIN, &p.SUSN)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}
