package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(db *sql.DB)gin.HandlerFunc {
	return func(ctx *gin.Context){
		if err := db.Ping(); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"Status": "DOWN",
				"Database":"Unreachable",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"Status":"OK",
			"Database": "connected",
		})
	}
}