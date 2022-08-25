package main

import (
	"context"
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

	alicev1 "github.com/alicenet/indexer/api/alice/v1"
	"github.com/alicenet/indexer/internal/alicenet"
	"github.com/alicenet/indexer/internal/flagz"
	"github.com/alicenet/indexer/internal/handler"
	"github.com/alicenet/indexer/internal/service"
	"github.com/alicenet/indexer/internal/service/frontend"
)

const (
	defaultPort  = 8080
	httpTimeouts = 10 * time.Second
)

func main() {
	port := flag.Uint64("port", defaultPort, "port to listen on")
	database := flag.String("database", "projects/mn-test-298216/instances/alicenet/databases/indexer", "spanner database")

	flagz.Parse()

	addr := fmt.Sprintf(":%d", *port)

	fmt.Println("running frontend")

	ctx, grpcServer := service.NewServer()

	fmt.Println("connecting to spanner")

	spannerClient, err := spanner.NewClient(ctx, *database)
	if err != nil {
		panic(err)
	}

	defer spannerClient.Close()

	stores := alicenet.InSpanner(spannerClient)
	service := frontend.NewService(stores)
	mux := runtime.NewServeMux()

	alicev1.RegisterAliceServiceServer(grpcServer, service)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := alicev1.RegisterAliceServiceHandlerFromEndpoint(ctx, mux, addr, opts); err != nil {
		panic(err)
	}

	httpServer := http.Server{
		Addr:              addr,
		Handler:           handler.GRPC(grpcServer, mux),
		ReadTimeout:       httpTimeouts,
		ReadHeaderTimeout: httpTimeouts,
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals

		_ = httpServer.Shutdown(context.Background())
	}()

	if err := httpServer.ListenAndServe(); err != nil {
		panic(err)
	}
}
