package service_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"trackpocket/internal/model/domain"
	"trackpocket/internal/model/web"
	"trackpocket/internal/repository"
	"trackpocket/internal/service"
)

// ============ HELPERS ============

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func newTestAuthService() (
	service.AuthService,
	*UserRepositoryMock,
	*RefreshTokenRepositoryMock,
	*PasswordResetTokenRepositoryMock,
	*EmailVerificationTokenRepositoryMock,
	*EmailSenderMock,
) {
	userRepoMock := new(UserRepositoryMock)
	refreshTokenRepoMock := new(RefreshTokenRepositoryMock)
	passwordResetRepoMock := new(PasswordResetTokenRepositoryMock)
	emailVerificationRepoMock := new(EmailVerificationTokenRepositoryMock)
	emailSenderMock := new(EmailSenderMock)

	authService := service.NewAuthService(
		userRepoMock,
		refreshTokenRepoMock,
		passwordResetRepoMock,
		emailVerificationRepoMock,
		emailSenderMock,
		"secret_key_123",
		1*time.Hour,
	)

	return authService, userRepoMock, refreshTokenRepoMock, passwordResetRepoMock, emailVerificationRepoMock, emailSenderMock
}

// ============ MOCKS ============

type UserRepositoryMock struct{ mock.Mock }

func (m *UserRepositoryMock) Register(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *UserRepositoryMock) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *UserRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *UserRepositoryMock) UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	args := m.Called(ctx, userID, newPasswordHash)
	return args.Error(0)
}

func (m *UserRepositoryMock) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type RefreshTokenRepositoryMock struct{ mock.Mock }

func (m *RefreshTokenRepositoryMock) Create(ctx context.Context, rt *repository.RefreshToken) error {
	args := m.Called(ctx, rt)
	return args.Error(0)
}

func (m *RefreshTokenRepositoryMock) FindByTokenHash(ctx context.Context, tokenHash string) (*repository.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.RefreshToken), args.Error(1)
}

func (m *RefreshTokenRepositoryMock) Revoke(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *RefreshTokenRepositoryMock) RevokeAllByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type PasswordResetTokenRepositoryMock struct{ mock.Mock }

func (m *PasswordResetTokenRepositoryMock) Create(ctx context.Context, token domain.PasswordResetToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *PasswordResetTokenRepositoryMock) FindByTokenHash(ctx context.Context, tokenHash string) (domain.PasswordResetToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(domain.PasswordResetToken), args.Error(1)
}

func (m *PasswordResetTokenRepositoryMock) MarkUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type EmailVerificationTokenRepositoryMock struct{ mock.Mock }

func (m *EmailVerificationTokenRepositoryMock) Create(ctx context.Context, token domain.EmailVerificationToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *EmailVerificationTokenRepositoryMock) FindByTokenHash(ctx context.Context, tokenHash string) (domain.EmailVerificationToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(domain.EmailVerificationToken), args.Error(1)
}

func (m *EmailVerificationTokenRepositoryMock) MarkUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type EmailSenderMock struct{ mock.Mock }

func (m *EmailSenderMock) SendPasswordResetEmail(to string, resetToken string) error {
	args := m.Called(to, resetToken)
	return args.Error(0)
}

func (m *EmailSenderMock) SendVerificationEmail(to string, verificationToken string) error {
	args := m.Called(to, verificationToken)
	return args.Error(0)
}

// ============ REGISTER ============

func TestRegister_Success(t *testing.T) {
	authService, userRepoMock, refreshTokenRepoMock, _, emailVerificationRepoMock, emailSenderMock := newTestAuthService()

	req := web.RegisterRequest{
		Name:     "Yurico",
		Email:    "yurico@example.com",
		Password: "password123",
	}

	expectedUser := domain.User{
		ID:    uuid.New(),
		Name:  req.Name,
		Email: req.Email,
	}

	userRepoMock.On("FindByEmail", mock.Anything, req.Email).Return(domain.User{}, errors.New("user not found"))
	userRepoMock.On("Register", mock.Anything, mock.Anything).Return(expectedUser, nil)
	refreshTokenRepoMock.On("Create", mock.Anything, mock.Anything).Return(nil)
	emailVerificationRepoMock.On("Create", mock.Anything, mock.Anything).Return(nil)
	emailSenderMock.On("SendVerificationEmail", expectedUser.Email, mock.Anything).Return(nil)

	response, err := authService.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, expectedUser.ID, response.User.ID)
	assert.Equal(t, req.Name, response.User.Name)
	assert.Equal(t, req.Email, response.User.Email)

	userRepoMock.AssertExpectations(t)
	refreshTokenRepoMock.AssertExpectations(t)
	emailVerificationRepoMock.AssertExpectations(t)
	emailSenderMock.AssertExpectations(t)
}

func TestRegister_Failed_EmailAlreadyExists(t *testing.T) {
	authService, userRepoMock, _, _, _, _ := newTestAuthService()

	req := web.RegisterRequest{
		Name:     "Yurico",
		Email:    "yurico@example.com",
		Password: "password123",
	}

	existingUser := domain.User{ID: uuid.New(), Email: req.Email}

	userRepoMock.On("FindByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	response, err := authService.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Empty(t, response.Token)
	assert.Equal(t, uuid.Nil, response.User.ID)

	userRepoMock.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	userRepoMock.AssertExpectations(t)
}

func TestRegister_Failed_RepositoryError(t *testing.T) {
	authService, userRepoMock, _, _, _, _ := newTestAuthService()

	req := web.RegisterRequest{
		Name:     "Yurico",
		Email:    "yurico@example.com",
		Password: "password123",
	}

	expectedErr := errors.New("database connection failed")

	userRepoMock.On("FindByEmail", mock.Anything, req.Email).Return(domain.User{}, errors.New("user not found"))
	userRepoMock.On("Register", mock.Anything, mock.Anything).Return(domain.User{}, expectedErr)

	response, err := authService.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	assert.Empty(t, response.Token)
	assert.Equal(t, uuid.Nil, response.User.ID)

	userRepoMock.AssertExpectations(t)
}

// ============ LOGIN ============

func TestLogin_Success(t *testing.T) {
	authService, userRepoMock, refreshTokenRepoMock, _, _, _ := newTestAuthService()

	passwordPlain := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(passwordPlain), bcrypt.DefaultCost)

	req := web.LoginRequest{
		Email:    "yurico@example.com",
		Password: passwordPlain,
	}

	existingUser := domain.User{
		ID:           uuid.New(),
		Name:         "Yurico",
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	userRepoMock.On("FindByEmail", mock.Anything, req.Email).Return(existingUser, nil)
	refreshTokenRepoMock.On("Create", mock.Anything, mock.Anything).Return(nil)

	response, err := authService.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, existingUser.ID, response.User.ID)
	assert.Equal(t, existingUser.Email, response.User.Email)

	userRepoMock.AssertExpectations(t)
	refreshTokenRepoMock.AssertExpectations(t)
}

func TestLogin_Failed_UserNotFound(t *testing.T) {
	authService, userRepoMock, _, _, _, _ := newTestAuthService()

	req := web.LoginRequest{
		Email:    "wrong@example.com",
		Password: "password123",
	}

	userRepoMock.On("FindByEmail", mock.Anything, req.Email).Return(domain.User{}, errors.New("user not found"))

	response, err := authService.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, "invalid email or password", err.Error())
	assert.Empty(t, response.Token)

	userRepoMock.AssertExpectations(t)
}

func TestLogin_Failed_WrongPassword(t *testing.T) {
	authService, userRepoMock, _, _, _, _ := newTestAuthService()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

	req := web.LoginRequest{
		Email:    "yurico@example.com",
		Password: "wrong_password",
	}

	existingUser := domain.User{
		ID:           uuid.New(),
		Name:         "Yurico",
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	userRepoMock.On("FindByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	response, err := authService.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, "invalid email or password", err.Error())
	assert.Empty(t, response.Token)

	userRepoMock.AssertExpectations(t)
}

// ============ GET PROFILE ============

func TestGetProfile_Success(t *testing.T) {
	authService, userRepoMock, _, _, _, _ := newTestAuthService()

	userID := uuid.New()
	existingUser := domain.User{
		ID:    userID,
		Name:  "Yurico",
		Email: "yurico@example.com",
	}

	userRepoMock.On("FindByID", mock.Anything, userID).Return(existingUser, nil)

	result, err := authService.GetProfile(context.Background(), userID.String())

	assert.NoError(t, err)
	assert.Equal(t, existingUser.ID, result.ID)
	assert.Equal(t, existingUser.Name, result.Name)
	assert.Equal(t, existingUser.Email, result.Email)

	userRepoMock.AssertExpectations(t)
}

func TestGetProfile_Failed_InvalidUUID(t *testing.T) {
	authService, userRepoMock, _, _, _, _ := newTestAuthService()

	result, err := authService.GetProfile(context.Background(), "bukan-uuid-valid")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result.ID)

	userRepoMock.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}

func TestGetProfile_Failed_UserNotFound(t *testing.T) {
	authService, userRepoMock, _, _, _, _ := newTestAuthService()

	userID := uuid.New()

	userRepoMock.On("FindByID", mock.Anything, userID).Return(domain.User{}, errors.New("not found"))

	result, err := authService.GetProfile(context.Background(), userID.String())

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result.ID)

	userRepoMock.AssertExpectations(t)
}

// ============ REFRESH TOKEN ============

func TestRefreshToken_Success(t *testing.T) {
	authService, userRepoMock, refreshTokenRepoMock, _, _, _ := newTestAuthService()

	userID := uuid.New()
	rawToken := "some-raw-refresh-token"
	tokenHash := sha256Hex(rawToken)

	stored := &repository.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID.String(),
		TokenHash: tokenHash,
		Revoked:   false,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}

	existingUser := domain.User{ID: userID, Name: "Yurico", Email: "yurico@example.com"}

	refreshTokenRepoMock.On("FindByTokenHash", mock.Anything, tokenHash).Return(stored, nil)
	refreshTokenRepoMock.On("Revoke", mock.Anything, stored.ID).Return(nil)
	userRepoMock.On("FindByID", mock.Anything, userID).Return(existingUser, nil)
	refreshTokenRepoMock.On("Create", mock.Anything, mock.Anything).Return(nil)

	response, err := authService.RefreshToken(context.Background(), rawToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, userID, response.User.ID)

	refreshTokenRepoMock.AssertExpectations(t)
	userRepoMock.AssertExpectations(t)
}

func TestRefreshToken_Failed_InvalidToken(t *testing.T) {
	authService, _, refreshTokenRepoMock, _, _, _ := newTestAuthService()

	refreshTokenRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	response, err := authService.RefreshToken(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Empty(t, response.Token)

	refreshTokenRepoMock.AssertExpectations(t)
}

func TestRefreshToken_Failed_Reused(t *testing.T) {
	authService, _, refreshTokenRepoMock, _, _, _ := newTestAuthService()

	stored := &repository.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Revoked:   true,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}

	refreshTokenRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(stored, nil)
	refreshTokenRepoMock.On("RevokeAllByUserID", mock.Anything, stored.UserID).Return(nil)

	response, err := authService.RefreshToken(context.Background(), "reused-token")

	assert.Error(t, err)
	assert.Empty(t, response.Token)

	refreshTokenRepoMock.AssertExpectations(t)
}

func TestRefreshToken_Failed_Expired(t *testing.T) {
	authService, _, refreshTokenRepoMock, _, _, _ := newTestAuthService()

	stored := &repository.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    uuid.New().String(),
		Revoked:   false,
		ExpiredAt: time.Now().Add(-1 * time.Hour),
	}

	refreshTokenRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(stored, nil)

	response, err := authService.RefreshToken(context.Background(), "expired-token")

	assert.Error(t, err)
	assert.Empty(t, response.Token)

	refreshTokenRepoMock.AssertExpectations(t)
}

// ============ LOGOUT ============

func TestLogout_Success(t *testing.T) {
	authService, _, refreshTokenRepoMock, _, _, _ := newTestAuthService()

	stored := &repository.RefreshToken{ID: uuid.New().String()}

	refreshTokenRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(stored, nil)
	refreshTokenRepoMock.On("Revoke", mock.Anything, stored.ID).Return(nil)

	err := authService.Logout(context.Background(), "some-token")

	assert.NoError(t, err)
	refreshTokenRepoMock.AssertExpectations(t)
}

func TestLogout_TokenNotFound_StillSucceeds(t *testing.T) {
	authService, _, refreshTokenRepoMock, _, _, _ := newTestAuthService()

	refreshTokenRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	err := authService.Logout(context.Background(), "unknown-token")

	assert.NoError(t, err)
	refreshTokenRepoMock.AssertExpectations(t)
}

// ============ FORGOT PASSWORD ============

func TestForgotPassword_Success(t *testing.T) {
	authService, userRepoMock, _, passwordResetRepoMock, _, emailSenderMock := newTestAuthService()

	existingUser := domain.User{ID: uuid.New(), Email: "yurico@example.com"}

	userRepoMock.On("FindByEmail", mock.Anything, existingUser.Email).Return(existingUser, nil)
	passwordResetRepoMock.On("Create", mock.Anything, mock.Anything).Return(nil)
	emailSenderMock.On("SendPasswordResetEmail", existingUser.Email, mock.Anything).Return(nil)

	err := authService.ForgotPassword(context.Background(), existingUser.Email)

	assert.NoError(t, err)
	userRepoMock.AssertExpectations(t)
	passwordResetRepoMock.AssertExpectations(t)
	emailSenderMock.AssertExpectations(t)
}

func TestForgotPassword_EmailNotFound_StillSucceeds(t *testing.T) {
	authService, userRepoMock, _, passwordResetRepoMock, _, _ := newTestAuthService()

	userRepoMock.On("FindByEmail", mock.Anything, "notfound@example.com").Return(domain.User{}, errors.New("not found"))

	err := authService.ForgotPassword(context.Background(), "notfound@example.com")

	assert.NoError(t, err)
	passwordResetRepoMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	userRepoMock.AssertExpectations(t)
}

// ============ RESET PASSWORD ============

func TestResetPassword_Success(t *testing.T) {
	authService, userRepoMock, refreshTokenRepoMock, passwordResetRepoMock, _, _ := newTestAuthService()

	rawToken := "raw-reset-token"
	tokenHash := sha256Hex(rawToken)
	userID := uuid.New()

	storedToken := domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		Used:      false,
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}

	passwordResetRepoMock.On("FindByTokenHash", mock.Anything, tokenHash).Return(storedToken, nil)
	userRepoMock.On("UpdatePassword", mock.Anything, userID, mock.Anything).Return(nil)
	passwordResetRepoMock.On("MarkUsed", mock.Anything, storedToken.ID.String()).Return(nil)
	refreshTokenRepoMock.On("RevokeAllByUserID", mock.Anything, userID.String()).Return(nil)

	err := authService.ResetPassword(context.Background(), rawToken, "newpassword123")

	assert.NoError(t, err)
	userRepoMock.AssertExpectations(t)
	passwordResetRepoMock.AssertExpectations(t)
	refreshTokenRepoMock.AssertExpectations(t)
}

func TestResetPassword_Failed_InvalidToken(t *testing.T) {
	authService, _, _, passwordResetRepoMock, _, _ := newTestAuthService()

	passwordResetRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(domain.PasswordResetToken{}, errors.New("not found"))

	err := authService.ResetPassword(context.Background(), "invalid-token", "newpassword123")

	assert.Error(t, err)
	passwordResetRepoMock.AssertExpectations(t)
}

func TestResetPassword_Failed_AlreadyUsed(t *testing.T) {
	authService, _, _, passwordResetRepoMock, _, _ := newTestAuthService()

	storedToken := domain.PasswordResetToken{
		ID:        uuid.New(),
		Used:      true,
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}

	passwordResetRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(storedToken, nil)

	err := authService.ResetPassword(context.Background(), "used-token", "newpassword123")

	assert.Error(t, err)
	passwordResetRepoMock.AssertExpectations(t)
}

func TestResetPassword_Failed_Expired(t *testing.T) {
	authService, _, _, passwordResetRepoMock, _, _ := newTestAuthService()

	storedToken := domain.PasswordResetToken{
		ID:        uuid.New(),
		Used:      false,
		ExpiredAt: time.Now().Add(-10 * time.Minute),
	}

	passwordResetRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(storedToken, nil)

	err := authService.ResetPassword(context.Background(), "expired-token", "newpassword123")

	assert.Error(t, err)
	passwordResetRepoMock.AssertExpectations(t)
}

// ============ VERIFY EMAIL ============

func TestVerifyEmail_Success(t *testing.T) {
	authService, userRepoMock, _, _, emailVerificationRepoMock, _ := newTestAuthService()

	rawToken := "raw-verify-token"
	tokenHash := sha256Hex(rawToken)
	userID := uuid.New()

	storedToken := domain.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		Used:      false,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}

	emailVerificationRepoMock.On("FindByTokenHash", mock.Anything, tokenHash).Return(storedToken, nil)
	userRepoMock.On("MarkEmailVerified", mock.Anything, userID).Return(nil)
	emailVerificationRepoMock.On("MarkUsed", mock.Anything, storedToken.ID.String()).Return(nil)

	err := authService.VerifyEmail(context.Background(), rawToken)

	assert.NoError(t, err)
	userRepoMock.AssertExpectations(t)
	emailVerificationRepoMock.AssertExpectations(t)
}

func TestVerifyEmail_Failed_Expired(t *testing.T) {
	authService, _, _, _, emailVerificationRepoMock, _ := newTestAuthService()

	storedToken := domain.EmailVerificationToken{
		ID:        uuid.New(),
		Used:      false,
		ExpiredAt: time.Now().Add(-1 * time.Hour),
	}

	emailVerificationRepoMock.On("FindByTokenHash", mock.Anything, mock.Anything).Return(storedToken, nil)

	err := authService.VerifyEmail(context.Background(), "expired-token")

	assert.Error(t, err)
	emailVerificationRepoMock.AssertExpectations(t)
}

// ============ RESEND VERIFICATION ============

func TestResendVerification_Success(t *testing.T) {
	authService, userRepoMock, _, _, emailVerificationRepoMock, emailSenderMock := newTestAuthService()

	existingUser := domain.User{ID: uuid.New(), Email: "yurico@example.com", IsVerified: false}

	userRepoMock.On("FindByEmail", mock.Anything, existingUser.Email).Return(existingUser, nil)
	emailVerificationRepoMock.On("Create", mock.Anything, mock.Anything).Return(nil)
	emailSenderMock.On("SendVerificationEmail", existingUser.Email, mock.Anything).Return(nil)

	err := authService.ResendVerification(context.Background(), existingUser.Email)

	assert.NoError(t, err)
	userRepoMock.AssertExpectations(t)
	emailVerificationRepoMock.AssertExpectations(t)
	emailSenderMock.AssertExpectations(t)
}

func TestResendVerification_AlreadyVerified_NoEmailSent(t *testing.T) {
	authService, userRepoMock, _, _, emailVerificationRepoMock, _ := newTestAuthService()

	existingUser := domain.User{ID: uuid.New(), Email: "yurico@example.com", IsVerified: true}

	userRepoMock.On("FindByEmail", mock.Anything, existingUser.Email).Return(existingUser, nil)

	err := authService.ResendVerification(context.Background(), existingUser.Email)

	assert.NoError(t, err)
	emailVerificationRepoMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	userRepoMock.AssertExpectations(t)
}