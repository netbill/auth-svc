package models

import (
	"time"

	"github.com/google/uuid"
)

type AccountActor struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Role      string    `json:"role"`
}

type Account struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	Version  int32     `json:"version"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type AccountEmail struct {
	AccountID uuid.UUID  `json:"account_id"`
	Email     string     `json:"email"`
	Verified  bool       `json:"verified"`
	Version   int32      `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type AccountPassword struct {
	AccountID uuid.UUID  `json:"account_id"`
	Hash      string     `json:"hash"`
	Version   int32      `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
