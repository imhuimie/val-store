package handlers

import (
	"net/http"

	"github.com/emper0r/val-store/server/internal/models"
	"github.com/emper0r/val-store/server/internal/services"
	"github.com/gin-gonic/gin"
)

// SkinsHandler 处理皮肤相关请求
type SkinsHandler struct {
	skinsService *services.SkinsService
}

// NewSkinsHandler 创建新的皮肤处理器
func NewSkinsHandler(skinsService *services.SkinsService) *SkinsHandler {
	return &SkinsHandler{
		skinsService: skinsService,
	}
}

// GetAllSkins 获取所有皮肤列表
func (h *SkinsHandler) GetAllSkins(c *gin.Context) {
	// 调用皮肤服务获取所有皮肤
	skins, err := h.skinsService.GetAllSkins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Status:  http.StatusInternalServerError,
			Message: "获取皮肤列表失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回皮肤列表
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "成功获取皮肤列表",
		Data:    skins,
	})
}

// GetSkinByID 根据ID获取皮肤信息
func (h *SkinsHandler) GetSkinByID(c *gin.Context) {
	// 获取URL参数中的皮肤ID
	skinID := c.Param("id")
	if skinID == "" {
		c.JSON(http.StatusBadRequest, models.APIError{
			Status:  http.StatusBadRequest,
			Message: "缺少皮肤ID",
		})
		return
	}

	// 调用皮肤服务获取皮肤信息
	skin, err := h.skinsService.GetSkinByID(skinID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIError{
			Status:  http.StatusNotFound,
			Message: "皮肤未找到",
			Error:   err.Error(),
		})
		return
	}

	// 返回皮肤信息
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "成功获取皮肤信息",
		Data:    skin,
	})
}

// RegisterRoutes 注册皮肤相关路由
func (h *SkinsHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/skins", h.GetAllSkins)
	router.GET("/skins/:id", h.GetSkinByID)
}
