package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"trackpocket/internal/exception"
	"trackpocket/internal/model/domain"
	"trackpocket/internal/model/web"
	"trackpocket/internal/repository"
)

type TransactionService interface {
	Create(ctx context.Context, userID uuid.UUID, request web.CreateTransactionRequest) (web.TransactionResponse, error)
	FindAll(ctx context.Context, userID uuid.UUID, params web.TransactionQueryParams) (web.PaginatedTransactionResponse, error)
	Update(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID, request web.UpdateTransactionRequest) (web.TransactionResponse, error)
	Delete(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error
}

type TransactionServiceImpl struct {
	TransactionRepository repository.TransactionRepository
	CategoryRepository    repository.CategoryRepository
}

func NewTransactionService(transactionRepo repository.TransactionRepository, categoryRepo repository.CategoryRepository) TransactionService {
	return &TransactionServiceImpl{
		TransactionRepository: transactionRepo,
		CategoryRepository:    categoryRepo,
	}
}

func toTransactionResponse(t domain.Transaction) web.TransactionResponse {
	return web.TransactionResponse{
		ID:              t.ID,
		CategoryID:      t.CategoryID,
		Type:            t.Type,
		Amount:          t.Amount.String(),
		Note:            t.Note,
		TransactionDate: t.TransactionDate,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

func (service *TransactionServiceImpl) validateCategoryOwnership(ctx context.Context, userID uuid.UUID, categoryIDStr string) (uuid.UUID, error) {
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return uuid.Nil, exception.New(http.StatusBadRequest, "INVALID_CATEGORY_ID", "invalid category id")
	}

	category, err := service.CategoryRepository.FindByID(ctx, categoryID)
	if err != nil {
		return uuid.Nil, exception.New(http.StatusNotFound, "CATEGORY_NOT_FOUND", "category not found")
	}

	if category.UserID != userID {
		return uuid.Nil, exception.New(http.StatusForbidden, "FORBIDDEN", "you don't have access to this category")
	}

	return categoryID, nil
}

func (service *TransactionServiceImpl) Create(ctx context.Context, userID uuid.UUID, request web.CreateTransactionRequest) (web.TransactionResponse, error) {
	categoryID, err := service.validateCategoryOwnership(ctx, userID, request.CategoryID)
	if err != nil {
		return web.TransactionResponse{}, err
	}

	amount, err := decimal.NewFromString(request.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return web.TransactionResponse{}, exception.New(http.StatusUnprocessableEntity, "INVALID_AMOUNT", "amount must be a valid positive number")
	}

	transaction := domain.Transaction{
		UserID:          userID,
		CategoryID:      categoryID,
		Type:            request.Type,
		Amount:          amount,
		Note:            request.Note,
		TransactionDate: request.TransactionDate,
	}

	saved, err := service.TransactionRepository.Create(ctx, transaction)
	if err != nil {
		return web.TransactionResponse{}, err
	}

	return toTransactionResponse(saved), nil
}

func (service *TransactionServiceImpl) FindAll(ctx context.Context, userID uuid.UUID, params web.TransactionQueryParams) (web.PaginatedTransactionResponse, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	limit := params.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := repository.TransactionFilter{
		Search: params.Search,
		Type:   params.Type,
		Page:   page,
		Limit:  limit,
	}

	if params.CategoryID != "" {
		categoryID, err := uuid.Parse(params.CategoryID)
		if err != nil {
			return web.PaginatedTransactionResponse{}, exception.New(http.StatusBadRequest, "INVALID_CATEGORY_ID", "invalid category id")
		}
		filter.CategoryID = &categoryID
	}

	if params.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", params.StartDate)
		if err != nil {
			return web.PaginatedTransactionResponse{}, exception.New(http.StatusBadRequest, "INVALID_DATE", "invalid start_date format, use YYYY-MM-DD")
		}
		filter.StartDate = &startDate
	}

	if params.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", params.EndDate)
		if err != nil {
			return web.PaginatedTransactionResponse{}, exception.New(http.StatusBadRequest, "INVALID_DATE", "invalid end_date format, use YYYY-MM-DD")
		}
		filter.EndDate = &endDate
	}

	transactions, totalCount, err := service.TransactionRepository.FindAllByUserID(ctx, userID, filter)
	if err != nil {
		return web.PaginatedTransactionResponse{}, err
	}

	responses := make([]web.TransactionResponse, 0, len(transactions))
	for _, t := range transactions {
		responses = append(responses, toTransactionResponse(t))
	}

	return web.PaginatedTransactionResponse{
		Data:       responses,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
	}, nil
}

func (service *TransactionServiceImpl) Update(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID, request web.UpdateTransactionRequest) (web.TransactionResponse, error) {
	existing, err := service.TransactionRepository.FindByID(ctx, transactionID)
	if err != nil {
		return web.TransactionResponse{}, exception.New(http.StatusNotFound, "TRANSACTION_NOT_FOUND", "transaction not found")
	}

	if existing.UserID != userID {
		return web.TransactionResponse{}, exception.New(http.StatusForbidden, "FORBIDDEN", "you don't have access to this transaction")
	}

	categoryID, err := service.validateCategoryOwnership(ctx, userID, request.CategoryID)
	if err != nil {
		return web.TransactionResponse{}, err
	}

	amount, err := decimal.NewFromString(request.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return web.TransactionResponse{}, exception.New(http.StatusUnprocessableEntity, "INVALID_AMOUNT", "amount must be a valid positive number")
	}

	existing.CategoryID = categoryID
	existing.Type = request.Type
	existing.Amount = amount
	existing.Note = request.Note
	existing.TransactionDate = request.TransactionDate

	updated, err := service.TransactionRepository.Update(ctx, existing)
	if err != nil {
		return web.TransactionResponse{}, err
	}

	return toTransactionResponse(updated), nil
}

func (service *TransactionServiceImpl) Delete(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	existing, err := service.TransactionRepository.FindByID(ctx, transactionID)
	if err != nil {
		return exception.New(http.StatusNotFound, "TRANSACTION_NOT_FOUND", "transaction not found")
	}

	if existing.UserID != userID {
		return exception.New(http.StatusForbidden, "FORBIDDEN", "you don't have access to this transaction")
	}

	return service.TransactionRepository.Delete(ctx, transactionID)
}