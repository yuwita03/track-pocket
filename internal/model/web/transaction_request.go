package web

import "time"

type CreateTransactionRequest struct {
	CategoryID      string    `json:"category_id" validate:"required,uuid"`
	Type            string    `json:"type" validate:"required,oneof=INCOME EXPENSE"`
	Amount          string    `json:"amount" validate:"required"`
	Note            string    `json:"note" validate:"max=255"`
	TransactionDate time.Time `json:"transaction_date" validate:"required"`
}

type UpdateTransactionRequest struct {
	CategoryID      string    `json:"category_id" validate:"required,uuid"`
	Type            string    `json:"type" validate:"required,oneof=INCOME EXPENSE"`
	Amount          string    `json:"amount" validate:"required"`
	Note            string    `json:"note" validate:"max=255"`
	TransactionDate time.Time `json:"transaction_date" validate:"required"`
}

type TransactionQueryParams struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	Search     string `form:"search"`
	Type       string `form:"type"`
	CategoryID string `form:"category_id"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
}