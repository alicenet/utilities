// Package service handles running GRPC services gracefully.
package service

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewServer set up with GRPC reflection and graceful shutdown on receiving an INT or TERM os signal.
func NewServer() (context.Context, *grpc.Server) {
	server := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	reflection.Register(server)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-signals
		cancel()
		server.GracefulStop()
	}()

	return ctx, server
}
