package model

import (
	"encoding/json"
	"time"
)

type Idempotency struct {
	ID             int             `json:"id"`
	UserID         int             `json:"user_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	Response       json.RawMessage `json:"response_body"`
	StatusCode     int             `json:"status_code"`
	CreatedAt      time.Time       `json:"created_at"`
}
