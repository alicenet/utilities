package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	alicev1 "github.com/alicenet/utilities/api/alice/v1"
	"github.com/alicenet/utilities/internal/alicenet"
	"github.com/alicenet/utilities/internal/flagz"
	"github.com/alicenet/utilities/internal/handler"
	"github.com/alicenet/utilities/internal/logz"
	"github.com/alicenet/utilities/internal/service"
	"github.com/alicenet/utilities/internal/service/frontend"
)

const (
	defaultPort  = 8080
	httpTimeouts = 10 * time.Second
)

func main() {
	logz.Notice("starting up")

	port := flag.Uint64("port", defaultPort, "port to listen on")
	database := flag.String("database", "projects/mn-test-298216/instances/alicenet/databases/indexer", "spanner database")

	flagz.Parse()

	addr := fmt.Sprintf(":%d", *port)

	logz.Info("creating GRPC server")

	ctx, grpcServer := service.NewServer()

	logz.WithDetail("database", *database).Info("connecting to spanner")

	spannerClient, err := spanner.NewClient(ctx, *database)
	if err != nil {
		logz.WithDetail("err", err).Criticalf("could not conect to spanner: %v", err)
		panic(err)
	}

	defer spannerClient.Close()

	stores := alicenet.InSpanner(spannerClient)
	service := frontend.NewService(stores)
	mux := runtime.NewServeMux()

	alicev1.RegisterAliceServiceServer(grpcServer, service)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := alicev1.RegisterAliceServiceHandlerFromEndpoint(ctx, mux, addr, opts); err != nil {
		logz.WithDetail("err", err).Criticalf("could not register gateway: %v", err)
		panic(err)
	}

	httpServer := http.Server{
		Addr:              addr,
		Handler:           handler.CORS(handler.GRPC(grpcServer, mux)),
		ReadTimeout:       httpTimeouts,
		ReadHeaderTimeout: httpTimeouts,
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logz.Debug("waiting on signals")

		s := <-signals
		logz.WithDetail("signal", s).Info("got shutdown signal")

		if err := httpServer.Shutdown(context.Background()); err != nil {
			logz.WithDetail("err", err).Errorf("shutting down http server: %v", err)
		}
	}()

	logz.WithDetail("address", addr).Info("listening")

	if err := httpServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logz.WithDetail("err", err).Criticalf("listening: %v", err)
			panic(err)
		}

		logz.Notice("shutting down")
	}
}
