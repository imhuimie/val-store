package api

import (
	"net/http"
	"strings"

	"github.com/emper0r/val-store/server/internal/api/handlers"
	"github.com/emper0r/val-store/server/internal/api/middleware"
	"github.com/emper0r/val-store/server/internal/config"
	"github.com/emper0r/val-store/server/internal/repositories"
	"github.com/emper0r/val-store/server/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置所有API路由
func SetupRouter(router *gin.Engine) *gin.Engine {
	// 配置CORS中间件
	router.Use(corsMiddleware())

	// 初始化存储库
	valorantAPI, err := repositories.NewValorantAPI()
	if err != nil {
		panic(err)
	}

	skinDatabase, err := repositories.NewSkinDatabase("")
	if err != nil {
		panic(err)
	}

	// 初始化服务
	authService := services.NewAuthService(valorantAPI)
	shopService := services.NewShopService(valorantAPI, skinDatabase)
	userService := services.NewUserService(valorantAPI)
	skinsService := services.NewSkinsService(valorantAPI, skinDatabase)

	// 设置AuthService的会话缓存为ShopService
	authService.SetSessionCache(shopService)

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(authService, shopService)
	shopHandler := handlers.NewShopHandler(shopService)
	userHandler := handlers.NewUserHandler(userService, shopService)
	skinsHandler := handlers.NewSkinsHandler(skinsService)

	// 创建身份验证中间件
	authMiddleware := middleware.AuthMiddleware(authService)

	// API路由组
	api := router.Group("/api")
	{
		// 注册各个处理器的路由
		authHandler.RegisterRoutes(api)
		shopHandler.RegisterRoutes(api, authMiddleware)
		userHandler.RegisterRoutes(api, authMiddleware)
		skinsHandler.RegisterRoutes(api)
	}

	return router
}

// corsMiddleware 创建CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取允许的域名列表
		allowedOrigins := config.GetEnv("ALLOWED_ORIGINS", "http://localhost:3000")
		origins := strings.Split(allowedOrigins, ",")

		// 获取请求的Origin
		origin := c.Request.Header.Get("Origin")

		// 检查请求的Origin是否在允许列表中
		allowOrigin := false
		for _, o := range origins {
			if origin == o {
				allowOrigin = true
				break
			}
		}

		// 如果Origin在允许列表中，则设置CORS头
		if allowOrigin {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}

		// 处理OPTIONS请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
