package services

import (
	"fmt"
	"time"

	"github.com/emper0r/val-store/server/internal/models"
	"github.com/emper0r/val-store/server/internal/repositories"
)

// ShopService 处理商店相关的业务逻辑
type ShopService struct {
	valorantAPI  *repositories.ValorantAPI
	skinDatabase *repositories.SkinDatabase
	sessionCache map[string]*models.UserSession // 缓存用户会话数据（UserID -> UserSession）
}

// NewShopService 创建新的商店服务
func NewShopService(valorantAPI *repositories.ValorantAPI, skinDatabase *repositories.SkinDatabase) *ShopService {
	return &ShopService{
		valorantAPI:  valorantAPI,
		skinDatabase: skinDatabase,
		sessionCache: make(map[string]*models.UserSession),
	}
}

// GetShop 获取用户的商店数据
func (s *ShopService) GetShop(userID, accessToken, entitlementToken string) (*models.ShopResponse, error) {
	// 从会话缓存中获取用户区域
	session, exists := s.sessionCache[userID]
	if exists && session.Region != "" {
		// 设置用户特定的区域
		fmt.Printf("从会话缓存中获取用户区域: %s\n", session.Region)
		s.valorantAPI.SetRegion(session.Region)
	} else {
		fmt.Printf("未找到用户区域设置，使用默认区域\n")
	}

	// 调用 Valorant API 获取原始商店数据
	storeData, err := s.valorantAPI.GetStoreOffers(userID, accessToken, entitlementToken)
	if err != nil {
		return nil, fmt.Errorf("获取商店数据失败: %w", err)
	}

	// 创建商店响应
	shopResponse := &models.ShopResponse{
		DailyOffers: make([]models.ShopItem, 0, len(storeData.SkinsPanelLayout.SingleItemOffers)),
		ExpiresAt:   time.Now().Unix() + storeData.SkinsPanelLayout.SingleItemOffersRemainingDurationInSeconds,
	}

	// 处理每日商店
	for _, skinID := range storeData.SkinsPanelLayout.SingleItemOffers {
		// 从皮肤数据库中获取皮肤详细信息
		skinInfo, found := s.skinDatabase.GetSkinByID(skinID)
		if !found {
			// 如果皮肤未找到，使用占位符
			skinInfo = models.Skin{
				UUID:       skinID,
				Name:       "未知皮肤",
				IconURL:    "",
				TierUUID:   "",
				TierName:   "未知",
				Price:      0,
				WeaponID:   "",
				WeaponName: "未知武器",
			}
		}

		// 添加到每日商店
		shopResponse.DailyOffers = append(shopResponse.DailyOffers, models.ShopItem{
			Skin:       skinInfo,
			FinalPrice: skinInfo.Price, // 使用标准价格
		})
	}

	// 处理精选套装
	if len(storeData.FeaturedBundle.Bundles) > 0 {
		bundle := storeData.FeaturedBundle.Bundles[0]

		// 设置精选套装信息
		shopResponse.FeaturedBundle.Name = bundle.DataAssetID
		shopResponse.FeaturedBundle.Price = 0           // 将在循环中累加
		shopResponse.FeaturedBundle.DiscountedPrice = 0 // 将在循环中累加
		shopResponse.FeaturedBundle.ExpiresAt = time.Now().Unix() + bundle.DurationRemainingInSeconds
		shopResponse.FeaturedBundle.Items = make([]models.ShopItem, 0, len(bundle.Items))

		// 处理套装内的物品
		for _, item := range bundle.Items {
			if item.Item.ItemTypeID == "e7c63390-eda7-46e0-bb7a-a6abdacd2433" { // 检查是否为皮肤
				skinID := item.Item.ItemID
				skinInfo, found := s.skinDatabase.GetSkinByID(skinID)

				if !found {
					// 如果皮肤未找到，使用占位符
					skinInfo = models.Skin{
						UUID:       skinID,
						Name:       "未知皮肤",
						IconURL:    "",
						TierUUID:   "",
						TierName:   "未知",
						Price:      item.BasePrice,
						WeaponID:   "",
						WeaponName: "未知武器",
					}
				}

				// 添加套装物品
				shopItem := models.ShopItem{
					Skin:            skinInfo,
					DiscountPercent: item.DiscountPercent,
					FinalPrice:      item.DiscountedPrice,
				}

				shopResponse.FeaturedBundle.Items = append(shopResponse.FeaturedBundle.Items, shopItem)
				shopResponse.FeaturedBundle.Price += item.BasePrice
				shopResponse.FeaturedBundle.DiscountedPrice += item.DiscountedPrice
			}
		}
	}

	// 处理特惠商店（夜市）如果存在
	if len(storeData.BonusStore.BonusStoreOffers) > 0 {
		shopResponse.BonusOffers = make([]models.ShopItem, 0, len(storeData.BonusStore.BonusStoreOffers))

		for _, offer := range storeData.BonusStore.BonusStoreOffers {
			skinID := offer.Offer.ItemID
			skinInfo, found := s.skinDatabase.GetSkinByID(skinID)

			if !found {
				// 如果皮肤未找到，使用占位符
				skinInfo = models.Skin{
					UUID:       skinID,
					Name:       "未知皮肤",
					IconURL:    "",
					TierUUID:   "",
					TierName:   "未知",
					Price:      0,
					WeaponID:   "",
					WeaponName: "未知武器",
				}
			}

			// 获取折扣价格
			var finalPrice int
			for _, price := range offer.DiscountCosts {
				finalPrice = price
				break
			}

			// 添加特惠物品
			shopResponse.BonusOffers = append(shopResponse.BonusOffers, models.ShopItem{
				Skin:            skinInfo,
				DiscountPercent: offer.DiscountPercent,
				FinalPrice:      finalPrice,
			})
		}
	}

	return shopResponse, nil
}

// CacheUserSession 缓存用户会话
func (s *ShopService) CacheUserSession(userID string, session *models.UserSession) {
	s.sessionCache[userID] = session
}

// GetCachedSession 获取缓存的用户会话
func (s *ShopService) GetCachedSession(userID string) (*models.UserSession, bool) {
	session, exists := s.sessionCache[userID]
	return session, exists
}

// UpdateUserRegion 更新用户会话中的区域设置
func (s *ShopService) UpdateUserRegion(userID, region string) error {
	session, exists := s.sessionCache[userID]
	if !exists {
		return fmt.Errorf("用户会话不存在，请先登录")
	}

	// 更新会话中的区域设置
	session.Region = region
	s.sessionCache[userID] = session

	// 同时更新API客户端的区域设置
	s.valorantAPI.SetRegion(region)

	fmt.Printf("已更新用户 %s 的区域设置为 %s\n", userID, region)
	return nil
}
