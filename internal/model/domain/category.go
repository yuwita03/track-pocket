package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Type      string
	IsDefault bool
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}