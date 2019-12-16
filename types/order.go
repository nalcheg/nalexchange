package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	ID     uuid.UUID       `db:"id" json:"id"`
	Time   time.Time       `db:"time" json:"time"`
	UserID uuid.UUID       `db:"user_id" json:"user_id"`
	Side   Side            `db:"side" json:"side"`
	Price  float64         `db:"price" json:"price"`
	Amount decimal.Decimal `db:"amount" json:"amount"`
}
