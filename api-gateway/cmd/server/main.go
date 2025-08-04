package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "api-gateway/proto/gen"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	log.Println("API Gateway подключается к :50051")
	if err := pb.RegisterTerminalEventServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts); err != nil {
		log.Fatalf("Не удалось зарегистрировать TerminalEventService: %v", err)
	}
	if err := pb.RegisterSocialWalletServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts); err != nil {
		log.Fatalf("Не удалось зарегистрировать SocialWalletService: %v", err)
	}

	log.Println("API Gateway слушает на :8090")
	if err := http.ListenAndServe(":8090", mux); err != nil {
		log.Fatalf("Ошибка запуска gateway: %v", err)
	}
}
