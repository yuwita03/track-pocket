package web

import (
	"github.com/google/uuid"
)

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type UserResponse struct {
	ID    uuid.UUID   	`json:"id"`
	Name  string 		`json:"name"`
	Email string 		`json:"email"`
}
