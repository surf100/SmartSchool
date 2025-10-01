package repository

import (
	"context"
	"time"

	"social-wallet/internal/entity"
)

type Repository interface {
	SaveTransaction(ctx context.Context, iin, schoolBin string, now time.Time, setSocPay string) error
	SyncSusnStatuses(ctx context.Context) error
	FindPersonsByIIN(ctx context.Context, iin, schoolBin string) ([]*entity.Person, error)
	FindExternalSusnDataByIIN(ctx context.Context, iin, schoolBin string) ([]*entity.ExternalSusnData, error)
	GetPersonByPin(ctx context.Context, pin string) (*entity.Person, error)
	FindExternalSusnDataByIINOnly(ctx context.Context, iin string) ([]*entity.ExternalSusnData, error)
	UpdateSchoolBinByIIN(ctx context.Context, iin, newSchoolBin string) error
	AssignCardNumberByIIN(ctx context.Context, iin, schoolBin, card string) error
}
