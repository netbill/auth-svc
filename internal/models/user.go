package models

import (
	"time"

	"github.com/google/uuid"
)

type UserActor struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Role      string    `json:"role"`
}

type User struct {
	ID          uuid.UUID `json:"id"`
	Role        string    `json:"role"`
	Username    string    `json:"username"`
	Pseudonym   *string   `json:"pseudonym,omitempty"`
	Description *string   `json:"description,omitempty"`
	AvatarKey   *string   `json:"avatar_key,omitempty"`
	Version     int32     `json:"version"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UploadUserMediaLinks struct {
	Avatar UploadMediaLink `json:"avatar"`
}

type UploadMediaLink struct {
	Key        string `json:"key"`
	UploadURL  string `json:"upload_url"`
	PreloadUrl string `json:"preload_url"`
}

type UserEmail struct {
	UserID    uuid.UUID  `json:"user_id"`
	Email     string     `json:"email"`
	Verified  bool       `json:"verified"`
	Version   int32      `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UserPassword struct {
	UserID    uuid.UUID  `json:"user_id"`
	Hash      string     `json:"hash"`
	Version   int32      `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
