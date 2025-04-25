package models

import (
	"github.com/golang-jwt/jwt/v5"
)

// UserCredentials 用户身份凭证
type UserCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CookieLoginRequest Cookie登录请求
type CookieLoginRequest struct {
	Cookies    string `json:"cookies" binding:"required"`
	RememberMe bool   `json:"remember_me"`
	Region     string `json:"region"` // 可选的区域设置参数
}

// UserSession 用户会话信息
type UserSession struct {
	UserID       string            `json:"user_id"`
	Username     string            `json:"username"`
	AccessToken  string            `json:"access_token"`
	Entitlement  string            `json:"entitlement_token"`
	RiotUsername string            `json:"riot_username"`
	RiotTagline  string            `json:"riot_tagline"`
	Region       string            `json:"region"` // 用户区域，如ap、na、eu等
	Cookies      map[string]string `json:"-"`      // Cookie不会返回给客户端
}

// JWTClaims 定义JWT令牌的声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// UserTokensResponse 登录成功后的响应
type UserTokensResponse struct {
	Token string `json:"token"` // JWT令牌
	User  struct {
		Username string `json:"username"`
		UserID   string `json:"user_id"`
	} `json:"user"`
}

// APIError 统一API错误响应格式
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// APISuccess 统一API成功响应格式
type APISuccess struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RegionRequest 设置用户区域的请求
type RegionRequest struct {
	Region string `json:"region" binding:"required"`
}

// 支持的区域常量
const (
	RegionAP    = "ap"    // 亚太地区
	RegionNA    = "na"    // 北美
	RegionEU    = "eu"    // 欧洲
	RegionKR    = "kr"    // 韩国
	RegionLATAM = "latam" // 拉丁美洲
	RegionBR    = "br"    // 巴西
)
