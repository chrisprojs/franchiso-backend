package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	tableName    struct{}  `pg:"franchiso.users"`
	ID           uuid.UUID `pg:"id" json:"id"`
	Name         string    `pg:"name" json:"name"`
	Email        string    `pg:"email" json:"email"`
	PasswordHash string    `pg:"password_hash" json:"-"`
	Role         string    `pg:"role" json:"role"`
	CreatedAt    time.Time `pg:"created_at" json:"created_at"`
	UpdatedAt    time.Time `pg:"updated_at" json:"updated_at"`
} 