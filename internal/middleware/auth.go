package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"trackpocket/internal/exception"
	"trackpocket/internal/helper"
)

func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			helper.Error(c, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			helper.Error(c, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			helper.Error(c, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			helper.Error(c, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid token claims"))
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			helper.Error(c, exception.New(http.StatusUnauthorized, "UNAUTHORIZED", "invalid token claims"))
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}