package web

type RegisterRequest struct {
	Name     string `json:"name" validate:"required, max=100"`
	Email    string `json:"email" validate:"required, max=100"`
	Password string `json:"password" validate:"required, max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required, email"`
	Password string `json:"password" validate:"required"`
}
