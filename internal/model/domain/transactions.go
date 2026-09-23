package domain


import (
	"time"
	"github.com/shopspring/decimal"
	"github.com/google/uuid"
)

type Transaction struct {
	ID					uuid.UUID
	UserID				uuid.UUID
	CategoryID			uuid.UUID
	Type				string
	Amount				decimal.Decimal
	Note				string
	TransactionDate		time.Time
	CreatedAt 			time.Time
	UpdatedAt 			time.Time

}