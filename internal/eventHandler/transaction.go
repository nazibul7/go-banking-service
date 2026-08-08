package eventhandler

import (
	"banking-app/internal/kafka"
	"banking-app/internal/model"
	"context"
	"encoding/json"
	"log"

	kafkago "github.com/segmentio/kafka-go"
)

func TransactionHandler(ctx context.Context, msg kafkago.Message) error {
	var event struct {
		TransactionType model.TransactionType `json:"transaction_type"`
	}

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	switch event.TransactionType {
	case model.TransactionDeposit:
		var deposit kafka.DepositCompletedEvent
		if err := json.Unmarshal(msg.Value, &deposit); err != nil {
			return err
		}

		log.Printf(
			"Deposit: transaction_id=%d account=%d amount=%d",
			deposit.TransactionID,
			deposit.AccountID,
			deposit.Amount,
		)
	case model.TransactionWithdraw:
		var withdraw kafka.WithdrawCompletedEvent
		if err := json.Unmarshal(msg.Value, &withdraw); err != nil {
			return err
		}

		log.Printf(
			"Withdraw: transaction_id=%d account=%d amount=%d",
			withdraw.TransactionID,
			withdraw.AccountID,
			withdraw.Amount,
		)

	case model.TransactionTransfer:
		var transfer kafka.TransferCompletedEvent
		if err := json.Unmarshal(msg.Value, &transfer); err != nil {
			return err
		}
		log.Printf(
			"Transfer: transaction_id=%d from=%d to=%d amount=%d",
			transfer.TransactionID,
			transfer.FromAccountID,
			transfer.ToAccountID,
			transfer.Amount,
		)
	}
	return nil
}
