package storage

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// DBStorage 数据库存储实现
// 将令牌数据持久化到关系型数据库
type DBStorage struct {
	db *gorm.DB
}

// NewDBStorage 创建数据库存储实例
func NewDBStorage(db *gorm.DB) (*DBStorage, error) {
	// 自动迁移数据库表
	if err := db.AutoMigrate(
		&DBComponentToken{},
		&DBPreAuthCode{},
		&DBAuthorizerToken{},
	); err != nil {
		return nil, err
	}

	return &DBStorage{db: db}, nil
}

// DBComponentToken 组件令牌数据库模型
type DBComponentToken struct {
	ID          uint      `gorm:"primaryKey"`
	AccessToken string    `gorm:"type:varchar(512);not null"`
	ExpiresIn   int       `gorm:"not null"`
	ExpiresAt   time.Time `gorm:"not null;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DBPreAuthCode 预授权码数据库模型
type DBPreAuthCode struct {
	ID          uint      `gorm:"primaryKey"`
	PreAuthCode string    `gorm:"type:varchar(256);not null"`
	ExpiresIn   int       `gorm:"not null"`
	ExpiresAt   time.Time `gorm:"not null;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DBAuthorizerToken 授权方令牌数据库模型
type DBAuthorizerToken struct {
	ID                     uint      `gorm:"primaryKey"`
	AuthorizerAppID        string    `gorm:"type:varchar(64);not null;uniqueIndex"`
	AuthorizerAccessToken  string    `gorm:"type:varchar(512);not null"`
	AuthorizerRefreshToken string    `gorm:"type:varchar(512)"`
	ExpiresIn              int       `gorm:"not null"`
	ExpiresAt              time.Time `gorm:"not null;index"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// SaveComponentToken 保存组件令牌到数据库
func (s *DBStorage) SaveComponentToken(ctx context.Context, token *ComponentAccessToken) error {
	dbToken := &DBComponentToken{
		AccessToken: token.AccessToken,
		ExpiresIn:   token.ExpiresIn,
		ExpiresAt:   token.ExpiresAt,
	}

	// 使用事务确保数据一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除旧的令牌
		if err := tx.Where("1 = 1").Delete(&DBComponentToken{}).Error; err != nil {
			return err
		}

		// 保存新的令牌
		return tx.Create(dbToken).Error
	})
}

// GetComponentToken 从数据库读取组件令牌
func (s *DBStorage) GetComponentToken(ctx context.Context) (*ComponentAccessToken, error) {
	var dbToken DBComponentToken

	// 获取最新的令牌记录
	if err := s.db.Order("created_at DESC").First(&dbToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(dbToken.ExpiresAt) {
		// 自动删除过期令牌
		s.db.Delete(&dbToken)
		return nil, nil
	}

	return &ComponentAccessToken{
		AccessToken: dbToken.AccessToken,
		ExpiresIn:   dbToken.ExpiresIn,
		ExpiresAt:   dbToken.ExpiresAt,
	}, nil
}

// DeleteComponentToken 删除组件令牌
func (s *DBStorage) DeleteComponentToken(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&DBComponentToken{}).Error
}

// SavePreAuthCode 保存预授权码到数据库
func (s *DBStorage) SavePreAuthCode(ctx context.Context, code *PreAuthCode) error {
	dbCode := &DBPreAuthCode{
		PreAuthCode: code.PreAuthCode,
		ExpiresIn:   code.ExpiresIn,
		ExpiresAt:   code.ExpiresAt,
	}

	// 使用事务确保数据一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除旧的预授权码
		if err := tx.Where("1 = 1").Delete(&DBPreAuthCode{}).Error; err != nil {
			return err
		}

		// 保存新的预授权码
		return tx.Create(dbCode).Error
	})
}

// GetPreAuthCode 从数据库读取预授权码
func (s *DBStorage) GetPreAuthCode(ctx context.Context) (*PreAuthCode, error) {
	var dbCode DBPreAuthCode

	// 获取最新的预授权码记录
	if err := s.db.Order("created_at DESC").First(&dbCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(dbCode.ExpiresAt) {
		// 自动删除过期预授权码
		s.db.Delete(&dbCode)
		return nil, nil
	}

	return &PreAuthCode{
		PreAuthCode: dbCode.PreAuthCode,
		ExpiresIn:   dbCode.ExpiresIn,
		ExpiresAt:   dbCode.ExpiresAt,
	}, nil
}

// DeletePreAuthCode 删除预授权码
func (s *DBStorage) DeletePreAuthCode(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&DBPreAuthCode{}).Error
}

// SaveAuthorizerToken 保存授权方令牌到数据库
func (s *DBStorage) SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error {
	// 使用upsert操作（存在则更新，不存在则插入）
	return s.db.Where(DBAuthorizerToken{AuthorizerAppID: authorizerAppID}).
		Assign(DBAuthorizerToken{
			AuthorizerAccessToken:  token.AuthorizerAccessToken,
			AuthorizerRefreshToken: token.AuthorizerRefreshToken,
			ExpiresIn:              token.ExpiresIn,
			ExpiresAt:              token.ExpiresAt,
		}).
		FirstOrCreate(&DBAuthorizerToken{}, DBAuthorizerToken{AuthorizerAppID: authorizerAppID}).Error
}

// GetAuthorizerToken 从数据库读取授权方令牌
func (s *DBStorage) GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error) {
	var dbToken DBAuthorizerToken

	if err := s.db.Where("authorizer_app_id = ?", authorizerAppID).First(&dbToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(dbToken.ExpiresAt) {
		// 自动删除过期令牌
		s.db.Delete(&dbToken)
		return nil, nil
	}

	return &AuthorizerAccessToken{
		AuthorizerAppID:        authorizerAppID,
		AuthorizerAccessToken:  dbToken.AuthorizerAccessToken,
		AuthorizerRefreshToken: dbToken.AuthorizerRefreshToken,
		ExpiresIn:              dbToken.ExpiresIn,
		ExpiresAt:              dbToken.ExpiresAt,
	}, nil
}

// DeleteAuthorizerToken 删除授权方令牌
func (s *DBStorage) DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error {
	return s.db.Where("authorizer_app_id = ?", authorizerAppID).Delete(&DBAuthorizerToken{}).Error
}

// ClearAuthorizerTokens 清除所有授权方令牌
func (s *DBStorage) ClearAuthorizerTokens(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&DBAuthorizerToken{}).Error
}

// ListAuthorizerTokens 列出所有已存储的授权方appid
func (s *DBStorage) ListAuthorizerTokens(ctx context.Context) ([]string, error) {
	var tokens []DBAuthorizerToken
	if err := s.db.Select("authorizer_app_id").Find(&tokens).Error; err != nil {
		return nil, err
	}

	appids := make([]string, len(tokens))
	for i, token := range tokens {
		appids[i] = token.AuthorizerAppID
	}

	return appids, nil
}

// Ping 存储健康检查
func (s *DBStorage) Ping(ctx context.Context) error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}