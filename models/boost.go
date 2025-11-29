package models

import (
	"time"

	"github.com/google/uuid"
)

type Boost struct {
	tableName   struct{}  `pg:"franchiso.boosts"`
	ID          uuid.UUID `pg:"id" json:"id"`
	FranchiseID uuid.UUID `pg:"franchise_id" json:"franchise_id"`
	StartDate   time.Time `pg:"start_date" json:"start_date"`
	EndDate     time.Time `pg:"end_date" json:"end_date"`
	IsActive    bool      `pg:"is_active,use_zero" json:"is_active"`
	CreatedAt   time.Time `pg:"created_at" json:"created_at"`
	UpdatedAt   time.Time `pg:"updated_at" json:"updated_at"`

	Franchise *Franchise `pg:"rel:has-one,fk:franchise_id"`
}
