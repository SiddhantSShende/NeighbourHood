package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Integration struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Provider     string    `json:"provider" db:"provider"`
	AccessToken  string    `json:"-" db:"access_token"`
	RefreshToken string    `json:"-" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	Metadata     []byte    `json:"metadata" db:"metadata"` // JSONB
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type APIKey struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	UserID    uuid.UUID      `json:"user_id" db:"user_id"`
	KeyHash   string         `json:"-" db:"key_hash"`
	Scopes    pq.StringArray `json:"scopes" db:"scopes"`
	Name      string         `json:"name" db:"name"`
	ExpiresAt time.Time      `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}
