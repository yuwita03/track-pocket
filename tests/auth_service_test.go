package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"trackpocket/internal/model/domain"
	"trackpocket/internal/model/web"
	"trackpocket/internal/service"
)

// 1. Mocking UserRepository Interface
type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) Register(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, mock.AnythingOfType("domain.User"))
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *UserRepositoryMock) FindByEmail(ctx context.Context, email string)(domain.User, error){
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
	
}

// 2. Test Case SUCCESS
func TestRegister_Success(t *testing.T) {
	// Setup
	userRepoMock := new(UserRepositoryMock)
	authService := service.NewAuthService(userRepoMock, "secret_key_123", 1*time.Hour)

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

	// Mocking Behavior: Jika repo dipanggil, kembalikan user & nil error
	userRepoMock.On("Register", mock.Anything, mock.Anything).Return(expectedUser, nil)

	// Execute
	response, err := authService.Register(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, expectedUser.ID, response.User.ID)
	assert.Equal(t, req.Name, response.User.Name)
	assert.Equal(t, req.Email, response.User.Email)

	userRepoMock.AssertExpectations(t)
}

// 3. Test Case FAILED (Database / Repository Error)
func TestRegister_Failed_RepositoryError(t *testing.T) {
	// Setup
	userRepoMock := new(UserRepositoryMock)
	authService := service.NewAuthService(userRepoMock, "secret_key_123", 1*time.Hour)

	req := web.RegisterRequest{
		Name:     "Yurico",
		Email:    "yurico@example.com",
		Password: "password123",
	}

	expectedErr := errors.New("email already exists")

	// Mocking Behavior: Pura-pura database gagal/error (misal email duplikat)
	userRepoMock.On("Register", mock.Anything, mock.Anything).Return(domain.User{}, expectedErr)

	// Execute
	response, err := authService.Register(context.Background(), req)

	// Assertions
	assert.Error(t, err)                             // Harus mengembalikan error
	assert.Equal(t, expectedErr.Error(), err.Error()) // Pesan error harus sama
	assert.Empty(t, response.Token)                  // Token tidak boleh dibuat/kosong
	assert.Equal(t, uuid.Nil, response.User.ID)      // User ID kosong

	userRepoMock.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	userRepoMock := new(UserRepositoryMock)
	authService := service.NewAuthService(userRepoMock, "secret_key_123", 1*time.Hour)

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

	// Mock find user by email
	userRepoMock.On("FindByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	response, err := authService.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, existingUser.ID, response.User.ID)
	assert.Equal(t, existingUser.Email, response.User.Email)

	userRepoMock.AssertExpectations(t)
}

func TestLogin_Failed_UserNotFound(t *testing.T) {
	userRepoMock := new(UserRepositoryMock)
	authService := service.NewAuthService(userRepoMock, "secret_key_123", 1*time.Hour)

	req := web.LoginRequest{
		Email:    "wrong@example.com",
		Password: "password123",
	}

	// Mock DB return user tidak ditemukan
	userRepoMock.On("FindByEmail", mock.Anything, req.Email).Return(domain.User{}, errors.New("user not found"))

	response, err := authService.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, "invalid email or password", err.Error())
	assert.Empty(t, response.Token)

	userRepoMock.AssertExpectations(t)
}

func TestLogin_Failed_WrongPassword(t *testing.T) {
	userRepoMock := new(UserRepositoryMock)
	authService := service.NewAuthService(userRepoMock, "secret_key_123", 1*time.Hour)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

	req := web.LoginRequest{
		Email:    "yurico@example.com",
		Password: "wrong_password", // Password salah
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