package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"trackpocket/internal/model/web"
	"trackpocket/internal/repository"
)

type DashboardService interface {
	GetSummary(ctx context.Context, userID uuid.UUID, params web.DashboardQueryParams) (web.DashboardSummaryResponse, error)
	GetCategoryExpenses(ctx context.Context, userID uuid.UUID, params web.DashboardQueryParams) ([]web.CategoryExpenseResponse, error)
}

type DashboardServiceImpl struct {
	DashboardRepository repository.DashboardRepository
}

func NewDashboardService(dashboardRepo repository.DashboardRepository) DashboardService {
	return &DashboardServiceImpl{DashboardRepository: dashboardRepo}
}

func (service *DashboardServiceImpl) GetSummary(ctx context.Context, userID uuid.UUID, params web.DashboardQueryParams) (web.DashboardSummaryResponse, error) {
	year, month := resolveYearMonth(params)

	summary, err := service.DashboardRepository.GetSummary(ctx, userID, year, month)
	if err != nil {
		return web.DashboardSummaryResponse{}, err
	}

	balance := summary.TotalIncome.Sub(summary.TotalExpense)

	return web.DashboardSummaryResponse{
		TotalIncome:      summary.TotalIncome.String(),
		TotalExpense:     summary.TotalExpense.String(),
		Balance:          balance.String(),
		TransactionCount: summary.TransactionCount,
	}, nil
}

func (service *DashboardServiceImpl) GetCategoryExpenses(ctx context.Context, userID uuid.UUID, params web.DashboardQueryParams) ([]web.CategoryExpenseResponse, error) {
	year, month := resolveYearMonth(params)

	expenses, err := service.DashboardRepository.GetCategoryExpenses(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}

	totalExpense := decimal.Zero
	for _, e := range expenses {
		totalExpense = totalExpense.Add(e.Amount)
	}

	responses := make([]web.CategoryExpenseResponse, 0, len(expenses))
	for _, e := range expenses {
		var percentage decimal.Decimal
		if totalExpense.IsZero() {
			percentage = decimal.Zero
		} else {
			percentage = e.Amount.Div(totalExpense).Mul(decimal.NewFromInt(100)).Round(2)
		}

		responses = append(responses, web.CategoryExpenseResponse{
			CategoryID:   e.CategoryID.String(),
			CategoryName: e.CategoryName,
			Amount:       e.Amount.String(),
			Percentage:   percentage.String(),
		})
	}

	return responses, nil
}

func resolveYearMonth(params web.DashboardQueryParams) (int, int) {
	year := params.Year
	month := params.Month

	if year == 0 || month == 0 {
		now := timeNow()
		if year == 0 {
			year = now.Year()
		}
		if month == 0 {
			month = int(now.Month())
		}
	}

	return year, month
}