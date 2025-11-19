package storage

import (
	"github.com/jcbowen/jcbaseGo"
)

// StorageClient 存储客户端
type StorageClient struct{}

// NewStorageClient 创建新的存储客户端
func NewStorageClient() *StorageClient {
	return &StorageClient{}
}

// NewDBStorage 创建新的数据库存储实例
func (c *StorageClient) NewDBStorage(dbConfig jcbaseGo.DbStruct) (*DBStorage, error) {
	return NewDBStorage(dbConfig)
}

// NewFileStorage 创建新的文件存储实例
func (c *StorageClient) NewFileStorage(baseDir string) (*FileStorage, error) {
	return NewFileStorage(baseDir)
}

// NewRedisStorage 创建新的Redis存储实例
// @param config *RedisConfig Redis存储配置
// @return *RedisStorage Redis存储实例
// @return error 错误信息
func (c *StorageClient) NewRedisStorage(config *RedisConfig) (*RedisStorage, error) {
	return NewRedisStorage(config)
}
