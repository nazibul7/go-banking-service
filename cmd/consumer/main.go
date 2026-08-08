package main

import (
	eventhandler "banking-app/internal/eventHandler"
	"banking-app/internal/kafka"
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := kafka.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	transactionConsumer := kafka.NewConsumer(cfg, kafka.TransactionTopic, eventhandler.TransactionHandler)
	accountConsumer := kafka.NewConsumer(cfg, kafka.AccountTopic, eventhandler.AccountHandler)

	log.Printf(
		"transaction consumer started: topic=%s group=%s",
		kafka.TransactionTopic,
		cfg.KafkaConsumerGroup,
	)

	log.Printf(
		"account consumer started: topic=%s group=%s",
		kafka.AccountTopic,
		cfg.KafkaConsumerGroup,
	)

	errCh := make(chan error, 2)

	go func() {
		errCh <- transactionConsumer.Consume(ctx)
	}()

	go func() {
		errCh <- accountConsumer.Consume(ctx)
	}()

	defer transactionConsumer.Close()
	defer accountConsumer.Close()

	err = <-errCh
	err = <-errCh

	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("consumer stopped: %v", err)
	}

	cancel()

	log.Println("consumer shutting down")
}
