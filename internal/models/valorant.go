package models

// ValorantAuthResponse Riot认证服务返回的响应
type ValorantAuthResponse struct {
	Type     string `json:"type"`
	Response struct {
		Parameters struct {
			URI string `json:"uri"`
		} `json:"parameters"`
	} `json:"response"`
}

// ValorantTokenResponse 包含Riot的访问令牌
type ValorantTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	SessionState string `json:"session_state"`
}

// ValorantEntitlementResponse 包含Valorant的授权令牌
type ValorantEntitlementResponse struct {
	EntitlementToken string `json:"entitlements_token"`
}

// ValorantUserInfoResponse 包含用户ID和其他信息
type ValorantUserInfoResponse struct {
	Sub      string `json:"sub"` // 用户ID
	Email    string `json:"email"`
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Picture  string `json:"picture,omitempty"`
	Country  string `json:"country,omitempty"`
	Locale   string `json:"locale,omitempty"`
	PhoneID  string `json:"phone_id,omitempty"`
	Verified bool   `json:"email_verified"`
	// 添加acct字段，包含游戏名称和标签
	Acct struct {
		GameName string `json:"game_name"`
		TagLine  string `json:"tag_line"`
	} `json:"acct"`
}

// ValorantStoreResponse 包含商店中的皮肤信息
type ValorantStoreResponse struct {
	FeaturedBundle       FeaturedBundle       `json:"FeaturedBundle"`
	SkinsPanelLayout     SkinsPanelLayout     `json:"SkinsPanelLayout"`
	BonusStore           BonusStore           `json:"BonusStore,omitempty"`
	AccessoryStore       AccessoryStore       `json:"AccessoryStore"`
	UpgradeCurrencyStore UpgradeCurrencyStore `json:"UpgradeCurrencyStore"`
}

// FeaturedBundle 精选套装信息
type FeaturedBundle struct {
	Bundle                           Bundle   `json:"Bundle"`
	Bundles                          []Bundle `json:"Bundles"`
	BundleRemainingDurationInSeconds int64    `json:"BundleRemainingDurationInSeconds"`
}

// Bundle 套装信息
type Bundle struct {
	ID                         string       `json:"ID"`
	DataAssetID                string       `json:"DataAssetID"`
	CurrencyID                 string       `json:"CurrencyID"`
	Items                      []BundleItem `json:"Items"`
	DurationRemainingInSeconds int64        `json:"DurationRemainingInSeconds"`
	WholesaleOnly              bool         `json:"WholesaleOnly"`
}

// BundleItem 套装中的单个物品信息
type BundleItem struct {
	Item            ItemInfo `json:"Item"`
	BasePrice       int      `json:"BasePrice"`
	DiscountPercent int      `json:"DiscountPercent"`
	DiscountedPrice int      `json:"DiscountedPrice"`
	Quantity        int      `json:"Quantity"`
}

// ItemInfo 物品的详细信息
type ItemInfo struct {
	ItemTypeID string `json:"ItemTypeID"`
	ItemID     string `json:"ItemID"`
	Amount     int    `json:"Amount"`
}

// SkinsPanelLayout 每日商店皮肤展示区布局
type SkinsPanelLayout struct {
	SingleItemOffers                           []string `json:"SingleItemOffers"` // 皮肤ID列表
	SingleItemOffersRemainingDurationInSeconds int64    `json:"SingleItemOffersRemainingDurationInSeconds"`
}

// BonusStore 额外商店/特惠商店（如有夜市）
type BonusStore struct {
	BonusStoreOffers                     []BonusStoreOffer `json:"BonusStoreOffers"`
	BonusStoreRemainingDurationInSeconds int64             `json:"BonusStoreRemainingDurationInSeconds"`
}

// BonusStoreOffer 特惠商店中的物品
type BonusStoreOffer struct {
	BonusOfferID    string         `json:"BonusOfferID"`
	Offer           ItemInfo       `json:"Offer"`
	DiscountPercent int            `json:"DiscountPercent"`
	DiscountCosts   map[string]int `json:"DiscountCosts"`
	IsSeen          bool           `json:"IsSeen"`
}

// AccessoryStore 配件商店
type AccessoryStore struct {
	AccessoryStoreOffers                     []AccessoryStoreOffer `json:"AccessoryStoreOffers"`
	AccessoryStoreRemainingDurationInSeconds int64                 `json:"AccessoryStoreRemainingDurationInSeconds"`
}

// AccessoryStoreOffer 配件商店中的物品
type AccessoryStoreOffer struct {
	Offer      ItemInfo `json:"Offer"`
	ContractID string   `json:"ContractID"`
}

// UpgradeCurrencyStore 升级币商店
type UpgradeCurrencyStore struct {
	UpgradeCurrencyOffers []UpgradeCurrencyOffer `json:"UpgradeCurrencyOffers"`
}

// UpgradeCurrencyOffer 升级币商店中的物品
type UpgradeCurrencyOffer struct {
	OfferID          string   `json:"OfferID"`
	StorefrontItemID string   `json:"StorefrontItemID"`
	Offer            ItemInfo `json:"Offer"`
	Cost             int      `json:"Cost"`
}

// ValorantWalletResponse 用户钱包/余额信息
type ValorantWalletResponse struct {
	Balances map[string]int `json:"Balances"` // 键是货币ID，值是数量
}

// Skin 皮肤信息
type Skin struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	IconURL    string `json:"icon_url"`
	TierUUID   string `json:"tier_uuid"`
	TierName   string `json:"tier_name"`
	Price      int    `json:"price"`
	WeaponID   string `json:"weapon_id"`
	WeaponName string `json:"weapon_name"`
}

// SkinsDatabase 所有皮肤的本地数据库
type SkinsDatabase struct {
	Skins []Skin `json:"skins"`
}

// ShopItem 商店物品，包含完整的皮肤信息
type ShopItem struct {
	Skin            Skin `json:"skin"`
	DiscountPercent int  `json:"discount_percent,omitempty"`
	FinalPrice      int  `json:"final_price"`
}

// ShopResponse 客户端商店响应
type ShopResponse struct {
	DailyOffers    []ShopItem `json:"daily_offers"`
	FeaturedBundle struct {
		Name            string     `json:"name"`
		Description     string     `json:"description"`
		Items           []ShopItem `json:"items"`
		Price           int        `json:"price"`
		DiscountedPrice int        `json:"discounted_price"`
		ExpiresAt       int64      `json:"expires_at"` // Unix时间戳
	} `json:"featured_bundle"`
	BonusOffers []ShopItem `json:"bonus_offers,omitempty"` // 夜市
	ExpiresAt   int64      `json:"expires_at"`             // Unix时间戳
}

// WalletResponse 客户端钱包/余额响应
type WalletResponse struct {
	ValorantPoints  int `json:"valorant_points"`
	RadianitePoints int `json:"radianite_points"`
	KingdomCredits  int `json:"kingdom_credits,omitempty"`
}
