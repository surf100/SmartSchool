package handler

import (
	"context"
	"encoding/json"
	"time"

	"social-wallet/internal/entity"
	"social-wallet/internal/repository"
	"social-wallet/pkg/client"

	pb "social-wallet/api/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TerminalHandler struct {
	pb.UnimplementedTerminalEventServiceServer
	pb.UnimplementedSocialWalletServiceServer
	repo repository.Repository
}

func NewTerminalHandler(repo repository.Repository) *TerminalHandler {
	return &TerminalHandler{repo: repo}
}

type TransactionHandler struct {
	repo repository.Repository
}

func NewTransactionHandler(r repository.Repository) *TransactionHandler {
	return &TransactionHandler{repo: r}
}

func (h *TerminalHandler) HandleTransaction(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	parsedDate, err := time.Parse("2006-01-02 15:04:05", req.Date)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
	}

	setSocPay := ""
	if req.SetSocpay == "true" || req.SetSocpay == "1" {
		setSocPay = "1"
	}

	resetSocPay := ""
	if req.ResetSocpay == "true" || req.ResetSocpay == "1" {
		resetSocPay = "1"
	}

	err = h.repo.SaveTransaction(ctx, req.Iin, req.SchoolBin, parsedDate, setSocPay, resetSocPay)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save transaction: %v", err)
	}

	return &pb.TransactionResponse{Success: true}, nil
}

func (h *TerminalHandler) HandleTerminalEvent(ctx context.Context, req *pb.TerminalEventRequest) (*pb.TerminalEventResponse, error) {
	var outer map[string]any
	if err := json.Unmarshal([]byte(req.GetJsonPayload()), &outer); err != nil {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "неправильный формат JSON",
		}, nil
	}

	contentRaw, ok := outer["content"].(string)
	if !ok {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "не найдено поле content",
		}, nil
	}

	var content map[string]any
	if err := json.Unmarshal([]byte(contentRaw), &content); err != nil {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "неправильный формат поля content",
		}, nil
	}

	readerName, ok := content["readerName"].(string)
	if !ok || readerName == "" {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "readerName отсутствует или некорректен",
		}, nil
	}

	pin, ok := content["pin"].(string)
	if !ok || pin == "" {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "pin отсутствует или некорректен",
		}, nil
	}

	eventTimeMs, ok := content["eventTime"].(float64)
	if !ok {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "неправильный формат eventTime",
		}, nil
	}
	eventTime := time.UnixMilli(int64(eventTimeMs)).UTC()

	mapping, err := h.repo.GetMappingByReaderName(ctx, readerName)
	if err != nil {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "ошибка при поиске readerName в БД",
		}, nil
	}

	student, err := h.repo.GetStudentByPin(ctx, mapping.Pin)
	if err != nil {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "ошибка при поиске студента по pin",
		}, nil
	}

	if !student.SetSocPay {
		return &pb.TerminalEventResponse{
			Success: true,
			Message: "set_socpay = false, пропущено",
		}, nil
	}

	err = client.SendMealTransaction(student.IIN, eventTime, student.SchoolBin)
	if err != nil {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "ошибка отправки в соц. кошелек: " + err.Error(),
		}, nil
	}

	rawJson, err := json.Marshal(outer)
	if err != nil {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "ошибка сериализации JSON",
		}, nil
	}

	err = h.repo.SaveAccessEvent(ctx, &entity.AccessEvent{
		Pin:        pin,
		ReaderName: readerName,
		EventTime:  eventTime,
		RawJSON:    string(rawJson),
	})
	if err != nil {
		return &pb.TerminalEventResponse{
			Success: false,
			Message: "ошибка при сохранении события",
		}, nil
	}

	return &pb.TerminalEventResponse{
		Success: true,
		Message: "успешно обработано",
	}, nil
}
