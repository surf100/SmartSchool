package grpcadp

import (
	"context"
	"time"

	"person-dumper/internal/usecase"
	pb "person-dumper/proto/gen"
)

type Server struct {
	pb.UnimplementedPersonsDumperServer
	UC *usecase.SyncUC
}

func (s *Server) TriggerFullSync(ctx context.Context, r *pb.TriggerFullSyncRequest) (*pb.TriggerFullSyncResponse, error) {
	run, scheduled, err := s.UC.TriggerAll(ctx, r.GetDryRun())
	if err != nil {
		return nil, err
	}
	return &pb.TriggerFullSyncResponse{RunId: run, Scheduled: int32(scheduled)}, nil
}

func (s *Server) TriggerSchoolSync(ctx context.Context, r *pb.TriggerSchoolSyncRequest) (*pb.TriggerSchoolSyncResponse, error) {
	var id *int64
	var bin *string
	switch r.GetKey().(type) { 
	case *pb.TriggerSchoolSyncRequest_SchoolId:
		v := r.GetSchoolId()
		id = &v
	case *pb.TriggerSchoolSyncRequest_Bin:
		v := r.GetBin()
		bin = &v
	}

	run, err := s.UC.TriggerSchoolSync(ctx, id, bin, r.GetDryRun())
	if err != nil {
		return nil, err
	}
	return &pb.TriggerSchoolSyncResponse{RunId: run}, nil
}

func (s *Server) GetSyncRuns(ctx context.Context, r *pb.GetSyncRunsRequest) (*pb.GetSyncRunsResponse, error) {
	var id *int64
	if r.SchoolId != nil {
		id = r.SchoolId
	}
	rows, err := s.UC.GetRuns(ctx, id, int(r.GetLimit()))
	if err != nil {
		return nil, err
	}
	out := make([]*pb.SyncRun, 0, len(rows))
	for _, v := range rows {
		out = append(out, &pb.SyncRun{
			RunId: v.RunID, SchoolId: v.SchoolID,
			StartedAt:  v.StartedAt.Format(time.RFC3339),
			FinishedAt: v.FinishedAt.Format(time.RFC3339),
			Total:      int32(v.Total), Inserted: int32(v.Inserted),
			Updated: int32(v.Updated), Failed: int32(v.Failed), Ok: v.OK,
		})
	}
	return &pb.GetSyncRunsResponse{Runs: out}, nil
}
