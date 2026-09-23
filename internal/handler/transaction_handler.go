package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"trackpocket/internal/exception"
	"trackpocket/internal/helper"
	"trackpocket/internal/model/web"
	"trackpocket/internal/service"
)

type TransactionHandler struct {
	Service  service.TransactionService
	Validate *validator.Validate
}

func NewTransactionHandler(service service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		Service:  service,
		Validate: validator.New(),
	}
}

func (handler *TransactionHandler) Create(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	var req web.CreateTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	result, err := handler.Service.Create(ctx.Request.Context(), userID, req)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusCreated, result)
}

func (handler *TransactionHandler) FindAll(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	var params web.TransactionQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid query parameters"))
		return
	}

	result, err := handler.Service.FindAll(ctx.Request.Context(), userID, params)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}

func (handler *TransactionHandler) Update(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	transactionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_ID", "invalid transaction id"))
		return
	}

	var req web.UpdateTransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	result, err := handler.Service.Update(ctx.Request.Context(), userID, transactionID, req)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}

func (handler *TransactionHandler) Delete(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	transactionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_ID", "invalid transaction id"))
		return
	}

	if err := handler.Service.Delete(ctx.Request.Context(), userID, transactionID); err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, gin.H{"message": "transaction deleted"})
}