package models

import (
	"github.com/google/uuid"
	"time"
)

type Payment struct {
	tableName         struct{}  `pg:"franchiso.payments"`
	ID                uuid.UUID `pg:"id" json:"id"`
	BoostID           uuid.UUID `pg:"boost_id" json:"boost_id"`
	TransactionID     uuid.UUID `pg:"transaction_id" json:"transaction_id"`
	GrossAmount       float64   `pg:"gross_amount" json:"gross_amount"`
	PaymentType       string    `pg:"payment_type" json:"payment_type"`
	TransactionTime   time.Time `pg:"transaction_time" json:"transaction_time"`
	TransactionStatus string    `pg:"transaction_status" json:"transaction_status"`
	CreatedAt         time.Time `pg:"created_at" json:"created_at"`
	UpdatedAt         time.Time `pg:"updated_at" json:"updated_at"`

	Boost *Boost `pg:"rel:has-one,fk:boost_id" json:"boost"`
}
