package kafka

import (
    "fmt"
    "net"

    "github.com/segmentio/kafka-go"
)

type TopicConfig struct {
    Topic             string
    NumPartitions     int
    ReplicationFactor int
}

// CreateTopics ensures each topic exists with given config
func CreateTopics(broker string, configs []TopicConfig) error {
    conn, err := kafka.Dial("tcp", broker)
    if err != nil {
        return err
    }
    defer conn.Close()

    controller, err := conn.Controller()
    if err != nil {
        return err
    }
    hostPort := net.JoinHostPort(controller.Host, fmt.Sprint(controller.Port))
    ctrlConn, err := kafka.Dial("tcp", hostPort)
    if err != nil {
        return err
    }
    defer ctrlConn.Close()

    for _, cfg := range configs {
        err = ctrlConn.CreateTopics(kafka.TopicConfig{
            Topic:             cfg.Topic,
            NumPartitions:     cfg.NumPartitions,
            ReplicationFactor: cfg.ReplicationFactor,
        })
        if err != nil {
            return err
        }
    }
    return nil
}