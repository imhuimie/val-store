package services

import (
	"errors"
	"sort"

	"github.com/emper0r/val-store/server/internal/models"
	"github.com/emper0r/val-store/server/internal/repositories"
)

// SkinsService 处理皮肤相关的业务逻辑
type SkinsService struct {
	valorantAPI  *repositories.ValorantAPI
	skinDatabase *repositories.SkinDatabase
}

// NewSkinsService 创建新的皮肤服务
func NewSkinsService(valorantAPI *repositories.ValorantAPI, skinDatabase *repositories.SkinDatabase) *SkinsService {
	return &SkinsService{
		valorantAPI:  valorantAPI,
		skinDatabase: skinDatabase,
	}
}

// GetAllSkins 获取所有皮肤列表
func (s *SkinsService) GetAllSkins() ([]models.Skin, error) {
	// 从数据库获取所有皮肤
	skins := s.skinDatabase.GetAllSkins()

	// 如果数据库为空或需要更新，尝试更新
	if len(skins) == 0 || s.skinDatabase.NeedsUpdate() {
		if err := s.UpdateSkinsDatabase(); err != nil {
			// 如果数据库为空且更新失败，返回错误
			if len(skins) == 0 {
				return nil, err
			}
			// 否则继续使用旧数据
		}

		// 重新获取更新后的数据
		skins = s.skinDatabase.GetAllSkins()
	}

	// 如果仍然为空，返回错误
	if len(skins) == 0 {
		return nil, errors.New("皮肤数据库为空")
	}

	// 按武器名和皮肤名排序
	sort.Slice(skins, func(i, j int) bool {
		if skins[i].WeaponName == skins[j].WeaponName {
			return skins[i].Name < skins[j].Name
		}
		return skins[i].WeaponName < skins[j].WeaponName
	})

	return skins, nil
}

// GetSkinByID 根据ID获取皮肤信息
func (s *SkinsService) GetSkinByID(skinID string) (models.Skin, error) {
	skin, found := s.skinDatabase.GetSkinByID(skinID)
	if !found {
		return models.Skin{}, errors.New("皮肤未找到")
	}
	return skin, nil
}

// UpdateSkinsDatabase 更新皮肤数据库
// 注：这个方法应该被服务器启动时调用，或者通过管理端点触发
func (s *SkinsService) UpdateSkinsDatabase() error {
	// 这里需要与Valorant API交互获取最新的皮肤列表
	// 由于这需要直接调用Valorant的内容API，可能需要特殊处理
	// 在实际实现中，这可能是一个复杂的过程

	// 这里是一个简化的实现
	// 实际项目中，您需要按照Valorant API的文档实现具体逻辑
	// 例如解析Valorant的内容服务返回的数据

	// 作为示例，我们返回一个未实现错误
	// 在真实实现中，您需要替换为实际的API调用和数据处理逻辑

	return errors.New("更新皮肤数据库功能尚未实现")

	// 完整实现示例
	/*
		contentData, err := s.valorantAPI.GetContentInfo()
		if err != nil {
			return err
		}

		// 解析contentData，提取皮肤信息
		var skins []models.Skin

		// 处理逻辑...

		// 更新数据库
		return s.skinDatabase.UpdateSkinDatabase(skins)
	*/
}
