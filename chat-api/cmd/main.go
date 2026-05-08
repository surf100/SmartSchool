package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"chat-api/config"
	chatgrpc "chat-api/internal/delivery/grpc"
	"chat-api/internal/logger"
	"chat-api/internal/middleware"
	"chat-api/internal/repository"
	chatpb "chat-api/proto/gen"

	"google.golang.org/grpc"
)

func main() {
	slog.SetDefault(logger.New())

	cfg := config.LoadConfig()

	client, db, err := config.ConnectMongo(cfg)
	if err != nil {
		slog.Error("mongo connect failed", "err", err)
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())

	msgRepo := repository.NewMessageRepository(db)

	addr := ":50058"
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

	chatpb.RegisterChatServiceServer(grpcServer, chatgrpc.NewChatServer(msgRepo))

	slog.Info("chat-api: serving", "addr", addr)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("serve failed", "err", err)
		os.Exit(1)
	}
}
