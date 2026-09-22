package repository

import (
	"context"
	"trackpocket/internal/model/domain"

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

func (response *UserRepositoryImpl) Register(ctx context.Context, user domain.User) (domain.User, error) {
	query := 
	`INSERT INTO users (name, email, password_hash, is_Verified)
	VALUES ($1 $2 $3 $4) 
	RETURNING id, created_at, updated_at`

	err:=response.DB.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash).
	Scan(&user.ID, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return user, err
	}
	return user, nil
}

func (response *UserRepositoryImpl) FindByEmail(ctx context.Context, email string)(domain.User, error){
	query := `
	SELECT id, name, email, password_hash, is_verified, creted_at, updated_at
	FROM users WHERE email = $1`

	var user domain.User

	err := response.DB.QueryRow(ctx, query, user.Email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		user.IsVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}