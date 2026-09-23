package web

import (
	"time"

	"github.com/google/uuid"
)

type TransactionResponse struct {
	ID              uuid.UUID `json:"id"`
	CategoryID      uuid.UUID `json:"category_id"`
	Type            string    `json:"type"`
	Amount          string    `json:"amount"`
	Note            string    `json:"note"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PaginatedTransactionResponse struct {
	Data       []TransactionResponse `json:"data"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalCount int                   `json:"total_count"`
}