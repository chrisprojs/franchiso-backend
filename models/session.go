package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	tableName    struct{}  `pg:"franchiso.sessions"`
	ID           uuid.UUID `pg:"id" json:"id"`
	UserID       uuid.UUID `pg:"user_id" json:"user_id"`
	RefreshToken string    `pg:"refresh_token" json:"refresh_token"`
	ExpiresAt    time.Time `pg:"expires_at" json:"expires_at"`
	CreatedAt    time.Time `pg:"created_at" json:"created_at"`
	UpdatedAt    time.Time `pg:"updated_at" json:"updated_at"`
}
