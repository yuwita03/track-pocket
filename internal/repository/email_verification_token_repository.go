package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"trackpocket/internal/model/domain"
)

type EmailVerificationTokenRepository interface {
	Create(ctx context.Context, token domain.EmailVerificationToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (domain.EmailVerificationToken, error)
	MarkUsed(ctx context.Context, id string) error
}

type EmailVerificationTokenRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewEmailVerificationTokenRepository(db *pgxpool.Pool) EmailVerificationTokenRepository {
	return &EmailVerificationTokenRepositoryImpl{DB: db}
}

func (repository *EmailVerificationTokenRepositoryImpl) Create(ctx context.Context, token domain.EmailVerificationToken) error {
	query := `
	INSERT INTO email_verification_tokens (id, user_id, token_hash, used, expired_at)
	VALUES ($1, $2, $3, $4, $5)`

	_, err := repository.DB.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.Used, token.ExpiredAt)
	return err
}

func (repository *EmailVerificationTokenRepositoryImpl) FindByTokenHash(ctx context.Context, tokenHash string) (domain.EmailVerificationToken, error) {
	query := `
	SELECT id, user_id, token_hash, used, expired_at, created_at
	FROM email_verification_tokens WHERE token_hash = $1`

	var t domain.EmailVerificationToken
	err := repository.DB.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.Used, &t.ExpiredAt, &t.CreatedAt,
	)
	if err != nil {
		return t, err
	}
	return t, nil
}

func (repository *EmailVerificationTokenRepositoryImpl) MarkUsed(ctx context.Context, id string) error {
	query := `UPDATE email_verification_tokens SET used = true WHERE id = $1`
	_, err := repository.DB.Exec(ctx, query, id)
	return err
}