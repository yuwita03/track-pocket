package domain

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	Used      bool
	ExpiredAt time.Time
	CreatedAt time.Time
}