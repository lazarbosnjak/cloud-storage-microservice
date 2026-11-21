package handler

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (resp any, err error) {
	start := time.Now()
	slog.Info("rpc started", "method", info.FullMethod)

	if r := recover(); r != nil {
		slog.Error("panic in handler", "method", info.FullMethod, "panic", r, "stack", string(debug.Stack()))
		err = status.Error(codes.Internal, "internal error")
	}
	slog.Info("rpc finished",
		"method", info.FullMethod,
		"code", status.Code(err).String(),
		"duration", time.Since(start),
	)

	return next(ctx, req)
}
