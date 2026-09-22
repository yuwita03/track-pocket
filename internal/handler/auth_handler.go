package handler

import (
	"net/http"
	"trackpocket/internal/service"

	"github.com/gin-gonic/gin"
	"trackpocket/internal/model/web"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	Service 	service.AuthService
	Validate	*validator.Validate
}

func NewAuthHandler(service service.AuthService) *AuthHandler{
	return &AuthHandler{
		Service: service, 
		Validate: validator.New(),
	}
}

func (handler *AuthHandler) Register(ctx *gin.Context) {
	var req web.RegisterRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = handler.Validate.Struct(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return 
	}

	result, err := handler.Service.Register(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return 
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":   http.StatusCreated,
		"status": "CREATED",
		"data":   result,
	})
}

func (handler *AuthHandler) Login(ctx *gin.Context){
	var req web.LoginRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = handler.Validate.Struct(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := handler.Service.Login(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":   http.StatusCreated,
		"status": "CREATED",
		"data":   result,
	})
}