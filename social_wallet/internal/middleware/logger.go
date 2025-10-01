package middleware

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func UnaryLogging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		p, _ := peer.FromContext(ctx)

		resp, err := handler(ctx, req)
		st := status.Convert(err)

		slog.Info("grpc_request",
			"method", info.FullMethod,
			"code", st.Code().String(),
			"latency_ms", time.Since(start).Milliseconds(),
			"peer", func() string {
				if p != nil {
					return p.Addr.String()
				}
				return ""
			}(),
		)
		return resp, err
	}
}
