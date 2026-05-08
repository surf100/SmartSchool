package grpc

import (
	"attendance-api/internal/domain"
	"attendance-api/internal/domain/application"
	"context"
	"fmt"
	"log"

	pb "attendance-api/proto"
)

type AttendanceServer struct {
	pb.UnimplementedAttendanceServiceServer
	service application.AttendanceService
}

func NewAttendanceServer(service application.AttendanceService) *AttendanceServer {
	return &AttendanceServer{
		service: service,
	}
}

func (s *AttendanceServer) PushEvent(ctx context.Context, req *pb.PushEventRequest) (*pb.PushEventResponse, error) {
	log.Printf("Received ZKBio event: module=%s, pushType=%s", req.Module, req.PushType)

	zkbioRequest := &domain.ZKBioPushRequest{
		Module:       req.Module,
		DataType:     req.DataType,
		ModuleName:   req.ModuleName,
		Content:      req.Content,
		PushType:     req.PushType,
		PushTypeName: req.PushTypeName,
	}

	err := s.service.ProcessZKBioEvent(ctx, zkbioRequest)
	if err != nil {
		log.Printf("Error processing ZKBio event: %v", err)
		return &pb.PushEventResponse{
			Status: fmt.Sprintf("error: %v", err),
		}, nil
	}

	log.Printf("Successfully processed ZKBio event")
	return &pb.PushEventResponse{
		Status: "success",
	}, nil
}

func (s *AttendanceServer) GetAttendanceLogs(ctx context.Context, req *pb.GetAttendanceLogsRequest) (*pb.GetAttendanceLogsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	log.Printf("Getting attendance logs: limit=%d, offset=%d, pin=%s, device_sn=%s",
		limit, offset, req.Pin, req.DeviceSn)

	var events []*domain.AttendanceEvent
	var err error

	if req.Pin != "" {
		events, err = s.service.GetEventsByPin(ctx, req.Pin, limit, offset)
	} else if req.DeviceSn != "" {
		events, err = s.service.GetEventsByDeviceSN(ctx, req.DeviceSn, limit, offset)
	} else {
		events, err = s.service.ListAttendanceEvents(ctx, limit, offset)
	}

	if err != nil {
		log.Printf("Error getting attendance logs: %v", err)
		return &pb.GetAttendanceLogsResponse{
			Logs:  []*pb.AttendanceLog{},
			Total: 0,
		}, nil
	}

	pbLogs := make([]*pb.AttendanceLog, len(events))
	for i, event := range events {
		pbLogs[i] = &pb.AttendanceLog{
			Id:               int64(event.ID),
			Pin:              event.Pin,
			Name:             event.Name,
			LastName:         event.LastName,
			DeptCode:         event.DeptCode,
			DeptName:         event.DeptName,
			ReaderName:       event.ReaderName,
			DoorName:         event.DoorName,
			DeviceSn:         event.DeviceSN,
			CapturePhotoPath: event.CapturePhotoPath,
			VerifyModeName:   event.VerifyModeName,
			EventName:        event.EventName,
			EventTime:        event.EventTime.Format("2006-01-02 15:04:05"),
			RawPayload:       event.RawPayload,
			UniqueKey:        event.UniqueKey,
			LogId:            int32(event.LogID),
			CreatedAt:        event.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &pb.GetAttendanceLogsResponse{
		Logs:  pbLogs,
		Total: int32(len(pbLogs)),
	}, nil
}
