package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"trackpocket/internal/exception"
	"trackpocket/internal/model/domain"
	"trackpocket/internal/model/web"
	"trackpocket/internal/repository"
)

type CategoryService interface {
	Create(ctx context.Context, userID uuid.UUID, request web.CreateCategoryRequest) (web.CategoryResponse, error)
	FindAll(ctx context.Context, userID uuid.UUID) ([]web.CategoryResponse, error)
	Update(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, request web.UpdateCategoryRequest) (web.CategoryResponse, error)
	Delete(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) error
	SeedDefaultCategories(ctx context.Context, userID uuid.UUID) error
}

type CategoryServiceImpl struct {
	CategoryRepository repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &CategoryServiceImpl{CategoryRepository: categoryRepo}
}

func toCategoryResponse(c domain.Category) web.CategoryResponse {
	return web.CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		IsDefault: c.IsDefault,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (service *CategoryServiceImpl) Create(ctx context.Context, userID uuid.UUID, request web.CreateCategoryRequest) (web.CategoryResponse, error) {
	category := domain.Category{
		UserID:    userID,
		Name:      request.Name,
		Type:      request.Type,
		IsDefault: false,
		IsActive:  true,
	}

	saved, err := service.CategoryRepository.Create(ctx, category)
	if err != nil {
		return web.CategoryResponse{}, err
	}

	return toCategoryResponse(saved), nil
}

func (service *CategoryServiceImpl) FindAll(ctx context.Context, userID uuid.UUID) ([]web.CategoryResponse, error) {
	categories, err := service.CategoryRepository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]web.CategoryResponse, 0, len(categories))
	for _, c := range categories {
		responses = append(responses, toCategoryResponse(c))
	}

	return responses, nil
}

func (service *CategoryServiceImpl) Update(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, request web.UpdateCategoryRequest) (web.CategoryResponse, error) {
	existing, err := service.CategoryRepository.FindByID(ctx, categoryID)
	if err != nil {
		return web.CategoryResponse{}, exception.New(http.StatusNotFound, "CATEGORY_NOT_FOUND", "category not found")
	}

	// Data isolation - pastiin category ini milik user yang bener
	if existing.UserID != userID {
		return web.CategoryResponse{}, exception.New(http.StatusForbidden, "FORBIDDEN", "you don't have access to this category")
	}

	updated, err := service.CategoryRepository.Update(ctx, categoryID, request.Name)
	if err != nil {
		return web.CategoryResponse{}, err
	}

	return toCategoryResponse(updated), nil
}

func (service *CategoryServiceImpl) Delete(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) error {
	existing, err := service.CategoryRepository.FindByID(ctx, categoryID)
	if err != nil {
		return exception.New(http.StatusNotFound, "CATEGORY_NOT_FOUND", "category not found")
	}

	if existing.UserID != userID {
		return exception.New(http.StatusForbidden, "FORBIDDEN", "you don't have access to this category")
	}

	if existing.IsDefault {
		return exception.New(http.StatusForbidden, "CANNOT_DELETE_DEFAULT", "default category cannot be deleted")
	}

	hasTransactions, err := service.CategoryRepository.HasTransactions(ctx, categoryID)
	if err != nil {
		return err
	}

	if hasTransactions {
		// Masih punya transaksi -> soft delete aja, jangan hard delete
		return service.CategoryRepository.SoftDelete(ctx, categoryID)
	}

	// Gak ada transaksi terkait -> soft delete juga aman, konsisten
	return service.CategoryRepository.SoftDelete(ctx, categoryID)
}

func (service *CategoryServiceImpl) SeedDefaultCategories(ctx context.Context, userID uuid.UUID) error {
	defaults := []struct {
		Name string
		Type string
	}{
		{"Gaji", "INCOME"},
		{"Bonus", "INCOME"},
		{"Makanan", "EXPENSE"},
		{"Transportasi", "EXPENSE"},
		{"Belanja", "EXPENSE"},
		{"Tagihan", "EXPENSE"},
		{"Hiburan", "EXPENSE"},
		{"Lainnya", "EXPENSE"},
	}

	for _, d := range defaults {
		category := domain.Category{
			UserID:    userID,
			Name:      d.Name,
			Type:      d.Type,
			IsDefault: true,
			IsActive:  true,
		}

		if _, err := service.CategoryRepository.Create(ctx, category); err != nil {
			return err
		}
	}

	return nil
}