package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keshvan/go-common-forum/jwt"
	"github.com/mitchellh/mapstructure"
)

type AccessClaims struct {
	UserID int64  `mapstructure:"user_id"`
	Role   string `mapstructure:"role"`
	Exp    int64  `mapstructure:"exp"`
	Iat    int64  `mapstructure:"iat"`
}

type AuthMiddleware struct {
	jwt *jwt.JWT
}

func NewAuthMiddleware(jwt *jwt.JWT) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

const (
	ContextUserIDKey = "user_id"
	ContextRoleKey   = "role"
)

func (m *AuthMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		fmt.Println("authHeader", authHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		token := parts[1]
		claims, err := m.jwt.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		var accessClaims AccessClaims
		mapstructure.Decode(claims, &accessClaims)
		fmt.Println(time.Now().Unix(), accessClaims.Exp)
		c.Set(ContextUserIDKey, accessClaims.UserID)
		c.Set(ContextRoleKey, accessClaims.Role)

		c.Next()
	}
}

func (m *AuthMiddleware) ChatAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.jwt.ParseToken(token)
		if err != nil {
			c.Next()
			return
		}

		var accessClaims AccessClaims
		mapstructure.Decode(claims, &accessClaims)
		c.Set(ContextUserIDKey, accessClaims.UserID)
		c.Set(ContextRoleKey, accessClaims.Role)

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := GetRoleFromContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			return
		}

		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}
	return userID.(int64), true
}

func GetRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextRoleKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}
