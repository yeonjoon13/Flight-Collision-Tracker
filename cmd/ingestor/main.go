package main

import (
    "context"
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/yeonjoon13/Flight-Collision-Tracker/internal/kafka"
    "github.com/yeonjoon13/Flight-Collision-Tracker/internal/opensky"
)

func main() {
    var (
        broker = flag.String("broker", os.Getenv("KAFKA_BROKER"), "Kafka broker address")
        topic  = flag.String("topic", "flight_updates", "Kafka topic for flight data")
        url    = flag.String("url", "", "OpenSky API URL (optional)")
        interval = flag.Duration("interval", 2*time.Minute, "Poll interval")
    )
    flag.Parse()

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    ticker := time.NewTicker(*interval)
    defer ticker.Stop()

    log.Println("Starting ingestor...")
    for {
        select {
        case <-ctx.Done():
            log.Println("Shutting down ingestor")
            return
        case <-ticker.C:
            updates, err := opensky.FetchStates(ctx, *url)
			log.Printf("Fetched %d flights from OpenSky", len(updates))
            if err != nil {
                log.Printf("fetch error: %v", err)
                continue
            }
            if err := kafka.PublishFlights(ctx, *broker, *topic, updates); err != nil {
                log.Printf("publish error: %v", err)
            }
			log.Printf("Published %d flights to Kafka", len(updates))
        }
    }
}