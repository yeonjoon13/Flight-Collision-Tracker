package kafka

import (
    "context"
    "encoding/json"
    "github.com/segmentio/kafka-go"
    "github.com/yeonjoon13/Flight-Collision-Tracker/internal/model"
    "log"
)

// PublishFlights publishes flight updates to a Kafka topic
func PublishFlights(ctx context.Context, broker, topic string, msgs []model.FlightUpdate) error {
    w := kafka.NewWriter(kafka.WriterConfig{
        Brokers: []string{broker}, 
        Topic: topic,
    })
    defer w.Close()

    records := make([]kafka.Message, len(msgs))
    for i, f := range msgs {
        b, err := json.Marshal(f)
        if err != nil {
            log.Printf("Error marshaling flight %s: %v", f.ICAO24, err)
            continue
        }
        records[i] = kafka.Message{Key: []byte(f.ICAO24), Value: b}
    }
    
    return w.WriteMessages(ctx, records...)
}
