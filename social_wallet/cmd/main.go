package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"social-wallet/api/handler"
	"social-wallet/config"
	"social-wallet/internal/repository"

	pb "social-wallet/api/proto/gen"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	logger "social-wallet/internal/logger"
	"social-wallet/internal/middleware"
)

func main() {
	_ = godotenv.Load()
	slog.SetDefault(logger.New())

	db, err := config.ConnectPostgres()
	if err != nil {
		slog.Error("DB connection failed", "err", err)
		os.Exit(1)
	}
	repo := repository.NewPostgresRepo(db)

	terminalHandler := handler.NewTerminalHandler(repo)

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			slog.Info("susn sync: start")
			if err := repo.SyncSusnStatuses(context.Background()); err != nil {
				slog.Error("susn sync: error", "err", err)
			} else {
				slog.Info("susn sync: complete")
			}
		}
	}()

	portListen := os.Getenv("INTERNAL_PORT")
	if portListen == "" {
		portListen = "50051"
	}
	addr := ":" + portListen

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("listen failed", "err", err, "addr", addr)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryLogging(),
		),
	)

	pb.RegisterTerminalServiceServer(grpcServer, terminalHandler)

	slog.Info("social-wallet: serving gRPC", "addr", addr)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("serve failed", "err", err)
		os.Exit(1)
	}
}
