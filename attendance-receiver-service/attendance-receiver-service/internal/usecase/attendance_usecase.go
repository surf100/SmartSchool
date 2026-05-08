package usecase

import (
	"attendance-api/internal/domain"
	"attendance-api/internal/domain/application"
	"attendance-api/internal/domain/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type attendanceService struct {
	repo repository.AttendanceRepository
}

func NewAttendanceService(repo repository.AttendanceRepository) application.AttendanceService {
	return &attendanceService{
		repo: repo,
	}
}

func isCanteenSW(readerName, doorName string) bool {
	rn := strings.ToLower(readerName)
	dn := strings.ToLower(doorName)
	return rn == "canteensw" || strings.Contains(rn, "canteensw") || strings.Contains(dn, "canteensw")
}

func (s *attendanceService) ProcessZKBioEvent(ctx context.Context, request *domain.ZKBioPushRequest) error {
	log.Printf("Processing ZKBio event: module=%s, pushType=%s", request.Module, request.PushType)

	var content domain.ZKBioEventContent
	if err := json.Unmarshal([]byte(request.Content), &content); err != nil {
		return fmt.Errorf("failed to parse content JSON: %w", err)
	}
	log.Printf("DEBUG: reader=%q door=%q isCanteen=%v",
		content.ReaderName, content.DoorName, isCanteenSW(content.ReaderName, content.DoorName))

	existingEvent, err := s.repo.GetEventByUniqueKey(ctx, content.UniqueKey)
	if err == nil && existingEvent != nil {
		log.Printf("Event with unique_key %s already exists, skipping", content.UniqueKey)
		return nil
	}

	eventTime := time.Unix(content.EventTime/1000, (content.EventTime%1000)*1_000_000)

	event := &domain.AttendanceEvent{
		Pin:              content.Pin,
		Name:             content.Name,
		LastName:         content.LastName,
		DeptCode:         content.DeptCode,
		DeptName:         content.DeptName,
		ReaderName:       content.ReaderName,
		DoorName:         content.DoorName,
		DeviceSN:         content.DevSN, 
		CapturePhotoPath: content.CapturePhotoPath,
		VerifyModeName:   content.VerifyModeName,
		EventName:        content.EventName,
		EventTime:        eventTime,
		RawPayload:       request.Content,
		UniqueKey:        content.UniqueKey,
		LogID:            content.LogID,
		CreatedAt:        time.Now(),
	}

	if _, err = s.repo.CreateAttendanceEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save attendance event: %w", err)
	}
	log.Printf("Successfully saved attendance event for user %s %s", content.Name, content.LastName)

	if isCanteenSW(content.ReaderName, content.DoorName) {
		go forwardToSocialWalletHTTP(
			context.Background(), 
			request.Module,
			request.DataType,
			request.ModuleName,
			request.Content, 
			request.PushType,
			request.PushTypeName,
		)
	}

	return nil
}

func (s *attendanceService) GetEventsByPin(ctx context.Context, pin string, limit, offset int) ([]*domain.AttendanceEvent, error) {
	return s.repo.GetEventsByPin(ctx, pin, limit, offset)
}

func (s *attendanceService) GetEventsByDeviceSN(ctx context.Context, deviceSN string, limit, offset int) ([]*domain.AttendanceEvent, error) {
	return s.repo.GetEventsByDeviceSN(ctx, deviceSN, limit, offset)
}

func (s *attendanceService) GetAttendanceEvent(ctx context.Context, id int) (*domain.AttendanceEvent, error) {
	return s.repo.GetAttendanceEvent(ctx, id)
}

func (s *attendanceService) ListAttendanceEvents(ctx context.Context, limit, offset int) ([]*domain.AttendanceEvent, error) {
	return s.repo.ListAttendanceEvents(ctx, limit, offset)
}

func forwardToSocialWalletHTTP(ctx context.Context, module, dataType, moduleName, content, pushType, pushTypeName string) {
	url := os.Getenv("SOCIAL_WALLET_GATEWAY_URL")
	if url == "" {
		url = "https://school.kai/api/social-wallet/zkbio"
	}

	bodyMap := map[string]string{
		"module":       module,
		"dataType":     dataType,
		"moduleName":   moduleName,
		"content":      content,
		"pushType":     pushType,
		"pushTypeName": pushTypeName,
	}

	b, err := json.Marshal(bodyMap)
	if err != nil {
		log.Printf("social-wallet: marshal failed: %v", err)
		return
	}

	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx2, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		log.Printf("social-wallet: new request failed: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("social-wallet: http call failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("social-wallet: non-2xx status: %s", resp.Status)
	}
}
