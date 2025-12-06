package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/orm/base"
	"github.com/jcbowen/jcbaseGo/component/orm/sqlite"
	"gorm.io/gorm"
)

type SqliteStorage struct {
	db *gorm.DB
}

func NewSqliteStorage(conf jcbaseGo.SqlLiteStruct, opts ...string) (*SqliteStorage, error) {
	inst, err := sqlite.New(conf, opts...)
	if err != nil {
		return nil, err
	}
	db := inst.GetDb()
	if err := db.AutoMigrate(
		&DBComponentTokenSqlite{},
		&DBPreAuthCodeSqlite{},
		&DBAuthorizerTokenSqlite{},
		&DBPrevEncodingAESKeySqlite{},
		&DBComponentVerifyTicketSqlite{},
	); err != nil {
		return nil, err
	}
	return &SqliteStorage{db: db}, nil
}

type DBComponentTokenSqlite struct {
	base.SqliteBaseModel
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AccessToken string    `gorm:"column:access_token;type:varchar(512);not null" json:"access_token"`
	ExpiresIn   int       `gorm:"column:expires_in;not null" json:"expires_in"`
	ExpiresAt   time.Time `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type DBPreAuthCodeSqlite struct {
	base.SqliteBaseModel
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PreAuthCode string    `gorm:"column:pre_auth_code;type:varchar(256);not null" json:"pre_auth_code"`
	ExpiresIn   int       `gorm:"column:expires_in;not null" json:"expires_in"`
	ExpiresAt   time.Time `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type DBAuthorizerTokenSqlite struct {
	base.SqliteBaseModel
	ID                     uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AuthorizerAppID        string    `gorm:"column:authorizer_app_id;type:varchar(64);not null;uniqueIndex" json:"authorizer_app_id"`
	AuthorizerAccessToken  string    `gorm:"column:authorizer_access_token;type:varchar(512);not null" json:"authorizer_access_token"`
	AuthorizerRefreshToken string    `gorm:"column:authorizer_refresh_token;type:varchar(512)" json:"authorizer_refresh_token"`
	ExpiresIn              int       `gorm:"column:expires_in;not null" json:"expires_in"`
	ExpiresAt              time.Time `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt              time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type DBPrevEncodingAESKeySqlite struct {
	base.SqliteBaseModel
	ID              uint      `gorm:"primaryKey"`
	AppID           string    `gorm:"type:varchar(64);not null;uniqueIndex"`
	PrevEncodingKey string    `gorm:"column:prev_encoding_key;type:varchar(256);not null" json:"prev_encoding_key"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type DBComponentVerifyTicketSqlite struct {
	base.SqliteBaseModel
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Ticket    string    `gorm:"column:ticket;type:varchar(512);not null" json:"ticket"`
	ExpiresAt time.Time `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (s *SqliteStorage) SaveComponentToken(ctx context.Context, token *ComponentAccessToken) error {
	dbToken := &DBComponentTokenSqlite{
		AccessToken: token.AccessToken,
		ExpiresIn:   token.ExpiresIn,
		ExpiresAt:   token.ExpiresAt,
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&DBComponentTokenSqlite{}).Error; err != nil {
			return err
		}
		return tx.Create(dbToken).Error
	})
}

func (s *SqliteStorage) GetComponentToken(ctx context.Context) (*ComponentAccessToken, error) {
	var dbToken DBComponentTokenSqlite
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

func (s *SqliteStorage) DeleteComponentToken(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&DBComponentTokenSqlite{}).Error
}

func (s *SqliteStorage) SavePreAuthCode(ctx context.Context, code *PreAuthCode) error {
	dbCode := &DBPreAuthCodeSqlite{
		PreAuthCode: code.PreAuthCode,
		ExpiresIn:   code.ExpiresIn,
		ExpiresAt:   code.ExpiresAt,
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&DBPreAuthCodeSqlite{}).Error; err != nil {
			return err
		}
		return tx.Create(dbCode).Error
	})
}

func (s *SqliteStorage) GetPreAuthCode(ctx context.Context) (*PreAuthCode, error) {
	var dbCode DBPreAuthCodeSqlite
	if err := s.db.Order("created_at DESC").First(&dbCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &PreAuthCode{
		PreAuthCode: dbCode.PreAuthCode,
		ExpiresIn:   dbCode.ExpiresIn,
		ExpiresAt:   dbCode.ExpiresAt,
	}, nil
}

func (s *SqliteStorage) DeletePreAuthCode(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&DBPreAuthCodeSqlite{}).Error
}

func (s *SqliteStorage) SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error {
	return s.db.Where(DBAuthorizerTokenSqlite{AuthorizerAppID: authorizerAppID}).
		Assign(DBAuthorizerTokenSqlite{
			AuthorizerAccessToken:  token.AuthorizerAccessToken,
			AuthorizerRefreshToken: token.AuthorizerRefreshToken,
			ExpiresIn:              token.ExpiresIn,
			ExpiresAt:              token.ExpiresAt,
		}).
		FirstOrCreate(&DBAuthorizerTokenSqlite{}, DBAuthorizerTokenSqlite{AuthorizerAppID: authorizerAppID}).Error
}

func (s *SqliteStorage) GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error) {
	var dbToken DBAuthorizerTokenSqlite
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

func (s *SqliteStorage) DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error {
	return s.db.Where("authorizer_app_id = ?", authorizerAppID).Delete(&DBAuthorizerTokenSqlite{}).Error
}

func (s *SqliteStorage) ClearAuthorizerTokens(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&DBAuthorizerTokenSqlite{}).Error
}

func (s *SqliteStorage) ListAuthorizerTokens(ctx context.Context) ([]string, error) {
	var tokens []DBAuthorizerTokenSqlite
	if err := s.db.Select("authorizer_app_id").Find(&tokens).Error; err != nil {
		return nil, err
	}
	appids := make([]string, len(tokens))
	for i, token := range tokens {
		appids[i] = token.AuthorizerAppID
	}
	return appids, nil
}

func (s *SqliteStorage) Ping(ctx context.Context) error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}

func (s *SqliteStorage) SavePrevEncodingAESKey(ctx context.Context, appID string, prevKey string) error {
	return s.db.Where(DBPrevEncodingAESKeySqlite{AppID: appID}).
		Assign(DBPrevEncodingAESKeySqlite{PrevEncodingKey: prevKey}).
		FirstOrCreate(&DBPrevEncodingAESKeySqlite{}, DBPrevEncodingAESKeySqlite{AppID: appID}).Error
}

func (s *SqliteStorage) GetPrevEncodingAESKey(ctx context.Context, appID string) (*PrevEncodingAESKey, error) {
	var dbKey DBPrevEncodingAESKeySqlite
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

func (s *SqliteStorage) DeletePrevEncodingAESKey(ctx context.Context, appID string) error {
	return s.db.Where("app_id = ?", appID).Delete(&DBPrevEncodingAESKeySqlite{}).Error
}

func (s *SqliteStorage) SaveComponentVerifyTicket(ctx context.Context, ticket string) error {
	expiresAt := time.Now().Add(12 * time.Hour)
	dbTicket := &DBComponentVerifyTicketSqlite{
		Ticket:    ticket,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&DBComponentVerifyTicketSqlite{}).Error; err != nil {
			return err
		}
		return tx.Create(dbTicket).Error
	})
}

func (s *SqliteStorage) GetComponentVerifyTicket(ctx context.Context) (*ComponentVerifyTicket, error) {
	var dbTicket DBComponentVerifyTicketSqlite
	if err := s.db.Order("created_at DESC").First(&dbTicket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ComponentVerifyTicket{
		Ticket:    dbTicket.Ticket,
		CreatedAt: dbTicket.CreatedAt,
		ExpiresAt: dbTicket.ExpiresAt,
	}, nil
}

func (s *SqliteStorage) DeleteComponentVerifyTicket(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&DBComponentVerifyTicketSqlite{}).Error
}
