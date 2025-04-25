package repositories

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/emper0r/val-store/server/internal/models"
)

const (
	// API URLs
	loginURL         = "https://auth.riotgames.com/api/v1/authorization"
	loginUserPassURL = "https://auth.riotgames.com/api/v1/authorization"
	entitlementsURL  = "https://entitlements.auth.riotgames.com/api/token/v1"
	userInfoURL      = "https://auth.riotgames.com/userinfo"
	storeURL         = "https://pd.%s.a.pvp.net/store/v3/storefront/%s"
	storeURLV2       = "https://pd.%s.a.pvp.net/store/v2/storefront/%s"
	storeURLV1       = "https://pd.%s.a.pvp.net/store/v1/storefront/%s"
	walletURL        = "https://pd.%s.a.pvp.net/store/v1/wallet/%s"
	contentURL       = "https://shared.%s.a.pvp.net/content-service/v3/content"
	versionURL       = "https://valorant-api.com/v1/version"

	// HTTP Headers
	clientPlatform = "ew0KCSJwbGF0Zm9ybVR5cGUiOiAiUEMiLA0KCSJwbGF0Zm9ybU9TIjogIldpbmRvd3MiLA0KCSJwbGF0Zm9ybU9TVmVyc2lvbiI6ICIxMC4wLjE5MDQyLjEuMjU2LjY0Yml0IiwNCgkicGxhdGZvcm1DaGlwc2V0IjogIlVua25vd24iDQp9"

	// 默认的地区
	defaultRegion = "ap"

	// 认证方式
	authTypeCookies  = 1
	authTypeUserPass = 2
)

// ValorantAPI 处理与Valorant API的交互
type ValorantAPI struct {
	client            *http.Client
	region            string
	accessToken       string
	entitlementsToken string
	clientVersion     string
}

// 用于解析版本 API 响应的结构体
type versionResponse struct {
	Data struct {
		RiotClientVersion string `json:"riotClientVersion"`
	} `json:"data"`
}

// fetchLatestClientVersion 从 valorant-api.com 获取最新的客户端版本
func fetchLatestClientVersion(client *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodGet, versionURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建版本请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "val-store-backend")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求版本信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("获取版本信息失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	var versionData versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&versionData); err != nil {
		return "", fmt.Errorf("解析版本信息失败: %w", err)
	}

	if versionData.Data.RiotClientVersion == "" {
		return "", errors.New("从版本信息响应中未找到 riotClientVersion")
	}

	log.Printf("成功获取到最新的客户端版本: %s\n", versionData.Data.RiotClientVersion)
	return versionData.Data.RiotClientVersion, nil
}

// NewValorantAPI 创建一个新的ValorantAPI实例
func NewValorantAPI() (*ValorantAPI, error) {
	// 创建带cookie jar的HTTP客户端，以便能够维护会话
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// TLS密码套件，增强连接稳定性
	tlsCiphers := []uint16{
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	}

	// 创建具有更健壮配置的Transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true, // 启用IPv4/IPv6双栈
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		// 增强TLS配置
		TLSClientConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			CipherSuites: tlsCiphers,
		},
	}

	client := &http.Client{
		Jar:       jar,
		Timeout:   60 * time.Second, // 增加超时时间到60秒
		Transport: transport,
	}

	versionClient := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}

	fallbackClientVersion := "release-10.07-shipping-6-3399868"
	fetchedClientVersion, err := fetchLatestClientVersion(versionClient)
	currentClientVersion := fallbackClientVersion

	if err != nil {
		log.Printf("警告: 无法获取最新的客户端版本: %v。将使用备用版本: %s\n", err, fallbackClientVersion)
	} else {
		currentClientVersion = fetchedClientVersion
	}

	api := &ValorantAPI{
		client:        client,
		region:        defaultRegion,
		clientVersion: currentClientVersion,
	}

	return api, nil
}

// SetRegion 设置用户的地区
func (v *ValorantAPI) SetRegion(region string) {
	if region == "" {
		v.region = defaultRegion
		return
	}

	// 转小写处理区域代码
	region = strings.ToLower(region)

	// 根据API文档更新区域代码映射
	switch region {
	case "na", "latam", "br":
		// latam和br使用na区域
		v.region = "na"
	case "eu":
		v.region = "eu"
	case "ap", "kr":
		// 有些API可能需要kr和ap区分，但这里我们确保使用正确的ap区域代码
		v.region = region
	case "pbe":
		v.region = "pbe"
	default:
		// 对于其他输入，使用默认区域
		v.region = defaultRegion
	}

	fmt.Printf("区域设置已更新: 输入=%s, 最终区域=%s\n", region, v.region)
}

// GetPlayerRegion 获取玩家所在的区域
func (v *ValorantAPI) GetPlayerRegion(accessToken, entitlementToken string) (string, error) {
	// 按照以下区域顺序尝试
	regionsToTry := []string{"na", "eu", "ap", "kr", "latam", "br"}

	fmt.Printf("开始检测玩家区域...\n")

	// 获取用户ID (puuid)
	userInfo, err := v.getUserInfo(accessToken)
	if err != nil {
		return "", fmt.Errorf("获取用户信息失败: %w", err)
	}

	puuid := userInfo.Sub
	if puuid == "" {
		return "", fmt.Errorf("无法获取用户ID")
	}

	// 尝试每个区域
	for _, region := range regionsToTry {
		// 临时设置区域
		originalRegion := v.region
		v.SetRegion(region)

		fmt.Printf("尝试区域: %s (%s)\n", region, v.region)

		// 尝试获取玩家对局历史
		url := fmt.Sprintf("https://pd.%s.a.pvp.net/name-service/v2/players", v.region)

		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(`["`+puuid+`"]`)))
		if err != nil {
			v.region = originalRegion
			continue
		}

		// 暂时保存令牌以便addCommonHeaders使用
		prevAccessToken := v.accessToken
		prevEntitlementToken := v.entitlementsToken

		// 设置当前请求的令牌
		v.accessToken = accessToken
		v.entitlementsToken = entitlementToken

		// 添加通用头信息
		v.addCommonHeaders(req)

		// 恢复原来的令牌
		v.accessToken = prevAccessToken
		v.entitlementsToken = prevEntitlementToken

		resp, err := v.client.Do(req)
		if err != nil {
			v.region = originalRegion
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("找到玩家区域: %s\n", region)
			return region, nil
		}

		// 恢复原始区域
		v.region = originalRegion
	}

	return defaultRegion, fmt.Errorf("无法确定玩家区域，使用默认区域: %s", defaultRegion)
}

// Authenticate 使用用户名和密码进行认证
func (v *ValorantAPI) Authenticate(username, password string) (*models.UserSession, error) {
	// 第一步：获取认证cookie
	if err := v.requestAuth(); err != nil {
		return nil, fmt.Errorf("认证步骤1失败: %w", err)
	}

	// 第二步：使用用户名和密码登录
	authResponse, err := v.requestLogin(username, password)
	if err != nil {
		return nil, fmt.Errorf("认证步骤2失败: %w", err)
	}

	// 解析认证URI，获取访问令牌
	accessToken, err := parseAuthURI(authResponse.Response.Parameters.URI)
	if err != nil {
		return nil, fmt.Errorf("解析访问令牌失败: %w", err)
	}

	// 获取授权令牌
	entitlementToken, err := v.getEntitlementToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取授权令牌失败: %w", err)
	}

	// 获取用户信息
	userInfo, err := v.getUserInfo(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 获取用户区域
	region, err := v.GetPlayerRegion(accessToken, entitlementToken)
	if err != nil {
		fmt.Printf("警告: 获取用户区域失败: %v, 使用默认区域\n", err)
		region = defaultRegion
	} else {
		// 设置API客户端区域
		v.SetRegion(region)
	}

	// 创建用户会话
	session := &models.UserSession{
		UserID:      userInfo.Sub,
		Username:    username,
		AccessToken: accessToken,
		Entitlement: entitlementToken,
		Region:      region,
	}

	// 设置Riot用户名和标签
	if userInfo.Acct.GameName != "" && userInfo.Acct.TagLine != "" {
		session.RiotUsername = userInfo.Acct.GameName
		session.RiotTagline = userInfo.Acct.TagLine
	} else {
		// 兼容旧版API响应格式
		session.RiotUsername = userInfo.Name
		session.RiotTagline = userInfo.Tag
	}

	return session, nil
}

// requestAuth 初始化认证过程
func (v *ValorantAPI) requestAuth() error {
	data := map[string]interface{}{
		"client_id":     "play-valorant-web-prod",
		"nonce":         "1",
		"redirect_uri":  "https://playvalorant.com/opt_in",
		"response_type": "token id_token",
		"scope":         "account openid",
	}

	return v.makeRequest(http.MethodPost, loginURL, data, nil)
}

// requestLogin 使用用户凭证请求登录
func (v *ValorantAPI) requestLogin(username, password string) (*models.ValorantAuthResponse, error) {
	data := map[string]interface{}{
		"type":     "auth",
		"username": username,
		"password": password,
	}

	var resp models.ValorantAuthResponse
	err := v.makeRequest(http.MethodPut, loginUserPassURL, data, &resp)
	if err != nil {
		return nil, err
	}

	// 检查响应类型
	if resp.Type != "response" {
		return nil, errors.New("登录失败，请检查用户名和密码")
	}

	return &resp, nil
}

// getEntitlementToken 获取授权令牌
func (v *ValorantAPI) getEntitlementToken(accessToken string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, entitlementsURL, nil)
	if err != nil {
		return "", err
	}

	// 暂时保存令牌以便addCommonHeaders使用
	prevAccessToken := v.accessToken

	// 设置当前请求的令牌
	v.accessToken = accessToken

	// 添加通用头信息
	v.addCommonHeaders(req)

	// 恢复原来的令牌
	v.accessToken = prevAccessToken

	resp, err := v.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取授权令牌失败，状态码: %d", resp.StatusCode)
	}

	var entitlementResp models.ValorantEntitlementResponse
	if err := json.NewDecoder(resp.Body).Decode(&entitlementResp); err != nil {
		return "", err
	}

	return entitlementResp.EntitlementToken, nil
}

// getUserInfo 获取用户信息
func (v *ValorantAPI) getUserInfo(accessToken string) (*models.ValorantUserInfoResponse, error) {
	req, err := http.NewRequest(http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	// 暂时保存令牌以便addCommonHeaders使用
	prevAccessToken := v.accessToken

	// 设置当前请求的令牌
	v.accessToken = accessToken

	// 添加通用头信息
	v.addCommonHeaders(req)

	// 恢复原来的令牌
	v.accessToken = prevAccessToken

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取用户信息失败，状态码: %d", resp.StatusCode)
	}

	var userInfo models.ValorantUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// GetStoreOffers 获取商店物品
func (v *ValorantAPI) GetStoreOffers(userID, accessToken, entitlementToken string) (*models.ValorantStoreResponse, error) {
	// 确保使用有效的区域设置
	if v.region == "" {
		v.region = defaultRegion
		fmt.Printf("警告：未设置区域，将使用默认区域: %s\n", defaultRegion)
	}

	fmt.Printf("区域检查: 当前使用的区域是 %s\n", v.region)

	// 构建URL
	url := fmt.Sprintf(storeURL, v.region, userID)

	// 打印详细日志
	fmt.Printf("正在请求商店数据:\n")
	fmt.Printf("- 完整URL: %s\n", url)
	fmt.Printf("- 区域设置: %s\n", v.region)
	fmt.Printf("- 用户ID: %s\n", userID)

	fmt.Printf("- 请求方法: %s\n", http.MethodPost)

	// 创建请求 - 使用POST方法并包含空请求体
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString("{}"))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置完整的请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Riot-ClientPlatform", clientPlatform)
	req.Header.Set("X-Riot-ClientVersion", v.clientVersion)
	req.Header.Set("X-Riot-Entitlements-JWT", entitlementToken)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 打印完整的请求头信息（用于调试）
	fmt.Printf("- 请求头信息:\n")
	for key, values := range req.Header {
		// 跳过Authorization头以保护敏感信息
		if key == "Authorization" || key == "X-Riot-Entitlements-JWT" {
			fmt.Printf("  %s: ***敏感信息已隐藏***\n", key)
		} else {
			fmt.Printf("  %s: %s\n", key, values[0])
		}
	}

	resp, err := v.retryHTTPRequest(req, 3) // 重试3次
	if err != nil {
		fmt.Printf("获取商店数据失败: %v\n", err)
		return nil, fmt.Errorf("获取商店物品失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		fmt.Printf("请求失败:\n")
		fmt.Printf("- 状态码: %d\n", resp.StatusCode)
		fmt.Printf("- 响应体: %s\n", bodyStr)
		fmt.Printf("- 响应头:\n")
		for key, values := range resp.Header {
			fmt.Printf("  %s: %s\n", key, values[0])
		}

		// 针对特定错误提供更具体的错误信息
		if resp.StatusCode == 404 {
			fmt.Printf("警告: 获取到404错误，这可能意味着API接口路径已更改或用户区域不正确\n")
			fmt.Printf("尝试: 1. 确认用户所在区域 2. 验证API接口路径是否最新 3. 检查Riot客户端版本\n")

			// 尝试备用URL格式 - 有些区域可能使用不同的URL格式
			altURL := fmt.Sprintf("https://pd.%s.a.pvp.net/store/v2/storefront/%s", v.region, userID)
			fmt.Printf("将尝试备用URL: %s\n", altURL)

			// 创建备用请求
			altReq, err := http.NewRequest(http.MethodPost, altURL, bytes.NewBufferString("{}"))
			if err == nil {
				for key, values := range req.Header {
					if key != "X-Riot-ClientVersion" {
						for _, value := range values {
							altReq.Header.Add(key, value)
						}
					}
				}
				altReq.Header.Set("X-Riot-ClientVersion", v.clientVersion)

				fmt.Printf("正在尝试备用URL请求...\n")
				altResp, altErr := v.retryHTTPRequest(altReq, 2)
				if altErr == nil && altResp.StatusCode >= 200 && altResp.StatusCode < 300 {
					fmt.Printf("备用URL请求成功，状态码: %d\n", altResp.StatusCode)

					// 解析响应
					var storeResp models.ValorantStoreResponse
					if err := json.NewDecoder(altResp.Body).Decode(&storeResp); err == nil {
						altResp.Body.Close()
						fmt.Printf("成功获取商店数据(备用URL)\n")
						return &storeResp, nil
					}
					altResp.Body.Close()
				}

				if altResp != nil {
					altResp.Body.Close()
				}
			}
		}

		return nil, fmt.Errorf("获取商店数据失败，状态码: %d, 响应: %s", resp.StatusCode, bodyStr)
	}

	// 解析响应
	var storeResp models.ValorantStoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&storeResp); err != nil {
		return nil, fmt.Errorf("解析商店数据失败: %w", err)
	}

	fmt.Printf("成功获取商店数据\n")
	return &storeResp, nil
}

// GetWallet 获取用户钱包/余额
func (v *ValorantAPI) GetWallet(userID, accessToken, entitlementToken string) (*models.ValorantWalletResponse, error) {
	url := fmt.Sprintf(walletURL, v.region, userID)

	fmt.Printf("正在请求钱包数据:\n")
	fmt.Printf("- URL: %s\n", url)
	fmt.Printf("- 区域: %s\n", v.region)

	// 创建请求
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 暂时保存令牌以便addCommonHeaders使用
	prevAccessToken := v.accessToken
	prevEntitlementToken := v.entitlementsToken

	// 设置当前请求的令牌
	v.accessToken = accessToken
	v.entitlementsToken = entitlementToken

	// 添加通用头信息
	v.addCommonHeaders(req)

	// 恢复原来的令牌
	v.accessToken = prevAccessToken
	v.entitlementsToken = prevEntitlementToken

	// 添加特定于Riot客户端的头信息
	req.Header.Add("X-Riot-ClientPlatform", clientPlatform)
	req.Header.Add("X-Riot-ClientVersion", v.clientVersion)

	resp, err := v.retryHTTPRequest(req, 3) // 重试3次
	if err != nil {
		fmt.Printf("获取钱包数据失败: %v\n", err)
		return nil, fmt.Errorf("获取钱包信息失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		fmt.Printf("请求失败:\n")
		fmt.Printf("- 状态码: %d\n", resp.StatusCode)
		fmt.Printf("- 响应体: %s\n", bodyStr)

		return nil, fmt.Errorf("获取钱包数据失败，状态码: %d, 响应: %s", resp.StatusCode, bodyStr)
	}

	// 解析响应
	var walletResp models.ValorantWalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&walletResp); err != nil {
		return nil, fmt.Errorf("解析钱包数据失败: %w", err)
	}

	fmt.Printf("成功获取钱包数据\n")
	return &walletResp, nil
}

// GetContentInfo 获取游戏内容信息(包括皮肤等)
func (v *ValorantAPI) GetContentInfo() (interface{}, error) {
	url := fmt.Sprintf(contentURL, v.region)
	var contentResp interface{}

	err := v.makeRequest(http.MethodGet, url, nil, &contentResp)
	if err != nil {
		return nil, fmt.Errorf("获取内容信息失败: %w", err)
	}

	return contentResp, nil
}

// makeRequest 执行一个HTTP请求
func (v *ValorantAPI) makeRequest(method, url string, data interface{}, result interface{}) error {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	// 设置通用请求头
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "RiotClient/"+v.clientVersion)

	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 如果需要解析结果
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
	}

	return nil
}

// makeAuthorizedRequest 执行一个需要认证的HTTP请求
func (v *ValorantAPI) makeAuthorizedRequest(method, url string, data, result interface{}, accessToken, entitlementToken string) error {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	// 设置当前请求的令牌
	v.accessToken = accessToken
	v.entitlementsToken = entitlementToken

	// 添加通用头信息
	v.addCommonHeaders(req)

	fmt.Printf("发送HTTP请求:\n")
	fmt.Printf("- 方法: %s\n", method)
	fmt.Printf("- URL: %s\n", url)
	fmt.Printf("- 请求头:\n")
	for key, values := range req.Header {
		// 跳过Authorization头以保护敏感信息
		if key == "Authorization" || key == "X-Riot-Entitlements-JWT" {
			fmt.Printf("  %s: ***敏感信息已隐藏***\n", key)
		} else {
			fmt.Printf("  %s: %s\n", key, values[0])
		}
	}

	resp, err := v.client.Do(req)
	if err != nil {
		fmt.Printf("HTTP请求失败: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		fmt.Printf("请求失败:\n")
		fmt.Printf("- 状态码: %d\n", resp.StatusCode)
		fmt.Printf("- 响应体: %s\n", bodyStr)
		fmt.Printf("- 响应头:\n")
		for key, values := range resp.Header {
			fmt.Printf("  %s: %s\n", key, values[0])
		}

		return fmt.Errorf("授权请求失败，状态码: %d, 响应: %s", resp.StatusCode, bodyStr)
	}

	// 如果需要解析结果
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			fmt.Printf("解析响应体失败: %v\n", err)
			return err
		}
	}

	return nil
}

// parseAuthURI 从认证URI中解析访问令牌
func parseAuthURI(uri string) (string, error) {
	if uri == "" {
		return "", errors.New("空的认证URI")
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	fragment := parsedURL.Fragment
	params := strings.Split(fragment, "&")
	for _, param := range params {
		if strings.HasPrefix(param, "access_token=") {
			return strings.TrimPrefix(param, "access_token="), nil
		}
	}

	return "", errors.New("无法找到访问令牌")
}

// FilterEssentialCookies 过滤并保留可能有用的Cookie
func FilterEssentialCookies(cookies map[string]string) map[string]string {
	// 关键Cookie列表，这些是最重要的认证Cookie
	essentialNames := []string{"ssid", "sub", "csid", "clid", "tdid", "asid"}

	// 如果没有任何Cookie，返回空map
	if len(cookies) == 0 {
		return cookies
	}

	// 检查是否有任何关键Cookie
	hasAnyEssential := false
	for _, name := range essentialNames {
		if _, ok := cookies[name]; ok {
			hasAnyEssential = true
			break
		}
	}

	// 如果有任何关键Cookie，保留所有Cookie
	// 原始项目更宽松地处理Cookie，因此我们不再强制要求所有关键Cookie都存在
	if hasAnyEssential {
		return cookies
	}

	// 如果没有任何关键Cookie，返回原始的map
	// 可能会导致认证失败，但我们也可以继续尝试
	return cookies
}

// EnhancedParseCookieString 增强版Cookie解析函数，处理更多格式
func EnhancedParseCookieString(cookieStr string) map[string]string {
	cookieMap := make(map[string]string)

	// 如果输入为空，返回空map
	if cookieStr == "" {
		return cookieMap
	}

	// 清理字符串 - 处理多行和特殊字符
	cleanCookieStr := strings.ReplaceAll(cookieStr, "\r", "")
	cleanCookieStr = strings.ReplaceAll(cleanCookieStr, "\n", "")

	// 尝试多种分隔符
	var parts []string
	if strings.Contains(cleanCookieStr, ";") {
		parts = strings.Split(cleanCookieStr, ";")
	} else if strings.Contains(cleanCookieStr, ",") {
		parts = strings.Split(cleanCookieStr, ",")
	} else {
		// 单个Cookie的情况
		parts = []string{cleanCookieStr}
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 尝试几种不同的分隔方式
		var name, value string

		// 标准的name=value格式
		equalsIndex := strings.Index(part, "=")
		if equalsIndex != -1 {
			name = strings.TrimSpace(part[:equalsIndex])
			value = strings.TrimSpace(part[equalsIndex+1:])

			// 移除值两端的引号（如果有）
			value = strings.Trim(value, `"'`)

			if name != "" {
				cookieMap[name] = value
			}
			continue
		}

		// 尝试其他可能的格式
		colonIndex := strings.Index(part, ":")
		if colonIndex != -1 {
			name = strings.TrimSpace(part[:colonIndex])
			value = strings.TrimSpace(part[colonIndex+1:])
			value = strings.Trim(value, `"'`)

			if name != "" {
				cookieMap[name] = value
			}
		}
	}

	return cookieMap
}

// ParseCookieString 修改ParseCookieString函数，使用增强版
func ParseCookieString(cookieStr string) map[string]string {
	return EnhancedParseCookieString(cookieStr)
}

// AuthenticateWithCookies 使用Cookie进行认证
func (v *ValorantAPI) AuthenticateWithCookies(cookies map[string]string) (*models.UserSession, error) {
	// 尝试使用authorize端点进行认证
	session, err := v.authenticateWithCookiesViaAuthorizeEndpoint(cookies)
	if err != nil {
		// 如果失败，尝试使用auth端点
		session, err = v.authenticateWithCookiesViaAuthEndpoint(cookies)
		if err != nil {
			return nil, err
		}
	}

	// 获取用户信息并设置Riot用户名
	userInfo, err := v.getUserInfo(session.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	session.RiotUsername = userInfo.Name
	session.RiotTagline = userInfo.Tag

	// 设置默认区域
	session.Region = ""

	return session, nil
}

// 主要认证方法 - 通过authorize端点
func (v *ValorantAPI) authenticateWithCookiesViaAuthorizeEndpoint(cookies map[string]string) (*models.UserSession, error) {
	// 创建带有cookie jar的新HTTP客户端
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// 使用原始项目的策略 - 禁用重定向后处理
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 禁止自动跟随重定向
			return http.ErrUseLastResponse
		},
	}

	// 使用这个带cookie的客户端替换当前的客户端
	v.client = client

	// 过滤保留有用的Cookie（但不再强制要求特定Cookie）
	essentialCookies := FilterEssentialCookies(cookies)
	if len(essentialCookies) == 0 {
		return nil, errors.New("没有提供任何Cookie")
	}

	// 尝试方法一：直接使用cookie点击另一个端点
	userInfoReq, err := http.NewRequest(http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	setRiotRequestHeaders(userInfoReq, essentialCookies)
	userInfoResp, err := v.client.Do(userInfoReq)

	// 如果直接获取用户信息成功，说明cookie有效
	if err == nil && userInfoResp.StatusCode == http.StatusOK {
		defer userInfoResp.Body.Close()
		var userInfo models.ValorantUserInfoResponse
		if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err == nil {
			// 获取授权令牌
			entitlementToken, err := v.getEntitlementToken(essentialCookies["ssid"])
			if err == nil {
				// 创建用户会话
				session := &models.UserSession{
					UserID:       userInfo.Sub,
					Username:     userInfo.Email,
					AccessToken:  essentialCookies["ssid"], // 使用ssid作为token
					Entitlement:  entitlementToken,
					RiotUsername: userInfo.Name,
					RiotTagline:  userInfo.Tag,
					Cookies:      essentialCookies,
				}
				return session, nil
			}
		}
	}
	if userInfoResp != nil {
		userInfoResp.Body.Close()
	}

	// 尝试方法二：通过authorize端点获取token
	// 构建请求URL - 使用和原始项目完全相同的URL
	authURL := "https://auth.riotgames.com/authorize?redirect_uri=https%3A%2F%2Fplayvalorant.com%2Fopt_in&client_id=play-valorant-web-prod&response_type=token%20id_token&scope=account%20openid&nonce=1"

	// 创建请求
	req, err := http.NewRequest(http.MethodGet, authURL, nil)
	if err != nil {
		return nil, err
	}

	// 添加所有必要的头部
	setRiotRequestHeaders(req, essentialCookies)

	// 发送请求
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查状态码，应该是302或303（重定向）
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("认证请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 获取Location头部
	location := resp.Header.Get("Location")
	if location == "" {
		return nil, errors.New("响应中没有Location头")
	}

	if strings.Contains(location, "/login") {
		return nil, errors.New("Cookie无效或已过期")
	}

	if !strings.Contains(location, "access_token=") {
		return nil, errors.New("无法从响应中提取令牌")
	}

	// 从Location URL中提取访问令牌
	accessToken, err := parseAccessTokenFromURI(location)
	if err != nil {
		return nil, fmt.Errorf("提取访问令牌失败: %w", err)
	}

	// 获取授权令牌
	entitlementToken, err := v.getEntitlementToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取授权令牌失败: %w", err)
	}

	// 获取用户信息
	userInfo, err := v.getUserInfo(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 创建用户会话
	session := &models.UserSession{
		UserID:       userInfo.Sub,
		Username:     userInfo.Email,
		AccessToken:  accessToken,
		Entitlement:  entitlementToken,
		RiotUsername: userInfo.Name,
		RiotTagline:  userInfo.Tag,
		Cookies:      essentialCookies,
	}

	return session, nil
}

// 备用认证方法 - 通过auth端点 (模仿原始项目)
func (v *ValorantAPI) authenticateWithCookiesViaAuthEndpoint(cookies map[string]string) (*models.UserSession, error) {
	// 创建带有cookie jar的新HTTP客户端
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	// 使用这个带cookie的客户端替换当前的客户端
	v.client = client

	// 过滤保留有用的Cookie
	essentialCookies := FilterEssentialCookies(cookies)
	if len(essentialCookies) == 0 {
		return nil, errors.New("没有提供任何Cookie")
	}

	// 定义API端点URL
	userInfoURL := "https://auth.riotgames.com/userinfo"

	// 尝试原始项目的方法: 直接通过ssid获取信息
	if ssid, ok := essentialCookies["ssid"]; ok {
		// 尝试使用ssid直接请求用户信息
		req, err := http.NewRequest(http.MethodGet, userInfoURL, nil)
		if err != nil {
			return nil, err
		}

		// 使用ssid作为Bearer令牌
		req.Header.Add("Authorization", "Bearer "+ssid)
		req.Header.Add("Content-Type", "application/json")

		resp, err := v.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// 如果成功获取用户信息
		if resp.StatusCode == http.StatusOK {
			var userInfo models.ValorantUserInfoResponse
			if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
				return nil, fmt.Errorf("解析用户信息失败: %w", err)
			}

			// 尝试获取entitlement令牌
			entitlementToken, err := v.getEntitlementToken(ssid)
			if err != nil {
				return nil, fmt.Errorf("获取授权令牌失败: %w", err)
			}

			// 创建会话
			session := &models.UserSession{
				UserID:       userInfo.Sub,
				Username:     userInfo.Email,
				AccessToken:  ssid,
				Entitlement:  entitlementToken,
				RiotUsername: userInfo.Name,
				RiotTagline:  userInfo.Tag,
				Cookies:      essentialCookies,
			}

			return session, nil
		}
	}

	// 第一步：尝试获取当前的cookie状态
	authCheckURL := "https://auth.riotgames.com/authorize?redirect_uri=https%3A%2F%2Fplayvalorant.com%2Fopt_in&client_id=play-valorant-web-prod&response_type=token%20id_token&nonce=1"
	req, err := http.NewRequest(http.MethodGet, authCheckURL, nil)
	if err != nil {
		return nil, err
	}

	// 添加Cookie和请求头
	setRiotRequestHeaders(req, essentialCookies)

	// 发送请求
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// 第二步：尝试获取认证状态
	req, err = http.NewRequest(http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	// 尝试从Response中获取Cookie并添加到请求
	for _, cookie := range v.client.Jar.Cookies(req.URL) {
		essentialCookies[cookie.Name] = cookie.Value
	}

	// 添加Cookie和请求头
	setRiotRequestHeaders(req, essentialCookies)

	// 发送请求
	resp, err = v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 如果能够获取用户信息，说明Cookie有效
	if resp.StatusCode == http.StatusOK {
		var userInfo models.ValorantUserInfoResponse
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return nil, fmt.Errorf("解析用户信息失败: %w", err)
		}

		// 尝试获取authorization token
		req, err = http.NewRequest(http.MethodPost, "https://auth.riotgames.com/api/token/v1", bytes.NewBuffer([]byte("{}")))
		if err != nil {
			return nil, err
		}

		setRiotRequestHeaders(req, essentialCookies)

		resp, err = v.client.Do(req)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		// 尝试获取entitlements token
		req, err = http.NewRequest(http.MethodPost, "https://entitlements.auth.riotgames.com/api/token/v1", bytes.NewBuffer([]byte("{}")))
		if err != nil {
			return nil, err
		}

		setRiotRequestHeaders(req, essentialCookies)

		resp, err = v.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var entitlementResp struct {
			EntitlementToken string `json:"entitlements_token"`
		}

		if resp.StatusCode == http.StatusOK {
			if err := json.NewDecoder(resp.Body).Decode(&entitlementResp); err != nil {
				return nil, fmt.Errorf("解析entitlement令牌失败: %w", err)
			}

			// 创建临时会话
			session := &models.UserSession{
				UserID:       userInfo.Sub,
				Username:     userInfo.Email,
				AccessToken:  getTokenValue(essentialCookies), // 优先使用ssid
				Entitlement:  entitlementResp.EntitlementToken,
				RiotUsername: userInfo.Name,
				RiotTagline:  userInfo.Tag,
				Cookies:      essentialCookies,
			}

			return session, nil
		}
	}

	return nil, errors.New("备用认证方法失败")
}

// getTokenValue 从cookie中获取合适的token值
func getTokenValue(cookies map[string]string) string {
	if token, ok := cookies["ssid"]; ok {
		return token
	}
	return "cookie-session"
}

// setRiotRequestHeaders 统一设置Riot API请求头
func setRiotRequestHeaders(req *http.Request, cookies map[string]string) {
	cookieString := StringifyCookies(cookies)
	req.Header.Set("Cookie", cookieString)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RiotClient/58.0.0.4640299.4552318 rso-auth (Windows;10;;Professional, x64)")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://auth.riotgames.com")
	req.Header.Set("DNT", "1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
}

// parseAccessTokenFromURI 从授权URI解析access_token
// 不同于完整的extractTokensFromUri，这个函数只提取access_token
func parseAccessTokenFromURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	fragment := u.Fragment
	if fragment == "" {
		return "", errors.New("URI中没有找到fragment部分")
	}

	values, err := url.ParseQuery(fragment)
	if err != nil {
		return "", err
	}

	accessToken := values.Get("access_token")
	if accessToken == "" {
		return "", errors.New("没有找到access_token")
	}

	return accessToken, nil
}

// StringifyCookies 将cookie map转换为字符串
func StringifyCookies(cookies map[string]string) string {
	parts := make([]string, 0, len(cookies))
	for name, value := range cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", name, value))
	}
	return strings.Join(parts, "; ")
}

// retryHTTPRequest 执行HTTP请求，带有指数退避重试
func (v *ValorantAPI) retryHTTPRequest(req *http.Request, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error
	var attempt int

	// 基础等待时间（毫秒）
	baseWaitMS := 500

	for attempt = 0; attempt <= maxRetries; attempt++ {
		// 除了第一次尝试外，记录重试次数
		if attempt > 0 {
			waitTime := time.Duration(baseWaitMS*1<<uint(attempt-1)) * time.Millisecond
			if waitTime > 10*time.Second {
				waitTime = 10 * time.Second // 最大等待10秒
			}
			fmt.Printf("第%d次重试请求 %s %s，等待%v后重试...\n",
				attempt, req.Method, req.URL.String(), waitTime)
			time.Sleep(waitTime)

			// 重新创建请求，避免使用已关闭的请求
			newReq := req.Clone(req.Context())
			// 复制原始请求头
			newReq.Header = req.Header.Clone()
			req = newReq
		}

		resp, err = v.client.Do(req)

		// 如果没有错误，直接返回响应
		if err == nil {
			if attempt > 0 {
				fmt.Printf("在第%d次尝试后成功执行请求\n", attempt)
			}
			return resp, nil
		}

		// 检查错误类型，决定是否继续重试
		if attempt == maxRetries {
			fmt.Printf("达到最大重试次数(%d)，放弃请求\n", maxRetries)
			break
		}

		// 判断错误类型，只对网络相关错误进行重试
		if isNetworkError(err) {
			fmt.Printf("检测到网络错误: %v, 将重试...\n", err)
			continue
		} else {
			// 其他错误不重试
			fmt.Printf("收到非网络错误，放弃重试: %v\n", err)
			break
		}
	}

	return nil, fmt.Errorf("请求失败，经过%d次尝试: %w", attempt, err)
}

// isNetworkError 检查是否为网络连接相关错误
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// 检查常见的网络错误
	if _, ok := err.(net.Error); ok {
		// 网络错误通常都需要重试
		return true
	}

	// 检查URL错误，比如DNS查询失败等
	if urlErr, ok := err.(*url.Error); ok {
		// URL错误通常是由底层错误引起的，递归检查
		return isNetworkError(urlErr.Err)
	}

	// 检查HTTP客户端错误
	if strings.Contains(err.Error(), "Client.Timeout") {
		return true
	}

	// 检查标准错误字符串
	errMsg := err.Error()

	// 连接相关错误
	if strings.Contains(errMsg, "connection") && (strings.Contains(errMsg, "reset") ||
		strings.Contains(errMsg, "closed") ||
		strings.Contains(errMsg, "refused") ||
		strings.Contains(errMsg, "broken")) {
		return true
	}

	// EOF相关错误
	if strings.Contains(errMsg, "EOF") ||
		strings.Contains(errMsg, "unexpected EOF") ||
		strings.Contains(errMsg, "read: connection closed") ||
		strings.Contains(errMsg, "body closed") {
		return true
	}

	// 超时相关错误
	if strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline exceeded") {
		return true
	}

	// 网络繁忙或服务器错误
	if strings.Contains(errMsg, "server closed idle connection") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "TLS handshake timeout") ||
		strings.Contains(errMsg, "503") || // 服务不可用
		strings.Contains(errMsg, "504") { // 网关超时
		return true
	}

	// 默认情况下，未知错误不视为网络错误
	return false
}

// addCommonHeaders 添加HTTP请求所需的通用头信息
func (v *ValorantAPI) addCommonHeaders(req *http.Request) {
	// 设置通用请求头
	req.Header.Set("User-Agent", "RiotClient/43.0.1.4195386.4190634 rso-auth (Windows;10;;Professional, x64)")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://playvalorant.com")
	req.Header.Set("Referer", "https://playvalorant.com/")
	req.Header.Set("Content-Type", "application/json")

	// 添加一个dummy cookie帮助避免Cloudflare问题
	req.Header.Set("Cookie", "dummy=value")

	// 特定头信息设置
	if v.entitlementsToken != "" {
		req.Header.Set("X-Riot-Entitlements-JWT", v.entitlementsToken)
	}
	if v.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+v.accessToken)
	}
}
