package repository

import (
	"context"
	"time"

	"social-wallet/internal/entity"
)

type Repository interface {
	GetMappingByReaderName(ctx context.Context, readerName string) (*entity.AccessMapping, error)
	GetStudentByPin(ctx context.Context, pin string) (*entity.Student, error)
	SaveAccessEvent(ctx context.Context, event *entity.AccessEvent) error
	SaveTransaction(ctx context.Context, iin string, schoolBin string, date time.Time, set_socpay string, reset_socpay string) error
	SyncSusnStatuses(ctx context.Context) error

}
