package handlers

import (
	"net/http"

	"github.com/emper0r/val-store/server/internal/api/middleware"
	"github.com/emper0r/val-store/server/internal/models"
	"github.com/emper0r/val-store/server/internal/services"
	"github.com/gin-gonic/gin"
)

// UserHandler 处理用户相关请求
type UserHandler struct {
	userService *services.UserService
	shopService *services.ShopService
}

// NewUserHandler 创建新的用户处理器
func NewUserHandler(userService *services.UserService, shopService *services.ShopService) *UserHandler {
	return &UserHandler{
		userService: userService,
		shopService: shopService,
	}
}

// GetUserInfo 获取用户基本信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	// 从上下文中获取用户ID和用户名
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	if userID == "" || username == "" {
		c.JSON(http.StatusBadRequest, models.APIError{
			Status:  http.StatusBadRequest,
			Message: "无效的用户ID或用户名",
		})
		return
	}

	// 获取用户信息
	userInfo := h.userService.GetUserInfo(userID, username)

	// 返回用户信息
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "成功获取用户信息",
		Data:    userInfo,
	})
}

// GetUserWallet 获取用户钱包/余额信息
func (h *UserHandler) GetUserWallet(c *gin.Context) {
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

	// 调用用户服务获取钱包数据
	walletData, err := h.userService.GetUserWallet(userID, session.AccessToken, session.Entitlement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Status:  http.StatusInternalServerError,
			Message: "获取钱包数据失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回钱包数据
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "成功获取钱包数据",
		Data:    walletData,
	})
}

// SetUserRegion 设置用户游戏区域
func (h *UserHandler) SetUserRegion(c *gin.Context) {
	// 从上下文中获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIError{
			Status:  http.StatusBadRequest,
			Message: "无效的用户ID",
		})
		return
	}

	// 解析请求体
	var req models.RegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{
			Status:  http.StatusBadRequest,
			Message: "无效的请求参数",
			Error:   err.Error(),
		})
		return
	}

	// 验证区域有效性
	if err := h.userService.SetUserRegion(req.Region); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{
			Status:  http.StatusBadRequest,
			Message: "设置区域失败",
			Error:   err.Error(),
		})
		return
	}

	// 更新用户会话中的区域设置
	if err := h.shopService.UpdateUserRegion(userID, req.Region); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Status:  http.StatusInternalServerError,
			Message: "更新用户区域失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "成功设置区域",
		Data: map[string]string{
			"region": req.Region,
		},
	})
}

// GetSupportedRegions 获取支持的游戏区域列表
func (h *UserHandler) GetSupportedRegions(c *gin.Context) {
	// 获取支持的区域列表
	regions := h.userService.GetSupportedRegions()

	// 返回区域列表
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "成功获取支持的区域列表",
		Data: map[string]interface{}{
			"regions": regions,
		},
	})
}

// RegisterRoutes 注册用户相关路由
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	protected := router.Group("")
	protected.Use(authMiddleware)

	protected.GET("/user/info", h.GetUserInfo)
	protected.GET("/user/wallet", h.GetUserWallet)
	protected.POST("/user/region", h.SetUserRegion)

	// 区域列表可以不需要认证
	router.GET("/regions", h.GetSupportedRegions)
}
