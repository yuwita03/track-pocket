package web

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"Token"`
}

type UserResponse struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
