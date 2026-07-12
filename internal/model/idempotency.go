package model

import (
	"encoding/json"
	"time"
)

type Idempotency struct {
	ID             int
	UserID         int
	IdempotencyKey string
	StatusCode     int
	Response       json.RawMessage
	CreatedAt      time.Time
	ExpiresAt      time.Time
}
