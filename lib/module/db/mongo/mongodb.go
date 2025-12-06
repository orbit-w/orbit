package mongo

import (
	"context"
	"errors"
	"sync"

	mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	globalMongoDBManager *MongoDBManager
	once                 sync.Once
	initErr              error
)

// VirtualClient 获取全局 MongoDB 客户端
func VirtualClient() *mongodbdriver.VirtualMongoClient {
	if globalMongoDBManager == nil {
		return nil
	}
	return globalMongoDBManager.GetClient()
}

// MongoDBManager MongoDB 管理器，单例模式
type MongoDBManager struct {
	client *mongodbdriver.VirtualMongoClient
	rw     sync.RWMutex
	dbMap  map[string]*mongo.Database
}

// Start 启动 MongoDB 管理器（全局单例）
// 多次调用只会初始化一次，后续调用会返回首次初始化的结果
func Start(cfg *mongodbdriver.MongoDBConfig) error {
	once.Do(func() {
		cli, err := mongodbdriver.NewMongoClient(*cfg)
		if err != nil {
			initErr = err
			return
		}
		globalMongoDBManager = &MongoDBManager{
			client: cli,
			dbMap:  make(map[string]*mongo.Database),
		}
	})
	return initErr
}

// Stop 停止 MongoDB 管理器（全局单例）
func Stop() error {
	if globalMongoDBManager == nil {
		return errors.New("mongodb not started")
	}
	err := globalMongoDBManager.client.Disconnect(context.Background())
	if err != nil {
		return err
	}
	// 清理全局单例
	globalMongoDBManager = nil
	return nil
}

// GetManager 获取全局 MongoDB 管理器实例
// 如果未初始化，返回 nil
func GetManager() *MongoDBManager {
	return globalMongoDBManager
}

func (m *MongoDBManager) GetClient() *mongodbdriver.VirtualMongoClient {
	return m.client
}

func (m *MongoDBManager) GetDatabase(name string) *mongo.Database {
	// 第一次检查：使用读锁快速路径
	m.rw.RLock()
	db, ok := m.dbMap[name]
	m.rw.RUnlock()
	if ok {
		return db
	}

	// 第二次检查：使用写锁，确保只创建一次（Double-Check Locking）
	m.rw.Lock()
	defer m.rw.Unlock()
	// 再次检查，可能在等待写锁期间其他 goroutine 已经创建了
	db, ok = m.dbMap[name]
	if ok {
		return db
	}
	// 创建新的 database 实例
	db = m.client.Database(name)
	m.dbMap[name] = db
	return db
}
