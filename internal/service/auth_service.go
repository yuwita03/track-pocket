package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"trackpocket/internal/email" 

	"trackpocket/internal/exception"
	"trackpocket/internal/model/domain"
	"trackpocket/internal/model/web"
	"trackpocket/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, request web.RegisterRequest) (web.AuthResponse, error)
	Login(ctx context.Context, request web.LoginRequest) (web.AuthResponse, error)
	GetProfile(ctx context.Context, userID string) (web.UserResponse, error)
	RefreshToken(ctx context.Context, rawToken string) (web.AuthResponse, error)
	Logout(ctx context.Context, rawToken string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, rawToken string, newPassword string) error
	VerifyEmail(ctx context.Context, rawToken string) error
	ResendVerification(ctx context.Context, email string) error
}

type AuthServiceImpl struct {
	UserRepository               repository.UserRepository
	RefreshTokenRepository       repository.RefreshTokenRepository
	PasswordResetRepository      repository.PasswordResetTokenRepository
	EmailVerificationRepository  repository.EmailVerificationTokenRepository
	CategoryService              CategoryService // ← tambah
	EmailSender                  email.Sender
	JWTSecret                    string
	TokenExpiry                  time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	passwordResetRepo repository.PasswordResetTokenRepository,
	emailVerificationRepo repository.EmailVerificationTokenRepository,
	categoryService CategoryService,
	emailSender email.Sender,
	jwtSecret string,
	tokenExpiry time.Duration,
) AuthService {
	return &AuthServiceImpl{
		UserRepository:              userRepo,
		RefreshTokenRepository:      refreshTokenRepo,
		PasswordResetRepository:     passwordResetRepo,
		EmailVerificationRepository: emailVerificationRepo,
		CategoryService:             categoryService,
		EmailSender:                 emailSender,
		JWTSecret:                   jwtSecret,
		TokenExpiry:                 tokenExpiry,
	}
}

func generateRawRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashRefreshToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
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

func (service *AuthServiceImpl) issueRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	rawToken, err := generateRawRefreshToken()
	if err != nil {
		return "", err
	}

	refreshToken := repository.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID.String(),
		TokenHash: hashRefreshToken(rawToken),
		Revoked:   false,
		ExpiredAt: time.Now().Add(7 * 24 * time.Hour), // 7 hari
	}

	if err := service.RefreshTokenRepository.Create(ctx, &refreshToken); err != nil {
		return "", err
	}

	return rawToken, nil
}

func (service *AuthServiceImpl) Register(ctx context.Context, request web.RegisterRequest) (web.AuthResponse, error) {
	_, err := service.UserRepository.FindByEmail(ctx, request.Email)
	if err == nil {
		return web.AuthResponse{}, exception.New(http.StatusConflict, "EMAIL_ALREADY_EXISTS", "email already registered")
	}

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

	token, err := service.generateToken(savedUser.ID)
	if err != nil {
		return web.AuthResponse{}, err
	}
	refreshToken, err := service.issueRefreshToken(ctx, savedUser.ID)
	if err != nil {
		return web.AuthResponse{}, err
	}

	if err := service.issueEmailVerificationToken(ctx, savedUser); err != nil {
		return web.AuthResponse{}, err
	}
	if err := service.CategoryService.SeedDefaultCategories(ctx, savedUser.ID); err != nil { // ← tambah
		return web.AuthResponse{}, err
	}

	return web.AuthResponse{
		User: web.UserResponse{
			ID:    savedUser.ID,
			Name:  savedUser.Name,
			Email: savedUser.Email,
		},
		Token: token,
		RefreshToken: refreshToken,
	}, nil
}

func (service *AuthServiceImpl) Login(ctx context.Context, request web.LoginRequest) (web.AuthResponse, error) {
	user, err := service.UserRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		return web.AuthResponse{}, exception.New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		return web.AuthResponse{}, exception.New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
	}

	token, err := service.generateToken(user.ID)
	if err != nil {
		return web.AuthResponse{}, err
	}

	refreshToken, err := service.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return web.AuthResponse{}, err
	}

	return web.AuthResponse{
		User: web.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Token: token,
		RefreshToken: refreshToken,
	}, nil
}

func (service *AuthServiceImpl) GetProfile(ctx context.Context, userID string) (web.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return web.UserResponse{}, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
	}

	user, err := service.UserRepository.FindByID(ctx, id)
	if err != nil {
		return web.UserResponse{}, exception.New(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	return web.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (service *AuthServiceImpl) RefreshToken(ctx context.Context, rawToken string)(web.AuthResponse, error){
	tokenHash := hashRefreshToken(rawToken)

	stored, err := service.RefreshTokenRepository.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return web.AuthResponse{}, exception.New(http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "invalid refresh token")
	}

	if stored.Revoked {
		service.RefreshTokenRepository.RevokeAllByUserID(ctx, stored.UserID)
		return web.AuthResponse{}, exception.New(http.StatusUnauthorized, "REFRESH_TOKEN_REUSED", "refresh token reuse detected, all sessions revoked")
	}

	if time.Now().After(stored.ExpiredAt) {
		return web.AuthResponse{}, exception.New(http.StatusUnauthorized, "REFRESH_TOKEN_EXPIRED", "refresh token expired")
	}

	if err := service.RefreshTokenRepository.Revoke(ctx, stored.ID); err != nil {
		return web.AuthResponse{}, err
	}

	userID, err := uuid.Parse(stored.UserID)
	if err != nil {
		return web.AuthResponse{}, err
	}

	user, err := service.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return web.AuthResponse{}, exception.New(http.StatusNotFound, "USER_NOT_FOUND", "user not found")
	}

	newAccesToken, err := service.generateToken(user.ID)
	if err != nil {
		return web.AuthResponse{}, err
	}

	newRefreshToken, err := service.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return web.AuthResponse{}, err
	}

	return web.AuthResponse{
		User: web.UserResponse{
			ID: user.ID,
			Name: user.Name,
			Email: user.Email,
		},
		Token: newAccesToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (service *AuthServiceImpl) Logout(ctx context.Context, rawToken string)error{
	tokenHash := hashRefreshToken(rawToken)

	stored, err := service.RefreshTokenRepository.FindByTokenHash(ctx, tokenHash)
	if err!=nil {
		return nil
	}

	return service.RefreshTokenRepository.Revoke(ctx, stored.ID)
}

func (service *AuthServiceImpl) ForgotPassword(ctx context.Context, email string) error {
	user, err := service.UserRepository.FindByEmail(ctx, email)
	if err != nil {
		// Sengaja gak return error - biar gak bisa dipakai buat enumerasi email terdaftar
		return nil
	}

	rawToken, err := generateRawRefreshToken() // reuse fungsi random token yang udah ada
	if err != nil {
		return err
	}
	tokenHash := hashRefreshToken(rawToken) // reuse fungsi hash yang udah ada

	resetToken := domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		Used:      false,
		ExpiredAt: time.Now().Add(20 * time.Minute),
	}

	if err := service.PasswordResetRepository.Create(ctx, resetToken); err != nil {
		return err
	}

	return service.EmailSender.SendPasswordResetEmail(user.Email, rawToken)
}

func (service *AuthServiceImpl) ResetPassword(ctx context.Context, rawToken string, newPassword string) error {
	tokenHash := hashRefreshToken(rawToken) // reuse fungsi hash yang sama kayak refresh token

	resetToken, err := service.PasswordResetRepository.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return exception.New(http.StatusBadRequest, "INVALID_TOKEN", "invalid or expired reset token")
	}

	if resetToken.Used {
		return exception.New(http.StatusBadRequest, "TOKEN_ALREADY_USED", "reset token already used")
	}

	if time.Now().After(resetToken.ExpiredAt) {
		return exception.New(http.StatusBadRequest, "TOKEN_EXPIRED", "reset token expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := service.UserRepository.UpdatePassword(ctx, resetToken.UserID, string(hashedPassword)); err != nil {
		return err
	}

	if err := service.PasswordResetRepository.MarkUsed(ctx, resetToken.ID.String()); err != nil {
		return err
	}

	// Paksa re-login semua device setelah password direset
	return service.RefreshTokenRepository.RevokeAllByUserID(ctx, resetToken.UserID.String())
}

func (service *AuthServiceImpl) issueEmailVerificationToken(ctx context.Context, user domain.User) error {
	rawToken, err := generateRawRefreshToken()
	if err != nil {
		return err
	}
	tokenHash := hashRefreshToken(rawToken)

	verifyToken := domain.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		Used:      false,
		ExpiredAt: time.Now().Add(24 * time.Hour), // token verifikasi umurnya lebih panjang dari reset password
	}

	if err := service.EmailVerificationRepository.Create(ctx, verifyToken); err != nil {
		return err
	}

	return service.EmailSender.SendVerificationEmail(user.Email, rawToken)
}

func (service *AuthServiceImpl) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := hashRefreshToken(rawToken)

	verifyToken, err := service.EmailVerificationRepository.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return exception.New(http.StatusBadRequest, "INVALID_TOKEN", "invalid or expired verification token")
	}

	if verifyToken.Used {
		return exception.New(http.StatusBadRequest, "TOKEN_ALREADY_USED", "verification token already used")
	}

	if time.Now().After(verifyToken.ExpiredAt) {
		return exception.New(http.StatusBadRequest, "TOKEN_EXPIRED", "verification token expired")
	}

	if err := service.UserRepository.MarkEmailVerified(ctx, verifyToken.UserID); err != nil {
		return err
	}

	return service.EmailVerificationRepository.MarkUsed(ctx, verifyToken.ID.String())
}

func (service *AuthServiceImpl) ResendVerification(ctx context.Context, email string) error {
	user, err := service.UserRepository.FindByEmail(ctx, email)
	if err != nil {
		// sama kayak forgot-password, sengaja gak bocorin apakah email terdaftar
		return nil
	}

	if user.IsVerified {
		// udah verified, gak perlu kirim lagi - tapi tetep return sukses generic
		return nil
	}

	return service.issueEmailVerificationToken(ctx, user)
}