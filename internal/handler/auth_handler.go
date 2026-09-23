package handler

import (
	"net/http"
	"trackpocket/internal/exception"
	"trackpocket/internal/helper"
	"trackpocket/internal/service"

	"trackpocket/internal/model/web"

	"github.com/gin-gonic/gin"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	Service  service.AuthService
	Validate *validator.Validate
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{
		Service:  service,
		Validate: validator.New(),
	}
}

func (handler *AuthHandler) Register(ctx *gin.Context) {
	var req web.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body"))
		return
	}

	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	result, err := handler.Service.Register(ctx.Request.Context(), req)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusCreated, result)
}

func (handler *AuthHandler) Login(ctx *gin.Context) {
	var req web.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body"))
		return
	}

	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	result, err := handler.Service.Login(ctx.Request.Context(), req)
	if err != nil {
		helper.Error(ctx, err) // ← FIX: sebelumnya di-wrap ulang jadi INVALID_CREDENTIALS terus
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}

func (handler *AuthHandler) Me(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		helper.Error(ctx, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "user not found in context"))
		return
	}

	result, err := handler.Service.GetProfile(ctx.Request.Context(), userID.(string))
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}

func (handler *AuthHandler) Refresh(ctx *gin.Context) {
	var req web.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	result, err := handler.Service.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, result)
}

func (handler *AuthHandler) Logout(ctx *gin.Context) {
	var req web.LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	if err := handler.Service.Logout(ctx.Request.Context(), req.RefreshToken); err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (handler *AuthHandler) ForgotPassword(ctx *gin.Context) {
	var req web.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	if err := handler.Service.ForgotPassword(ctx.Request.Context(), req.Email); err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, gin.H{"message": "if the email is registered, a reset link has been sent"})
}

func (handler *AuthHandler) ResetPassword(ctx *gin.Context) {
	var req web.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	if err := handler.Service.ResetPassword(ctx.Request.Context(), req.Token, req.NewPassword); err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, gin.H{"message": "password reset successfully"})
}

func (handler *AuthHandler) VerifyEmail(ctx *gin.Context) {
	var req web.VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	if err := handler.Service.VerifyEmail(ctx.Request.Context(), req.Token); err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, gin.H{"message": "email verified successfully"})
}

func (handler *AuthHandler) ResendVerification(ctx *gin.Context) {
	var req web.ResendVerificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.Error(ctx, exception.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		helper.Error(ctx, exception.ErrValidation)
		return
	}

	if err := handler.Service.ResendVerification(ctx.Request.Context(), req.Email); err != nil {
		helper.Error(ctx, err)
		return
	}

	helper.Success(ctx, http.StatusOK, gin.H{"message": "if the email is registered and unverified, a verification link has been sent"})
}