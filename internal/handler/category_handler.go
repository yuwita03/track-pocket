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

type CategoryHandler struct {
	Service  service.CategoryService
	Validate *validator.Validate
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		Service:  service,
		Validate: validator.New(),
	}
}

func getUserID(ctx *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		return uuid.Nil, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "user not found in context")
	}
	return uuid.Parse(userIDStr.(string))
}

func (handler *CategoryHandler) Create(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	var req web.CreateCategoryRequest
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

func (handler *CategoryHandler) FindAll(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	result, err := handler.Service.FindAll(ctx.Request.Context(), userID)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}

func (handler *CategoryHandler) Update(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	categoryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_ID", "invalid category id"))
		return
	}

	var req web.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	result, err := handler.Service.Update(ctx.Request.Context(), userID, categoryID, req)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}

func (handler *CategoryHandler) Delete(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid user"))
		return
	}

	categoryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_ID", "invalid category id"))
		return
	}

	if err := handler.Service.Delete(ctx.Request.Context(), userID, categoryID); err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, gin.H{"message": "category deleted"})
}