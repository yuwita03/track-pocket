package service_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"trackpocket/internal/model/domain"
	"trackpocket/internal/repository"
)

// ============ USER REPOSITORY MOCK ============

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

// ============ REFRESH TOKEN REPOSITORY MOCK ============

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

// ============ PASSWORD RESET TOKEN REPOSITORY MOCK ============

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

// ============ EMAIL VERIFICATION TOKEN REPOSITORY MOCK ============

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

// ============ EMAIL SENDER MOCK ============

type EmailSenderMock struct{ mock.Mock }

func (m *EmailSenderMock) SendPasswordResetEmail(to string, resetToken string) error {
	args := m.Called(to, resetToken)
	return args.Error(0)
}

func (m *EmailSenderMock) SendVerificationEmail(to string, verificationToken string) error {
	args := m.Called(to, verificationToken)
	return args.Error(0)
}

// ============ CATEGORY REPOSITORY MOCK ============

type CategoryRepositoryMock struct{ mock.Mock }

func (m *CategoryRepositoryMock) Create(ctx context.Context, c domain.Category) (domain.Category, error) {
	args := m.Called(ctx, c)
	return args.Get(0).(domain.Category), args.Error(1)
}

func (m *CategoryRepositoryMock) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *CategoryRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Category), args.Error(1)
}

func (m *CategoryRepositoryMock) Update(ctx context.Context, id uuid.UUID, name string) (domain.Category, error) {
	args := m.Called(ctx, id, name)
	return args.Get(0).(domain.Category), args.Error(1)
}

func (m *CategoryRepositoryMock) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *CategoryRepositoryMock) HasTransactions(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	args := m.Called(ctx, categoryID)
	return args.Bool(0), args.Error(1)
}

// ============ TRANSACTION REPOSITORY MOCK ============

type TransactionRepositoryMock struct{ mock.Mock }

func (m *TransactionRepositoryMock) Create(ctx context.Context, t domain.Transaction) (domain.Transaction, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) FindAllByUserID(ctx context.Context, userID uuid.UUID, filter repository.TransactionFilter) ([]domain.Transaction, int, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).([]domain.Transaction), args.Int(1), args.Error(2)
}

func (m *TransactionRepositoryMock) Update(ctx context.Context, t domain.Transaction) (domain.Transaction, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}