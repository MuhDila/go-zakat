package middleware

import (
	"net/http"
	"strings"

	"go-zakat-be/internal/domain/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware menyimpan dependencies untuk validasi JWT
type AuthMiddleware struct {
	tokenSvc service.TokenService
}

func NewAuthMiddleware(tokenSvc service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokenSvc: tokenSvc}
}

// RequireAuth adalah middleware yang mengecek Authorization: Bearer <token>
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header kosong",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Format Authorization harus: Bearer <token>",
			})
			return
		}

		tokenStr := parts[1]
		userID, role, err := m.tokenSvc.ValidateAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "token tidak valid atau expired",
			})
			return
		}

		// Simpan userID dan role ke context supaya handler bisa pakai
		c.Set("user_id", userID)
		c.Set("user_role", role)

		c.Next()
	}
}

// RequireRole adalah middleware yang mengecek apakah user punya salah satu dari roles yang diizinkan
func (m *AuthMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Role tidak ditemukan",
			})
			return
		}

		userRole := role.(string)
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "Anda tidak memiliki akses ke resource ini",
		})
	}
}

// RequireAdmin adalah shortcut untuk require role admin
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}

// RequireStafOrAdmin adalah shortcut untuk require role staf atau admin
func (m *AuthMiddleware) RequireStafOrAdmin() gin.HandlerFunc {
	return m.RequireRole("staf", "admin")
}
