package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"trackpocket/internal/model/domain"
	"trackpocket/internal/model/web"
	"trackpocket/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, request web.RegisterRequest) (web.AuthResponse, error)
	Login(ctx context.Context, request web.LoginRequest) (web.AuthResponse, error)
}

type AuthServiceImpl struct {
	UserRepository repository.UserRepository
	JWTSecret		string
	TokenExpiry		time.Duration
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, tokenExpiry time.Duration) AuthService {
	return &AuthServiceImpl{
		UserRepository: userRepo,
		JWTSecret:      jwtSecret,
		TokenExpiry:    tokenExpiry,
	}
}

func (service *AuthServiceImpl) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":		time.Now().Add(service.TokenExpiry).Unix(), //unix TimeStamp
		"iat":		time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(service.JWTSecret))

	if err != nil {
		return "", err
	}

	return	signed, nil
}


func (service *AuthServiceImpl) Register(ctx context.Context, request web.RegisterRequest) (web.AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return web.AuthResponse{}, err
	}

	user := domain.User{
		Name:         request.Name,
		Email:        request.Email,
		PasswordHash: string(hashedPassword),
		IsVerified:   false,
	}

	savedUser, err := service.UserRepository.Register(ctx, user)
	if err != nil {
		return web.AuthResponse{}, err
	}

	// Memanggil fungsi generateToken
	token, err := service.generateToken(savedUser.ID)
	if err != nil {
		return web.AuthResponse{}, err
	}

	return web.AuthResponse{
		User: web.UserResponse{
			ID:    savedUser.ID,
			Name:  savedUser.Name,
			Email: savedUser.Email,
		},
		Token: token,
	}, nil
}

func (service *AuthServiceImpl) Login(ctx context.Context, request web.LoginRequest) (web.AuthResponse, error){
	user, err := service.UserRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		return web.AuthResponse{}, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		return web.AuthResponse{}, errors.New("invalid email or password")
	}

	token, err := service.generateToken(user.ID)

	return web.AuthResponse{
		User: web.UserResponse{
			ID:	user.ID,
			Name: user.Name,
			Email: user.Email,
		}, Token: token,
	}, nil

}