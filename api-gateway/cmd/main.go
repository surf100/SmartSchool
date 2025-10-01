package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"api-gateway/internal/logger"

	attendancepb "api-gateway/proto/gen/attendancepb"
	canteenpb "api-gateway/proto/gen/canteenpb"
	chatpb "api-gateway/proto/gen/chatpb"
	librarypb "api-gateway/proto/gen/librarypb"
	persondumperpb "api-gateway/proto/gen/persondumperpb"
	personpb "api-gateway/proto/gen/personpb"
	terminalpb "api-gateway/proto/gen/terminalpb"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustRegister(name string, fn func() error) {
	if err := fn(); err != nil {
		slog.Error("failed to register handler", "service", name, "err", err)
		os.Exit(1)
	}
}

func main() {
	slog.SetDefault(logger.New())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Canteen Service
	canteenGRPC := getenv("CANTEEN_GRPC_ADDR", "localhost:50051")
	mustRegister("canteen.meal", func() error {
		return canteenpb.RegisterMealServiceHandlerFromEndpoint(ctx, mux, canteenGRPC, opts)
	})
	mustRegister("canteen.payment", func() error {
		return canteenpb.RegisterPaymentServiceHandlerFromEndpoint(ctx, mux, canteenGRPC, opts)
	})

	// Person Service
	personGRPC := getenv("PERSON_GRPC_ADDR", "localhost:50051")
	mustRegister("person", func() error {
		return personpb.RegisterPersonServiceHandlerFromEndpoint(ctx, mux, personGRPC, opts)
	})

	// Library Service
	libraryGRPC := getenv("LIBRARY_GRPC_ADDR", "localhost:50051")
	mustRegister("library.book", func() error {
		return librarypb.RegisterBookServiceHandlerFromEndpoint(ctx, mux, libraryGRPC, opts)
	})
	rentalGRPC := getenv("RENTAL_GRPC_ADDR", "localhost:50054")
	mustRegister("library.rental", func() error {
		return librarypb.RegisterRentalServiceHandlerFromEndpoint(ctx, mux, rentalGRPC, opts)
	})

	// Attendance Service
	attendanceGRPC := getenv("ATTENDANCE_GRPC_ADDR", "localhost:50055")
	mustRegister("attendance", func() error {
		return attendancepb.RegisterAttendanceServiceHandlerFromEndpoint(ctx, mux, attendanceGRPC, opts)
	})

	// Chat Service
	chatGRPC := getenv("CHAT_GRPC_ADDR", "localhost:50058")
	mustRegister("chat", func() error {
		return chatpb.RegisterChatServiceHandlerFromEndpoint(ctx, mux, chatGRPC, opts)
	})

	// Persons Dumper Service
	personsDumperGRPC := getenv("PERSON_DUMPER_API_URL", "localhost:50059")
	mustRegister("person-dumper", func() error {
		return persondumperpb.RegisterPersonsDumperHandlerFromEndpoint(ctx, mux, personsDumperGRPC, opts)
	})

	// Social Wallet / Terminal Service
	socialWalletGRPC := getenv("SOCIAL_WALLET_GRPC_ADDR", "localhost:50051")
	mustRegister("terminal", func() error {
		// activate-voucher
		return terminalpb.RegisterTerminalServiceHandlerFromEndpoint(ctx, mux, socialWalletGRPC, opts)
	})

	// HTTP Server
	slog.Info(
		"API Gateway started",
		"addr", ":8080",
		"chat_upstream", chatGRPC,
		"attendance_upstream", attendanceGRPC,
	)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("http server stopped", "err", err)
	}
}
