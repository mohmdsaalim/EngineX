package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GRPCLoggerInterceptor is a unary interceptor that logs gRPC requests and errors.
func GRPCLoggerInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		log.Info("gRPC Request Started",
			"method", info.FullMethod,
		)

		// Process the actual request
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			log.Error("gRPC Request Failed",
				"method", info.FullMethod,
				"code", st.Code().String(),
				"error", err.Error(),
				"duration", duration.String(),
			)
		} else {
			log.Info("gRPC Request Completed",
				"method", info.FullMethod,
				"duration", duration.String(),
			)
		}

		return resp, err
	}
}

// New creates a JSON logger tagged with the service name.
func New(service string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(handler).With("service", service)
}
