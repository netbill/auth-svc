package models

import "github.com/google/uuid"

type TokensPair struct {
	SessionID uuid.UUID `json:"session_id"`
	Refresh   string    `json:"refresh"`
	Access    string    `json:"access"`
}
