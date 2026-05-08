package application

import (
	"attendance-api/internal/domain"
	"context"
)

type AttendanceService interface {
	ProcessZKBioEvent(ctx context.Context, request *domain.ZKBioPushRequest) error
	GetEventsByPin(ctx context.Context, pin string, limit, offset int) ([]*domain.AttendanceEvent, error)
	GetEventsByDeviceSN(ctx context.Context, deviceSN string, limit, offset int) ([]*domain.AttendanceEvent, error)
	GetAttendanceEvent(ctx context.Context, id int) (*domain.AttendanceEvent, error)
	ListAttendanceEvents(ctx context.Context, limit, offset int) ([]*domain.AttendanceEvent, error)
} 