package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"trackpocket/internal/model/domain"
)

type TransactionFilter struct {
	Search     string
	Type       string
	CategoryID *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	Limit      int
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.Transaction, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID, filter TransactionFilter) ([]domain.Transaction, int, error)
	Update(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type TransactionRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return &TransactionRepositoryImpl{DB: db}
}

func (repository *TransactionRepositoryImpl) Create(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error) {
	transaction.ID = uuid.New()

	query := `
	INSERT INTO transactions (id, user_id, category_id, type, amount, note, transaction_date)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING created_at, updated_at`

	err := repository.DB.QueryRow(ctx, query,
		transaction.ID, transaction.UserID, transaction.CategoryID, transaction.Type,
		transaction.Amount, transaction.Note, transaction.TransactionDate,
	).Scan(&transaction.CreatedAt, &transaction.UpdatedAt)

	if err != nil {
		return transaction, err
	}
	return transaction, nil
}

func (repository *TransactionRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
	query := `
	SELECT id, user_id, category_id, type, amount, note, transaction_date, created_at, updated_at
	FROM transactions WHERE id = $1`

	var t domain.Transaction
	err := repository.DB.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.CategoryID, &t.Type, &t.Amount, &t.Note, &t.TransactionDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return t, err
	}
	return t, nil
}

func (repository *TransactionRepositoryImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID, filter TransactionFilter) ([]domain.Transaction, int, error) {
	conditions := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("note ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}
	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Hitung total count dulu (buat pagination metadata)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM transactions WHERE %s", whereClause)
	var totalCount int
	if err := repository.DB.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, category_id, type, amount, note, transaction_date, created_at, updated_at
		FROM transactions
		WHERE %s
		ORDER BY transaction_date DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	rows, err := repository.DB.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.CategoryID, &t.Type, &t.Amount, &t.Note, &t.TransactionDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, t)
	}

	return transactions, totalCount, rows.Err()
}

func (repository *TransactionRepositoryImpl) Update(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error) {
	query := `
	UPDATE transactions
	SET category_id = $1, type = $2, amount = $3, note = $4, transaction_date = $5, updated_at = CURRENT_TIMESTAMP
	WHERE id = $6
	RETURNING id, user_id, category_id, type, amount, note, transaction_date, created_at, updated_at`

	var t domain.Transaction
	err := repository.DB.QueryRow(ctx, query,
		transaction.CategoryID, transaction.Type, transaction.Amount, transaction.Note, transaction.TransactionDate, transaction.ID,
	).Scan(&t.ID, &t.UserID, &t.CategoryID, &t.Type, &t.Amount, &t.Note, &t.TransactionDate, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return t, err
	}
	return t, nil
}

func (repository *TransactionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM transactions WHERE id = $1`
	_, err := repository.DB.Exec(ctx, query, id)
	return err
}