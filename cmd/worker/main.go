package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/spanner"
	_ "github.com/golang-migrate/migrate/v4/database/spanner"

	"github.com/alicenet/indexer/internal/alicenet"
	"github.com/alicenet/indexer/internal/flagz"
	"github.com/alicenet/indexer/internal/service/worker"
)

func main() {
	api := flag.String("api", "edge.staging.alice.net", "api hosting alicenet")
	database := flag.String("database", "projects/mn-test-298216/instances/alicenet/databases/indexer", "spanner database")

	flagz.Parse()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-signals
		cancel()
	}()

	fmt.Println("connecting to spanner")

	spannerClient, err := spanner.NewClient(ctx, *database)
	if err != nil {
		panic(err)
	}

	defer spannerClient.Close()

	alicenetClient := alicenet.Connect(*api)
	stores := alicenet.InSpanner(spannerClient)
	worker := worker.New(alicenetClient, stores)

	worker.Run(ctx)
}
