package repositories

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/emper0r/val-store/server/internal/models"
)

const (
	// SkinsDBPath 皮肤数据库的默认路径
	SkinsDBPath = "data/skins.json"
)

// SkinDatabase 皮肤数据库结构
type SkinDatabase struct {
	db        models.SkinsDatabase
	filePath  string
	mutex     sync.RWMutex
	lastCheck time.Time
}

// NewSkinDatabase 创建一个新的皮肤数据库实例
func NewSkinDatabase(filePath string) (*SkinDatabase, error) {
	if filePath == "" {
		// 确保data目录存在
		if err := os.MkdirAll("data", 0755); err != nil {
			return nil, fmt.Errorf("创建数据目录失败: %w", err)
		}
		filePath = SkinsDBPath
	}

	// 创建绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取绝对路径失败: %w", err)
	}

	db := &SkinDatabase{
		db:        models.SkinsDatabase{Skins: []models.Skin{}},
		filePath:  absPath,
		lastCheck: time.Time{},
	}

	// 尝试加载已有的数据库文件
	if err := db.loadFromFile(); err != nil {
		// 文件不存在不是错误，仅记录
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("加载皮肤数据库失败: %w", err)
		}
	}

	return db, nil
}

// loadFromFile 从文件加载皮肤数据库
func (s *SkinDatabase) loadFromFile() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查文件是否存在
	_, err := os.Stat(s.filePath)
	if err != nil {
		return err
	}

	// 读取文件内容
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("读取皮肤数据库文件失败: %w", err)
	}

	// 解析JSON
	var db models.SkinsDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return fmt.Errorf("解析皮肤数据库JSON失败: %w", err)
	}

	s.db = db
	s.lastCheck = time.Now()
	return nil
}

// saveToFile 将皮肤数据库保存到文件
func (s *SkinDatabase) saveToFile() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 将数据库序列化为JSON
	data, err := json.MarshalIndent(s.db, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化皮肤数据库失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入皮肤数据库文件失败: %w", err)
	}

	return nil
}

// GetSkinByID 根据ID获取皮肤信息
func (s *SkinDatabase) GetSkinByID(skinID string) (models.Skin, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, skin := range s.db.Skins {
		if skin.UUID == skinID {
			return skin, true
		}
	}

	return models.Skin{}, false
}

// GetAllSkins 获取所有皮肤
func (s *SkinDatabase) GetAllSkins() []models.Skin {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回皮肤列表的副本
	skins := make([]models.Skin, len(s.db.Skins))
	copy(skins, s.db.Skins)
	return skins
}

// UpdateSkinDatabase 使用从Valorant API获取的皮肤信息更新数据库
func (s *SkinDatabase) UpdateSkinDatabase(skins []models.Skin) error {
	s.mutex.Lock()
	s.db.Skins = skins
	s.mutex.Unlock()

	// 保存到文件
	return s.saveToFile()
}

// NeedsUpdate 检查数据库是否需要更新（超过24小时未更新）
func (s *SkinDatabase) NeedsUpdate() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 如果数据库为空，需要更新
	if len(s.db.Skins) == 0 {
		return true
	}

	// 如果距离上次检查超过24小时，需要更新
	return time.Since(s.lastCheck) > 24*time.Hour
}

// Count 返回数据库中皮肤的数量
func (s *SkinDatabase) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.db.Skins)
}
