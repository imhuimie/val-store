package handlers

import (
	"net/http"

	"github.com/emper0r/val-store/server/internal/api/middleware"
	"github.com/emper0r/val-store/server/internal/models"
	"github.com/emper0r/val-store/server/internal/services"
	"github.com/gin-gonic/gin"
)

// ShopHandler 处理商店相关请求
type ShopHandler struct {
	shopService *services.ShopService
}

// NewShopHandler 创建新的商店处理器
func NewShopHandler(shopService *services.ShopService) *ShopHandler {
	return &ShopHandler{
		shopService: shopService,
	}
}

// GetShop 获取用户的每日商店数据
func (h *ShopHandler) GetShop(c *gin.Context) {
	// 从上下文中获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIError{
			Status:  http.StatusBadRequest,
			Message: "无效的用户ID",
		})
		return
	}

	// 从缓存中获取用户会话数据
	session, exists := h.shopService.GetCachedSession(userID)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Status:  http.StatusUnauthorized,
			Message: "会话已过期",
			Error:   "请重新登录以刷新会话",
		})
		return
	}

	// 调用商店服务获取商店数据
	shopData, err := h.shopService.GetShop(userID, session.AccessToken, session.Entitlement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Status:  http.StatusInternalServerError,
			Message: "获取商店数据失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回商店数据
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "成功获取商店数据",
		Data:    shopData,
	})
}

// RegisterRoutes 注册商店相关路由
func (h *ShopHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	protected := router.Group("")
	protected.Use(authMiddleware)

	protected.GET("/shop", h.GetShop)
}
