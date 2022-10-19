/*
Worker continuously scans alicenet to index blocks and transactions.

It populates a shared Spanner database that the indexer frontend serves from.
*/
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/spanner"
	"contrib.go.opencensus.io/exporter/stackdriver"

	"github.com/alicenet/utilities/internal/alicenet"
	"github.com/alicenet/utilities/internal/flagz"
	"github.com/alicenet/utilities/internal/logz"
	"github.com/alicenet/utilities/internal/service/worker"
)

func main() {
	logz.Notice("starting up")

	api := flag.String("api", "edge.staging.alice.net", "api hosting alicenet")
	database := flag.String("database", "projects/mn-test-298216/instances/alicenet/databases/indexer", "spanner database")
	metrics := flag.Bool("exportmetrics", false, "whether or not to export metrics")

	flagz.Parse()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		logz.Debug("waiting on signals")

		s := <-signals
		logz.WithDetail("signal", s).Info("got shutdown signal")

		cancel()
	}()

	if *metrics {
		logz.Info("setting up metrics exporter")

		exporter, err := stackdriver.NewExporter(stackdriver.Options{})
		if err != nil {
			panic(err)
		}

		defer exporter.Flush()

		if err := exporter.StartMetricsExporter(); err != nil {
			panic(err)
		}

		defer exporter.StopMetricsExporter()
	}

	logz.WithDetail("database", *database).Info("connecting to spanner")

	spannerClient, err := spanner.NewClient(ctx, *database)
	if err != nil {
		logz.WithDetail("err", err).Criticalf("could not conect to spanner: %v", err)
		panic(err)
	}

	defer spannerClient.Close()

	alicenetClient := alicenet.Connect(*api)
	stores := alicenet.InSpanner(spannerClient)
	worker := worker.New(alicenetClient, stores)

	worker.Run(ctx)

	logz.Notice("shutting down")
}
