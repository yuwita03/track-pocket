package main

import (
	"context"
	"fmt"
	"os"
	"trackpocket/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context){
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	router.Run(":8080")

	fmt.Println("Connected to PostgreSQL suceed!")
}
