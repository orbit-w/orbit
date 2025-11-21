package persistence

import (
	"context"
	"sync"

	mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"
	mlog "gitee.com/orbit-w/meteor/modules/mlog"
	"github.com/asynkron/protoactor-go/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	globalPersistence *Persistence
	once              sync.Once
)

// InitSingletonPersistenceWithFile 初始化全局持久化
// 单例模式初始化
// pathConfig: 配置文件路径
// 配置文件格式为yaml或toml
// 配置文件内容为MongoDB连接配置
// 配置文件内容示例:
// mongodb:
//
//	uri: mongodb://localhost:27017
//	database: test
//	collection: test
func InitSingletonPersistenceWithFile(pathConfig string) {
	once.Do(func() {
		cfg, err := mongodbdriver.NewConfigLoader().LoadConfig(pathConfig)
		if err != nil {
			panic(err)
		}
		cli, err := mongodbdriver.NewMongoClient(*cfg)
		if err != nil {
			panic(err)
		}
		globalPersistence = NewPersistence(cli)
		if err := globalPersistence.Start(); err != nil {
			panic(err)
		}
		mlog.Infof("Global persistence initialized")
	})
}

//TODO: 支持配置中心加载配置

// SingletonGracefulStop 单例模式停止持久化系统
func SingletonGracefulStop() error {
	if globalPersistence == nil {
		return ErrDBNotInitialized
	}
	return globalPersistence.GracefulStop()
}

func Load[IDType any](database string, collection string, docID IDType) (*actor.Future, error) {
	if globalPersistence == nil {
		return nil, ErrDBNotInitialized
	}

	req := LoadRequest{
		Database:   database,
		Collection: collection,
		DocID:      docID,
	}

	future, err := globalPersistence.Call(req, DefaultTimeout)
	if err != nil {
		mlog.Errorf("Load failed: %v", err)
		return nil, err
	}
	return future, nil
}

func Persist[IDType any](database string, collection string, documentID IDType, doc bson.M) (*actor.Future, error) {
	if globalPersistence == nil {
		return nil, ErrDBNotInitialized
	}

	req := PersistenceRequest{
		Database:   database,
		Collection: collection,
		DocID:      documentID,
		Doc:        doc,
	}

	future, err := globalPersistence.Call(req, DefaultTimeout)
	if err != nil {
		mlog.Errorf("Persist failed: %v", err)
		return nil, err
	}
	return future, nil
}

// PersistSync 同步持久化数据
func PersistSync[IDType any](ctx context.Context, database string, collection string, documentID IDType, doc bson.M) (*PersistenceResponse, error) {
	if globalPersistence == nil {
		return nil, ErrDBNotInitialized
	}
	req := PersistenceRequest{
		Database:   database,
		Collection: collection,
		DocID:      documentID,
		Doc:        doc,
		Context:    ctx,
	}

	future, err := globalPersistence.Call(req, DefaultTimeout)
	if err != nil {
		mlog.Errorf("Persist failed: %v", err)
		return nil, err
	}

	result, err := future.Result()
	if err != nil {
		mlog.Errorf("Persist failed: %v", err)
		return nil, err
	}

	response, ok := result.(*PersistenceResponse)
	if !ok {
		return nil, ErrInvalidRequest
	}
	return response, nil
}
