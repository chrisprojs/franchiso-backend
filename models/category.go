package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	tableName struct{}  `pg:"franchiso.categories"`
	ID        uuid.UUID `pg:"id" json:"id"`
	Category  string    `pg:"category" json:"category"`
	CreatedAt time.Time `pg:"created_at" json:"created_at"`
	UpdatedAt time.Time `pg:"updated_at" json:"updated_at"`
}
