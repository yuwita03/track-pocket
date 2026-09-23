package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"trackpocket/internal/email"
	"trackpocket/internal/handler"
	"trackpocket/internal/middleware"
	"trackpocket/internal/repository"
	"trackpocket/internal/service"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	tokenExpiry := 15 * time.Minute

	// 1. Siapin semua repository dulu
	userRepo := repository.NewUserRepository(dbPool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(dbPool)
	passwordResetRepo := repository.NewPasswordResetTokenRepository(dbPool)
	emailVerificationRepo := repository.NewEmailVerificationTokenRepository(dbPool)

	// 2. Siapin dependency lain (non-DB)
	emailSender := email.NewLogSender()

	authService := service.NewAuthService(userRepo, refreshTokenRepo, passwordResetRepo, emailVerificationRepo, emailSender, jwtSecret, tokenExpiry)

	// 4. Baru bikin handler dari service yang udah lengkap
	authHandler := handler.NewAuthHandler(authService)

	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.GET("/me", middleware.JWTAuth(jwtSecret), authHandler.Me)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.POST("/resend-verification", authHandler.ResendVerification)
	}

	router.Run(":8080")
}