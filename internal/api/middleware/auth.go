package middleware

import (
	"net/http"
	"strings"

	"github.com/emper0r/val-store/server/internal/models"
	"github.com/emper0r/val-store/server/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 创建JWT认证中间件
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization信息
		authHeader := c.GetHeader("Authorization")

		// 检查Authorization头是否存在并符合格式
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIError{
				Status:  http.StatusUnauthorized,
				Message: "未授权",
				Error:   "缺少Authorization头",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, models.APIError{
				Status:  http.StatusUnauthorized,
				Message: "未授权",
				Error:   "无效的Authorization格式，应为Bearer令牌",
			})
			c.Abort()
			return
		}

		// 提取JWT令牌
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证JWT令牌
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIError{
				Status:  http.StatusUnauthorized,
				Message: "未授权",
				Error:   "令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中，以便后续处理程序使用
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetUsername 从上下文中获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}
