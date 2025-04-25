package main

import (
	"log"
	"net/http"
	"time"

	"github.com/emper0r/val-store/server/internal/api"
	"github.com/emper0r/val-store/server/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载环境变量配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 设置Gin模式（默认为debug模式）
	ginMode := config.GetEnv("GIN_MODE", "debug")
	gin.SetMode(ginMode)

	// 创建Gin引擎
	app := gin.Default()

	// 初始化路由
	api.SetupRouter(app)

	// 获取端口配置
	port := config.GetEnv("PORT", "8080")

	// 设置超时
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      app,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// 启动服务器
	log.Printf("服务器正在端口 %s 上启动，环境：%s", port, ginMode)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
