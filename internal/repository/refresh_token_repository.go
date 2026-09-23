package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	Revoked   bool
	ExpiredAt time.Time
	CreatedAt time.Time
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
}

type RefreshTokenRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) RefreshTokenRepository {
	return &RefreshTokenRepositoryImpl{DB: db}
}

func (repository *RefreshTokenRepositoryImpl) Create(ctx context.Context, rt *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, revoked, expired_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := repository.DB.Exec(ctx, query, rt.ID, rt.UserID, rt.TokenHash, rt.Revoked, rt.ExpiredAt, time.Now())
	return err
}

func (repository *RefreshTokenRepositoryImpl) FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, revoked, expired_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	row := repository.DB.QueryRow(ctx, query, tokenHash)

	var rt RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.Revoked, &rt.ExpiredAt, &rt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (repository *RefreshTokenRepositoryImpl) Revoke(ctx context.Context, id string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	_, err := repository.DB.Exec(ctx, query, id)
	return err
}

func (repository *RefreshTokenRepositoryImpl) RevokeAllByUserID(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := repository.DB.Exec(ctx, query, userID)
	return err
}