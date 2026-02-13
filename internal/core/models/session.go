package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
}

type TokensPair struct {
	SessionID uuid.UUID `json:"session_id"`
	Refresh   string    `json:"refresh"`
	Access    string    `json:"access"`
}
