package web

type DashboardSummaryResponse struct {
	TotalIncome      string `json:"total_income"`
	TotalExpense     string `json:"total_expense"`
	Balance          string `json:"balance"`
	TransactionCount int    `json:"transaction_count"`
}

type CategoryExpenseResponse struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	Amount       string `json:"amount"`
	Percentage   string `json:"percentage"`
}