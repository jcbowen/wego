package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/orm/mysql"
	"gorm.io/gorm"
)

// DBStorage 数据库存储实现
// 将令牌数据持久化到关系型数据库
type DBStorage struct {
	db *gorm.DB
}

// NewDBStorage 创建数据库存储实例
// @param dbConfig jcbaseGo.DbStruct 数据库配置结构
// @param opts ...string 可选参数
// @return *DBStorage 数据库存储实例
// @return error 错误信息
func NewDBStorage(dbConfig jcbaseGo.DbStruct, opts ...string) (*DBStorage, error) {
	// 创建mysql实例
	mysqlInstance := mysql.New(dbConfig, opts...)

	// 获取GORM数据库连接
	db := mysqlInstance.GetDb()

	// 自动迁移数据库表
	if err := db.AutoMigrate(
		&DBComponentToken{},
		&DBPreAuthCode{},
		&DBAuthorizerToken{},
		&DBPrevEncodingAESKey{},
		&DBComponentVerifyTicket{},
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

// DBPrevEncodingAESKey 上一次EncodingAESKey数据库模型
type DBPrevEncodingAESKey struct {
	ID              uint   `gorm:"primaryKey"`
	AppID           string `gorm:"type:varchar(64);not null;uniqueIndex"`
	PrevEncodingKey string `gorm:"type:varchar(256);not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DBComponentVerifyTicket 验证票据数据库模型
type DBComponentVerifyTicket struct {
	ID        uint      `gorm:"primaryKey"`
	Ticket    string    `gorm:"type:varchar(512);not null"` // 票据内容
	CreatedAt time.Time `gorm:"not null"`                   // 创建时间
	ExpiresAt time.Time `gorm:"not null;index"`             // 过期时间（创建时间+12小时）
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
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

// SavePrevEncodingAESKey 保存上一次EncodingAESKey到数据库
func (s *DBStorage) SavePrevEncodingAESKey(ctx context.Context, appID string, prevKey string) error {
	// 使用upsert操作（存在则更新，不存在则插入）
	return s.db.Where(DBPrevEncodingAESKey{AppID: appID}).
		Assign(DBPrevEncodingAESKey{
			PrevEncodingKey: prevKey,
		}).
		FirstOrCreate(&DBPrevEncodingAESKey{}, DBPrevEncodingAESKey{AppID: appID}).Error
}

// GetPrevEncodingAESKey 从数据库读取上一次EncodingAESKey
func (s *DBStorage) GetPrevEncodingAESKey(ctx context.Context, appID string) (*PrevEncodingAESKey, error) {
	var dbKey DBPrevEncodingAESKey

	if err := s.db.Where("app_id = ?", appID).First(&dbKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &PrevEncodingAESKey{
		AppID:              appID,
		PrevEncodingAESKey: dbKey.PrevEncodingKey,
		UpdatedAt:          time.Now(),
	}, nil
}

// DeletePrevEncodingAESKey 删除上一次EncodingAESKey
func (s *DBStorage) DeletePrevEncodingAESKey(ctx context.Context, appID string) error {
	return s.db.Where("app_id = ?", appID).Delete(&DBPrevEncodingAESKey{}).Error
}

// SaveComponentVerifyTicket 保存验证票据到数据库
// @param ctx context.Context 上下文
// @param ticket string 票据内容
// @return error 错误信息
func (s *DBStorage) SaveComponentVerifyTicket(ctx context.Context, ticket string) error {
	// 票据有效期为12小时
	expiresAt := time.Now().Add(12 * time.Hour)
	dbTicket := &DBComponentVerifyTicket{
		Ticket:    ticket,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	// 使用事务确保数据一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除旧的票据
		if err := tx.Where("1 = 1").Delete(&DBComponentVerifyTicket{}).Error; err != nil {
			return err
		}

		// 保存新的票据
		return tx.Create(dbTicket).Error
	})
}

// GetComponentVerifyTicket 从数据库读取验证票据
// @param ctx context.Context 上下文
// @return *ComponentVerifyTicket 票据结构，包含创建时间
// @return error 错误信息
func (s *DBStorage) GetComponentVerifyTicket(ctx context.Context) (*ComponentVerifyTicket, error) {
	var dbTicket DBComponentVerifyTicket

	// 获取最新的票据记录
	if err := s.db.Order("created_at DESC").First(&dbTicket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(dbTicket.ExpiresAt) {
		// 自动删除过期票据
		s.db.Delete(&dbTicket)
		return nil, nil
	}

	return &ComponentVerifyTicket{
		Ticket:    dbTicket.Ticket,
		CreatedAt: dbTicket.CreatedAt,
		ExpiresAt: dbTicket.ExpiresAt,
	}, nil
}

// DeleteComponentVerifyTicket 删除验证票据
// @param ctx context.Context 上下文
// @return error 错误信息
func (s *DBStorage) DeleteComponentVerifyTicket(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&DBComponentVerifyTicket{}).Error
}
