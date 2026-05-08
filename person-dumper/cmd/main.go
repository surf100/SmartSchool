package main

import (
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	zk "person-dumper/internal/device/zkteco"
	grpcadp "person-dumper/internal/grpc"
	personclient "person-dumper/internal/personclient"
	"person-dumper/internal/postgres"
	"person-dumper/internal/usecase"
	pb "person-dumper/proto/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func env(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))

	dsn := env("PG_DSN", "postgres://dev:DevUserKai1!@10.2.13.16:5432/school-dev?sslmode=disable")
	addr := env("GRPC_ADDR", ":50059")
	personAddr := env("PERSON_API_ADDR", "localhost:50053")

	// gRPC соединение с person-api
	personConn, err := grpc.Dial(personAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer personConn.Close()

	// Postgres repo
	repo, err := postgres.New(dsn)
	if err != nil {
		panic(err)
	}
	to := 5 * time.Second
	if v := os.Getenv("ZK_TIMEOUT_MS"); v != "" {
		if ms, _ := strconv.Atoi(v); ms > 0 {
			to = time.Duration(ms) * time.Millisecond
		}
	}

	// Usecase с реальным person-api клиентом
	uc := &usecase.SyncUC{
		Schools: repo, Lookup: repo, Runs: repo, Reads: repo,
		Dev:     zk.NewWithTimeout(to),
		Persons: personclient.NewGRPC(personConn, 3*time.Second),
	}

	// gRPC сервер
	s := grpc.NewServer()
	pb.RegisterPersonsDumperServer(s, &grpcadp.Server{UC: uc})
	if os.Getenv("GRPC_REFLECTION") != "0" {
		reflection.Register(s)
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	slog.Info("persons-dumper started", "addr", addr)
	if err := s.Serve(l); err != nil {
		panic(err)
	}
}
