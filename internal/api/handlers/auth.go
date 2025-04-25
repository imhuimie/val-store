package handlers

import (
	"net/http"

	"github.com/emper0r/val-store/server/internal/models"
	"github.com/emper0r/val-store/server/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler 处理认证相关请求
type AuthHandler struct {
	authService *services.AuthService
	shopService *services.ShopService
}

// NewAuthHandler 创建新的认证处理器
func NewAuthHandler(authService *services.AuthService, shopService *services.ShopService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		shopService: shopService,
	}
}

// Login 处理用户登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var credentials models.UserCredentials

	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{
			Status:  http.StatusBadRequest,
			Message: "无效的请求数据",
			Error:   err.Error(),
		})
		return
	}

	// 调用认证服务进行登录
	response, err := h.authService.Login(credentials.Username, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Status:  http.StatusUnauthorized,
			Message: "登录失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回登录成功响应
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "登录成功",
		Data:    response,
	})
}

// LoginWithCookies 处理Cookie登录请求
func (h *AuthHandler) LoginWithCookies(c *gin.Context) {
	var request models.CookieLoginRequest

	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{
			Status:  http.StatusBadRequest,
			Message: "无效的请求数据",
			Error:   err.Error(),
		})
		return
	}

	// 调用认证服务进行Cookie登录，传递区域参数
	response, err := h.authService.LoginWithCookies(request.Cookies, request.Region)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Status:  http.StatusUnauthorized,
			Message: "登录失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回登录成功响应
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "登录成功",
		Data:    response,
	})
}

// Ping 简单的健康检查端点
func (h *AuthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "服务正常运行",
	})
}

// RegisterRoutes 注册认证相关路由
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.Login)
	router.POST("/login/cookies", h.LoginWithCookies)
	router.GET("/ping", h.Ping)
}
