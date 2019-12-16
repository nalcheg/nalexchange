package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Deal struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	Time         time.Time       `db:"time" json:"time"`
	LeftOrderId  uuid.UUID       `db:"left_order_id" json:"left_order_id"`
	RigthOrderId uuid.UUID       `db:"rigth_order_id" json:"rigth_order_id"`
	Price        float64         `db:"price" json:"price"`
	Amount       decimal.Decimal `db:"amount" json:"amount"`
}
