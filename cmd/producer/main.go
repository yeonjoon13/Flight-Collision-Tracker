package main

import (
    "context"
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/yeonjoon13/Flight-Collision-Tracker/internal/collision"
    "github.com/yeonjoon13/Flight-Collision-Tracker/internal/kafka"
)

func main() {
    broker := flag.String("broker", os.Getenv("KAFKA_BROKER"), "Kafka broker address")
    topic  := flag.String("topic", "flight_updates", "Kafka topic for flight data")
    group  := flag.String("group", "collision-group", "Consumer group ID")
    flag.Parse()

    // ensure alert topic exists
    kafka.CreateTopics(*broker, []kafka.TopicConfig{{Topic: "collision_alerts", NumPartitions: 1, ReplicationFactor: 1}})

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    reader := kafka.NewReader(*broker, *topic, *group)
    log.Println("Starting collision detector...")
    collision.RunDetector(ctx, reader)
}