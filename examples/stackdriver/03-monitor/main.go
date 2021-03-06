package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"google.golang.org/api/option"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"golang.org/x/exp/rand"
)

var (
	// The task latency in milliseconds.
	latencyMs = stats.Float64("task_latency", "The task latency in milliseconds", "ms")
)

func main() {
	ctx := context.Background()
	v := &view.View{
		Name:        "task_latency_distribution",
		Measure:     latencyMs,
		Description: "The distribution of the task latencies",
		Aggregation: view.Distribution(0, 100, 200, 400, 1000, 2000, 4000),
	}
	if err := view.Register(v); err != nil {
		log.Fatalf("Failed to register the view: %v", err)
	}

	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
		MonitoringClientOptions: []option.ClientOption{
			option.WithCredentialsFile(path.Join("./.gcp/stackdriver-monitor-admin.json")),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	//view.RegisterExporter(exporter)
	//view.SetReportingPeriod(60 * time.Second)
	// Flush must be called before main() exits to ensure metrics are recorded.
	defer exporter.Flush()

	if err := exporter.StartMetricsExporter(); err != nil {
		log.Fatalf("Error starting metric exporter: %v", err)
	}
	defer exporter.StopMetricsExporter()

	// Record 100 fake latency values between 0 and 5 seconds.
	for i := 0; i < 100; i++ {
		ms := float64(5*time.Second/time.Millisecond) * rand.Float64()
		fmt.Printf("Latency %d: %f\n", i, ms)
		stats.Record(ctx, latencyMs.M(ms))
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Done recording metrics")
}
