package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"trackpocket/internal/model/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, category domain.Category) (domain.Category, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.Category, error)
	Update(ctx context.Context, id uuid.UUID, name string) (domain.Category, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HasTransactions(ctx context.Context, categoryID uuid.UUID) (bool, error)
}

type CategoryRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &CategoryRepositoryImpl{DB: db}
}

func (repository *CategoryRepositoryImpl) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	category.ID = uuid.New()

	query := `
	INSERT INTO categories (id, user_id, name, type, is_default, is_active)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING created_at, updated_at`

	err := repository.DB.QueryRow(ctx, query,
		category.ID, category.UserID, category.Name, category.Type, category.IsDefault, category.IsActive,
	).Scan(&category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return category, err
	}
	return category, nil
}

func (repository *CategoryRepositoryImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	query := `
	SELECT id, user_id, name, type, is_default, is_active, created_at, updated_at
	FROM categories
	WHERE user_id = $1 AND is_active = true
	ORDER BY created_at ASC`

	rows, err := repository.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.IsDefault, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, rows.Err()
}

func (repository *CategoryRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	query := `
	SELECT id, user_id, name, type, is_default, is_active, created_at, updated_at
	FROM categories WHERE id = $1`

	var c domain.Category
	err := repository.DB.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Type, &c.IsDefault, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (repository *CategoryRepositoryImpl) Update(ctx context.Context, id uuid.UUID, name string) (domain.Category, error) {
	query := `
	UPDATE categories SET name = $1, updated_at = CURRENT_TIMESTAMP
	WHERE id = $2
	RETURNING id, user_id, name, type, is_default, is_active, created_at, updated_at`

	var c domain.Category
	err := repository.DB.QueryRow(ctx, query, name, id).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Type, &c.IsDefault, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (repository *CategoryRepositoryImpl) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE categories SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := repository.DB.Exec(ctx, query, id)
	return err
}

func (repository *CategoryRepositoryImpl) HasTransactions(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE category_id = $1)`

	var exists bool
	err := repository.DB.QueryRow(ctx, query, categoryID).Scan(&exists)
	return exists, err
}