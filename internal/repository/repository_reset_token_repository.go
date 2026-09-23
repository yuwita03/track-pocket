package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"trackpocket/internal/model/domain"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token domain.PasswordResetToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (domain.PasswordResetToken, error)
	MarkUsed(ctx context.Context, id string) error
}

type PasswordResetTokenRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewPasswordResetTokenRepository(db *pgxpool.Pool) PasswordResetTokenRepository {
	return &PasswordResetTokenRepositoryImpl{DB: db}
}

func (repository *PasswordResetTokenRepositoryImpl) Create(ctx context.Context, token domain.PasswordResetToken) error {
	query := `
	INSERT INTO password_reset_tokens (id, user_id, token_hash, used, expired_at)
	VALUES ($1, $2, $3, $4, $5)`

	_, err := repository.DB.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.Used, token.ExpiredAt)
	return err
}

func (repository *PasswordResetTokenRepositoryImpl) FindByTokenHash(ctx context.Context, tokenHash string) (domain.PasswordResetToken, error) {
	query := `
	SELECT id, user_id, token_hash, used, expired_at, created_at
	FROM password_reset_tokens WHERE token_hash = $1`

	var t domain.PasswordResetToken
	err := repository.DB.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.Used, &t.ExpiredAt, &t.CreatedAt,
	)
	if err != nil {
		return t, err
	}
	return t, nil
}

func (repository *PasswordResetTokenRepositoryImpl) MarkUsed(ctx context.Context, id string) error {
	query := `UPDATE password_reset_tokens SET used = true WHERE id = $1`
	_, err := repository.DB.Exec(ctx, query, id)
	return err
}