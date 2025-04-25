package services

import (
	"fmt"

	"github.com/emper0r/val-store/server/internal/models"
	"github.com/emper0r/val-store/server/internal/repositories"
)

// UserService 处理用户相关的业务逻辑
type UserService struct {
	valorantAPI *repositories.ValorantAPI
}

// NewUserService 创建新的用户服务
func NewUserService(valorantAPI *repositories.ValorantAPI) *UserService {
	return &UserService{
		valorantAPI: valorantAPI,
	}
}

// GetUserWallet 获取用户钱包/余额信息
func (s *UserService) GetUserWallet(userID, accessToken, entitlementToken string) (*models.WalletResponse, error) {
	// 调用 Valorant API 获取用户钱包数据
	walletData, err := s.valorantAPI.GetWallet(userID, accessToken, entitlementToken)
	if err != nil {
		return nil, fmt.Errorf("获取用户钱包数据失败: %w", err)
	}

	// 创建钱包响应
	walletResponse := &models.WalletResponse{
		ValorantPoints:  0,
		RadianitePoints: 0,
		KingdomCredits:  0,
	}

	// 映射货币ID到对应的货币类型
	const (
		ValPointsCurrencyID       = "85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741" // VP
		RadianitePointsCurrencyID = "e59aa87c-4cbf-517a-5983-6e81511be9b7" // RP
		KingdomCreditsCurrencyID  = "85ca954a-41f2-ce94-9b45-8ca3dd39a00d" // KC
	)

	// 提取各种货币的余额
	for currencyID, amount := range walletData.Balances {
		switch currencyID {
		case ValPointsCurrencyID:
			walletResponse.ValorantPoints = amount
		case RadianitePointsCurrencyID:
			walletResponse.RadianitePoints = amount
		case KingdomCreditsCurrencyID:
			walletResponse.KingdomCredits = amount
		}
	}

	return walletResponse, nil
}

// GetUserInfo 获取用户基本信息（如等级、地区等）
// 注：这个方法可以扩展，目前只返回可从JWT中获取的信息
func (s *UserService) GetUserInfo(userID, username string) map[string]interface{} {
	return map[string]interface{}{
		"user_id":  userID,
		"username": username,
	}
}

// SetUserRegion 设置用户的游戏区域
func (s *UserService) SetUserRegion(region string) error {
	// 验证区域是否有效
	validRegions := []string{
		models.RegionAP,
		models.RegionNA,
		models.RegionEU,
		models.RegionKR,
		models.RegionLATAM,
		models.RegionBR,
	}

	isValid := false
	for _, validRegion := range validRegions {
		if region == validRegion {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("无效的区域代码: %s", region)
	}

	// 设置Valorant API客户端的区域
	s.valorantAPI.SetRegion(region)
	fmt.Printf("用户手动设置区域为: %s\n", region)

	return nil
}

// GetSupportedRegions 获取支持的游戏区域列表
func (s *UserService) GetSupportedRegions() []map[string]string {
	regions := []map[string]string{
		{"code": models.RegionAP, "name": "亚太地区 (Asia Pacific)", "default": "true"},
		{"code": models.RegionNA, "name": "北美 (North America)", "default": "false"},
		{"code": models.RegionEU, "name": "欧洲 (Europe)", "default": "false"},
		{"code": models.RegionKR, "name": "韩国 (Korea)", "default": "false"},
		{"code": models.RegionLATAM, "name": "拉丁美洲 (Latin America)", "default": "false"},
		{"code": models.RegionBR, "name": "巴西 (Brazil)", "default": "false"},
	}

	return regions
}
