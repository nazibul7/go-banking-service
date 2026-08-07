package kafka

import (
	"context"
	"encoding/json"

	kafkago "github.com/segmentio/kafka-go"
)

// Producer wraps a kafka.Writer and provides methods
// to publish messages to Kafka.
type Producer struct {
	writer *kafkago.Writer
}

// NewProducer creates a Kafka producer.
//
// A kafka.Writer maintains network connections and internal buffers,
// so it should be created once during application startup and reused
// for publishing multiple messages.
func NewProducer(cfg *KafkaConfig) *Producer {
	writer := &kafkago.Writer{
		Addr: kafkago.TCP(cfg.KafkaBroker), // Kafka broker address (e.g. localhost:9092)
		// Topic:    cfg.Topic,                    // Default topic for published messages
		// Balancer: &kafkago.LeastBytes{},        // Distribute messages across partitions
	}
	return &Producer{writer: writer}
}

// Publish sends a single message to Kafka.
//
// The key is optional. If provided, Kafka uses it to determine
// which partition the message should be written to, helping preserve
// ordering for messages with the same key.
func (p *Producer) Publish(ctx context.Context, topic string, key int, event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	keyByte, err := json.Marshal(key)
	if err != nil {
		return err
	}
	msg := kafkago.Message{
		Key:   keyByte,
		Value: data,
	}
	return p.writer.WriteMessages(ctx, msg)
}

// Close releases the producer's resources.
//
// It closes underlying network connections and flushes any buffered
// messages. Call this once when the application is shutting down.
func (p *Producer) Close() error {
	return p.writer.Close()
}
