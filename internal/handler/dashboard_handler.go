package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"trackpocket/internal/exception"
	"trackpocket/internal/helper"
	"trackpocket/internal/model/web"
	"trackpocket/internal/service"
)

type DashboardHandler struct {
	Service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{Service: service}
}

func (handler *DashboardHandler) Summary(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	var params web.DashboardQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid query parameters"))
		return
	}

	result, err := handler.Service.GetSummary(ctx.Request.Context(), userID, params)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}

func (handler *DashboardHandler) CategoryExpenses(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	var params web.DashboardQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid query parameters"))
		return
	}

	result, err := handler.Service.GetCategoryExpenses(ctx.Request.Context(), userID, params)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}