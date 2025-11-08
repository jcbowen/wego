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

// NewMemoryStorage 创建新的内存存储实例
func (c *StorageClient) NewMemoryStorage() *MemoryStorage {
	return NewMemoryStorage()
}

// NewDBStorage 创建新的数据库存储实例
func (c *StorageClient) NewDBStorage(dbConfig jcbaseGo.DbStruct) (*DBStorage, error) {
	return NewDBStorage(dbConfig)
}

// NewFileStorage 创建新的文件存储实例
func (c *StorageClient) NewFileStorage(baseDir string) (*FileStorage, error) {
	return NewFileStorage(baseDir)
}