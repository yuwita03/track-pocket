package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"trackpocket/internal/email"
	"trackpocket/internal/handler"
	"trackpocket/internal/middleware"
	"trackpocket/internal/repository"
	"trackpocket/internal/service"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("failed to parse database config: %v", err)
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	tokenExpiry := 15 * time.Minute

	// 1. Repositories
	userRepo := repository.NewUserRepository(dbPool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(dbPool)
	passwordResetRepo := repository.NewPasswordResetTokenRepository(dbPool)
	emailVerificationRepo := repository.NewEmailVerificationTokenRepository(dbPool)
	categoryRepo := repository.NewCategoryRepository(dbPool)
	transactionRepo := repository.NewTransactionRepository(dbPool)

	// 2. External dependencies
	emailSender := email.NewLogSender()

	// 3. Services
	categoryService := service.NewCategoryService(categoryRepo)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, passwordResetRepo, emailVerificationRepo, categoryService, emailSender, jwtSecret, tokenExpiry)
	transactionService := service.NewTransactionService(transactionRepo, categoryRepo)

	// 4. Handlers
	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	transactionHandler := handler.NewTransactionHandler(transactionService)

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
		auth.POST("/reset-password", authHandler.ResetPassword)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.POST("/resend-verification", authHandler.ResendVerification)
		auth.GET("/me", middleware.JWTAuth(jwtSecret), authHandler.Me)
	}

	categories := router.Group("/api/v1/categories")
	categories.Use(middleware.JWTAuth(jwtSecret))
	{
		categories.POST("", categoryHandler.Create)
		categories.GET("", categoryHandler.FindAll)
		categories.PATCH("/:id", categoryHandler.Update)
		categories.DELETE("/:id", categoryHandler.Delete)
	}

	transactions := router.Group("/api/v1/transactions")
	transactions.Use(middleware.JWTAuth(jwtSecret))
	{
		transactions.POST("", transactionHandler.Create)
		transactions.GET("", transactionHandler.FindAll)
		transactions.PATCH("/:id", transactionHandler.Update)
		transactions.DELETE("/:id", transactionHandler.Delete)
	}
	// Repositories — tambah 1 baris
	dashboardRepo := repository.NewDashboardRepository(dbPool)
	// Services — tambah 1 baris
	dashboardService := service.NewDashboardService(dashboardRepo)
	// Handlers — tambah 1 baris
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	// Routes — tambah group baru
	dashboard := router.Group("/api/v1/dashboard")
	dashboard.Use(middleware.JWTAuth(jwtSecret))
	{
		dashboard.GET("/summary", dashboardHandler.Summary)
		dashboard.GET("/category-expenses", dashboardHandler.CategoryExpenses)
	}

	router.Run(":8080")
}
