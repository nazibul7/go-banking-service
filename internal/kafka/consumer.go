package kafka

import (
	"context"
	"errors"
	"log"

	kafkago "github.com/segmentio/kafka-go"
)

type MessageHandler func(ctx context.Context, msg kafkago.Message) error

type Consumer struct {
	reader  *kafkago.Reader
	handler MessageHandler
}

func NewConsumer(cfg *KafkaConfig, topic string, handler MessageHandler) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   topic,
		GroupID: cfg.KafkaConsumerGroup,
	})
	return &Consumer{
		reader:  reader,
		handler: handler,
	}
}

func (c *Consumer) Read(ctx context.Context) (kafkago.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c *Consumer) Consume(ctx context.Context) error {
	for {
		// use fetchmessage for manual commits(tell kafka we have processed message or received this message)
		// readmessage is autocommit means that particular message u wont be getting again
		// ReadMessage() → automatic commit
		// FetchMessage() + CommitMessages() → manual commit
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Printf("fetch error: %v", err)
			continue
		}
		if err := c.handler(ctx, msg); err != nil {
			log.Printf("handler error for message at offset %d: %v", msg.Offset, err)
			// decide: skip and commit anyway, or retry, or dead-letter — see below
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit error: %v", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
