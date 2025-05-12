package kafka

import (
    "github.com/segmentio/kafka-go"
)

// NewReader returns a configured kafka.Reader
func NewReader(broker, topic, groupID string) *kafka.Reader {
    return kafka.NewReader(kafka.ReaderConfig{
	Brokers:        []string{broker},
	Topic:          topic,
	MinBytes:       10e3,
	MaxBytes:       10e6,
	GroupID: "collision-detector-test-1",
	ReadLagInterval: -1,  // 👈 disables auto-closing when no messages
	StartOffset:     kafka.FirstOffset, // 👈 starts from beginning

	})
}