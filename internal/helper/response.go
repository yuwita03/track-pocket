package helper

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"trackpocket/internal/exception"
)

type errorBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}	

func Error(c *gin.Context, err error) {
	var appErr *exception.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, errorBody{
			Success: false,
			Message: appErr.Message,
			Code:    appErr.Code,
		})
		return
	}

	// error gak dikenal -> fallback 500, jangan bocorin detail internal
	c.JSON(http.StatusInternalServerError, errorBody{
		Success: false,
		Message: "Internal server error",
		Code:    "INTERNAL_ERROR",
	})
}