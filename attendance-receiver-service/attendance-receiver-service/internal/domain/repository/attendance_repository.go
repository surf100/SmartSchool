package repository

import (
	"attendance-api/internal/domain"
	"context"
)

type AttendanceRepository interface {
	CreateAttendanceEvent(ctx context.Context, event *domain.AttendanceEvent) (*domain.AttendanceEvent, error)
	GetAttendanceEvent(ctx context.Context, id int) (*domain.AttendanceEvent, error)
	GetEventsByPin(ctx context.Context, pin string, limit, offset int) ([]*domain.AttendanceEvent, error)
	GetEventsByDeviceSN(ctx context.Context, deviceSN string, limit, offset int) ([]*domain.AttendanceEvent, error)
	ListAttendanceEvents(ctx context.Context, limit, offset int) ([]*domain.AttendanceEvent, error)
	GetEventByUniqueKey(ctx context.Context, uniqueKey string) (*domain.AttendanceEvent, error)
	GetPersonByPin(ctx context.Context, pin string) (*domain.Person, error)
}
