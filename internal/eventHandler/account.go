package eventhandler

import (
	"banking-app/internal/kafka"
	"context"
	"encoding/json"
	"log"

	kafkago "github.com/segmentio/kafka-go"
)

func AccountHandler(ctx context.Context, msg kafkago.Message) error {
	var event kafka.AccountCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	log.Printf(
		"Account created: account_id=%d user_id=%d balance=%d",
		event.AccountID,
		event.UserID,
		event.Balance,
	)
	return nil
}
