package main

import (
	"context"
	"log"
	"net"
	"time"

	"social-wallet/api/handler"
	"social-wallet/config"
	"social-wallet/internal/repository"

	pb "social-wallet/api/proto/gen"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	_ = godotenv.Load()

	db, err := config.ConnectPostgres()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	repo := repository.NewPostgresRepo(db)

	terminalHandler := handler.NewTerminalHandler(repo)
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			log.Println("Starting susn sync...")
			if err := repo.SyncSusnStatuses(context.Background()); err != nil {
				log.Printf("Sync error: %v\n", err)
			} else {
				log.Println("Susn sync complete.")
			}
		}
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterTerminalEventServiceServer(grpcServer, terminalHandler)
	pb.RegisterSocialWalletServiceServer(grpcServer, terminalHandler)

	log.Println("gRPC server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
