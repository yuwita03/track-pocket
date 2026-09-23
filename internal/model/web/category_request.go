package web

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,max=100"`
	Type string `json:"type" validate:"required,oneof=INCOME EXPENSE"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}