package kafka

import (
    "context"
    "encoding/json"
    "github.com/segmentio/kafka-go"
    "github.com/yeonjoon13/Flight-Collision-Tracker/internal/model"
)

func PublishFlights(ctx context.Context, broker, topic string, msgs []model.FlightUpdate) error {
    w := kafka.NewWriter(kafka.WriterConfig{Brokers: []string{broker}, Topic: topic})
    defer w.Close()

    records := make([]kafka.Message, len(msgs))
    for i, f := range msgs {
        b, _ := json.Marshal(f)
        records[i] = kafka.Message{Key: []byte(f.ICAO24), Value: b}
    }
    return w.WriteMessages(ctx, records...)
}