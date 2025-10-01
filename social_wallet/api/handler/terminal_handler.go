package handler

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"social-wallet/internal/repository"

	pb "social-wallet/api/proto/gen"

	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type TerminalHandler struct {
	pb.UnimplementedTerminalServiceServer
	repo repository.Repository
}

func NewTerminalHandler(repo repository.Repository) *TerminalHandler {
	return &TerminalHandler{repo: repo}
}
func genCardSCH() (string, error) {
	const n = 8
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := 0; i < n; i++ {
		buf[i] = '0' + (buf[i] % 10)
	}
	return "SCH" + string(buf), nil
}

func (s *TerminalHandler) ActivateVoucher(ctx context.Context, req *pb.ActivateVoucherRequest) (*pb.ActivateVoucherResponse, error) {
	iin := req.GetIin()
	schoolBin := req.GetSchoolBin()

	if iin == "" || schoolBin == "" {
		return nil, status.Errorf(codes.InvalidArgument, "code: 10, message: Нет входных параметров")
	}
	if len(iin) != 12 || !isDigitsOnly(iin) {
		return nil, status.Errorf(codes.InvalidArgument, "code: 30, message: Неправильный формат ИИН")
	}

	// нормализуем setSocPay к "1"/"0"
	setSocPay := ""
	switch req.SetSocpay {
	case "true", "1":
		setSocPay = "1"
	case "false", "0":
		setSocPay = "0"
	}

	// reset_socpay имеет приоритет над set_socpay
	if req.ResetSocpay != nil {
		switch *req.ResetSocpay {
		case "true", "1":
			setSocPay = "0"
		}
	}

	exist := "0"
	social_payments := "0"

	// 1) ищем точное совпадение IIN+BIN
	records, err := s.repo.FindExternalSusnDataByIIN(ctx, iin, schoolBin)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "code: 20, message: Нет ответа от сервера: %v", err)
	}

	// 2) если пусто — ищем по одному IIN и обновляем BIN
	if len(records) == 0 {
		recordsByIIN, err := s.repo.FindExternalSusnDataByIINOnly(ctx, iin)
		if err == nil && len(recordsByIIN) > 0 {
			if errUpd := s.repo.UpdateSchoolBinByIIN(ctx, iin, schoolBin); errUpd == nil {
				records = recordsByIIN
			} else {
				fmt.Printf("Не удалось обновить school_bin для IIN=%s: %v\n", iin, errUpd)
			}
		}
	}

	// 3) если есть записи — exist=1 и берём текущее social_payment
	if len(records) > 0 {
		exist = "1"
		for _, r := range records {
			if r.SocialPayment {
				social_payments = "1"
				break
			}
		}
	}

	// 4) апсёрт транзакции с новым значением social_payment
	now := time.Now()
	if err := s.repo.SaveTransaction(ctx, iin, schoolBin, now, setSocPay); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save external_susn_data: %v", err)
	}

	// 5) В ответ social_payments всегда по setSocPay (если оно передано)
	if setSocPay != "" {
		social_payments = setSocPay
	}

	// 6) card_number из persons или сгенерированный SCH########
	card_number := ""
	persons, err := s.repo.FindPersonsByIIN(ctx, iin, schoolBin)
	if err == nil && len(persons) > 0 && persons[0].CardNumber != "" {
		card_number = persons[0].CardNumber
	} else {
		if num, gErr := genCardSCH(); gErr == nil {
			card_number = num
		} else {
			fmt.Printf("gen card error: %v\n", gErr)
		}
	}

	// 7) ответ
	resp := &pb.ActivateVoucherResponse{
		Result:         "data",
		Exist:          exist,
		SocialPayments: social_payments,
		CardNumber:     card_number,
	}
	if exist == "2" {
		resp.Warning = proto.String("1")
		resp.WarnComment = proto.String("Данный ИИН уже существует")
	}

	return resp, nil
}

func isDigitsOnly(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
