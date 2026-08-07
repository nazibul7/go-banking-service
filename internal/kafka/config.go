package kafka

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type KafkaConfig struct {
	KafkaBroker   string
	KafkaConsumerGroup string
}

func LoadConfig() (*KafkaConfig, error) {
	if err := godotenv.Load(); err != nil {
		return nil, errors.New("could not load .env file")
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		return nil, errors.New("kafka broker is missing")
	}

	consumeGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	if consumeGroup == "" {
		return nil, errors.New("kafka consumer group is missing")
	}
	return &KafkaConfig{
		KafkaBroker:   broker,
		KafkaConsumerGroup: consumeGroup,
	}, nil
}
