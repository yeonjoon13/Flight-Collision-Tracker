package kafka

import (
    "github.com/segmentio/kafka-go"
	"time"
)

// NewReader returns a configured kafka.Reader
func NewReader(broker, topic, groupID string) *kafka.Reader {
    return kafka.NewReader(kafka.ReaderConfig{
        Brokers:         []string{broker},
        Topic:           topic,
        MinBytes:        10e3,
        MaxBytes:        10e6,
        GroupID:         groupID,
        ReadLagInterval: -1,             // Disables auto-closing when no messages
        MaxWait:         1 * time.Second, // Maximum wait time for polling
        StartOffset:     kafka.FirstOffset, // Start from beginning when group is new
    })
}