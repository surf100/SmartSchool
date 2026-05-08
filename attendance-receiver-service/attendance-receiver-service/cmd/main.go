package main

import (
	"attendance-api/internal/adapters/grpc"
	"attendance-api/internal/adapters/grpc/middleware"
	"attendance-api/internal/infrastructure/postgres"
	logger "attendance-api/internal/log"
	"attendance-api/internal/usecase"
	pb "attendance-api/proto"
	"database/sql"
	"log/slog"
	"net"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	grpcserver "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1) .env до логгера, чтобы LOG_LEVEL/LOG_FORMAT подхватились
	_ = godotenv.Load()

	// 2) глобальный slog
	slog.SetDefault(logger.New())

	// 3) DSN
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		slog.Error("POSTGRES_DSN env var is required")
		os.Exit(1)
	}

	// 4) DB connect
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping database", "err", err)
		os.Exit(1)
	}

	repo := postgres.NewAttendanceRepository(db)
	service := usecase.NewAttendanceService(repo)

	// 5) gRPC server + наш лог-интерцептор
	s := grpcserver.NewServer(
		grpcserver.ChainUnaryInterceptor(
			middleware.UnaryLogging(),
		),
	)

	attendanceServer := grpc.NewAttendanceServer(service)
	pb.RegisterAttendanceServiceServer(s, attendanceServer)
	reflection.Register(s)

	// 6) порт
	internalPort := os.Getenv("INTERNAL_PORT")
	if internalPort == "" {
		internalPort = ":50055"
	} else {
		internalPort = ":" + internalPort
	}

	slog.Info("attendance-receiver: starting", "listen", internalPort)

	lis, err := net.Listen("tcp", internalPort)
	if err != nil {
		slog.Error("Failed to listen", "err", err, "addr", internalPort)
		os.Exit(1)
	}

	if err := s.Serve(lis); err != nil {
		slog.Error("Failed to serve gRPC", "err", err)
		os.Exit(1)
	}
}
