package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadConfig 加载.env文件中的配置
func LoadConfig() error {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		log.Println("无法获取工作目录:", err)
		return err
	}

	// 首先尝试从当前目录加载.env文件
	err = godotenv.Load()
	if err != nil {
		// 如果失败，尝试从项目根目录加载
		rootPath := filepath.Join(wd, "../../.env")
		err = godotenv.Load(rootPath)
		if err != nil {
			log.Println("警告: .env文件未找到，将使用环境变量或默认值")
		}
	}

	return nil
}

// GetEnv 获取环境变量值，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
