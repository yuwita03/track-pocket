package repository

import (
	"context"
	"trackpocket/internal/model/domain"
	"github.com/google/uuid"


	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Register(ctx context.Context, user domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

type UserRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &UserRepositoryImpl{DB: db}
}

func (repository *UserRepositoryImpl) Register(ctx context.Context, user domain.User) (domain.User, error) {
	user.ID = uuid.New()
	query := 
	`INSERT INTO users (id, name, email, password_hash, is_Verified)
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING created_at, updated_at`

	err:=repository.DB.QueryRow(ctx, query,user.ID, user.Name, user.Email, user.PasswordHash, user.IsVerified).
	Scan(
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (repository *UserRepositoryImpl) FindByEmail(ctx context.Context, email string)(domain.User, error){
	query := `
	SELECT id, name, email, password_hash, is_verified, created_at, updated_at
	FROM users WHERE email = $1`

	var user domain.User

	err := repository.DB.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}