package model

import "encoding/json"

type Idempotency struct {
	UserID         int
	IdempotencyKey string
	Response       json.RawMessage
}
