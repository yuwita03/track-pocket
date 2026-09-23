package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type DashboardSummary struct {
	TotalIncome      decimal.Decimal
	TotalExpense     decimal.Decimal
	TransactionCount int
}

type CategoryExpense struct {
	CategoryID   uuid.UUID
	CategoryName string
	Amount       decimal.Decimal
}

type DashboardRepository interface {
	GetSummary(ctx context.Context, userID uuid.UUID, year int, month int) (DashboardSummary, error)
	GetCategoryExpenses(ctx context.Context, userID uuid.UUID, year int, month int) ([]CategoryExpense, error)
}

type DashboardRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewDashboardRepository(db *pgxpool.Pool) DashboardRepository {
	return &DashboardRepositoryImpl{DB: db}
}

func (repository *DashboardRepositoryImpl) GetSummary(ctx context.Context, userID uuid.UUID, year int, month int) (DashboardSummary, error) {
	query := `
	SELECT
		COALESCE(SUM(amount) FILTER (WHERE type = 'INCOME'), 0) AS total_income,
		COALESCE(SUM(amount) FILTER (WHERE type = 'EXPENSE'), 0) AS total_expense,
		COUNT(*) AS transaction_count
	FROM transactions
	WHERE user_id = $1
		AND EXTRACT(YEAR FROM transaction_date) = $2
		AND EXTRACT(MONTH FROM transaction_date) = $3`

	var summary DashboardSummary
	err := repository.DB.QueryRow(ctx, query, userID, year, month).Scan(
		&summary.TotalIncome, &summary.TotalExpense, &summary.TransactionCount,
	)
	if err != nil {
		return summary, err
	}
	return summary, nil
}

func (repository *DashboardRepositoryImpl) GetCategoryExpenses(ctx context.Context, userID uuid.UUID, year int, month int) ([]CategoryExpense, error) {
	query := `
	SELECT c.id, c.name, SUM(t.amount) AS total_amount
	FROM transactions t
	JOIN categories c ON t.category_id = c.id
	WHERE t.user_id = $1
		AND t.type = 'EXPENSE'
		AND EXTRACT(YEAR FROM t.transaction_date) = $2
		AND EXTRACT(MONTH FROM t.transaction_date) = $3
	GROUP BY c.id, c.name
	ORDER BY total_amount DESC`

	rows, err := repository.DB.Query(ctx, query, userID, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []CategoryExpense
	for rows.Next() {
		var e CategoryExpense
		if err := rows.Scan(&e.CategoryID, &e.CategoryName, &e.Amount); err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}

	return expenses, rows.Err()
}