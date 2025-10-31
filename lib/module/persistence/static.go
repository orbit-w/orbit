package persistence

import (
	"context"
	"sync"
	"time"

	mlog "gitee.com/orbit-w/meteor/modules/mlog"
	"github.com/asynkron/protoactor-go/actor"
)

var (
	globalPersistence *Persistence
	once              sync.Once
)

const (
	DefaultTimeout = time.Second * 10
)

func init() {
	once.Do(func() {
		globalPersistence = NewPersistence(nil)
	})
}

func Persist[IDType any](collection string, documentID IDType, wrapper Wrapper) (*actor.Future, error) {
	if globalPersistence == nil {
		return nil, ErrDBNotInitialized
	}

	req := &PersistenceRequest[IDType]{
		Collection: collection,
		DocID:      documentID,
		Wrapper:    wrapper,
	}

	future, err := globalPersistence.Call(req, DefaultTimeout)
	if err != nil {
		mlog.Errorf("Persist failed: %v", err)
		return nil, err
	}
	return future, nil
}

// PersistWithContext 使用自定义上下文持久化数据
func PersistSync[IDType any](ctx context.Context, collection string, documentID IDType, wrapper Wrapper) (*PersistenceResponse, error) {
	req := &PersistenceRequest[IDType]{
		Collection: collection,
		DocID:      documentID,
		Wrapper:    wrapper,
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

// BatchPersist 批量持久化数据
func BatchPersist[IDType any](requests []*PersistenceRequest[IDType]) (*actor.Future, error) {
	req := &BatchPersistenceRequest[IDType]{
		Requests: requests,
	}

	future, err := globalPersistence.Call(req, DefaultTimeout)
	if err != nil {
		mlog.Errorf("Persist failed: %v", err)
		return nil, err
	}
	return future, nil
}

// BatchPersistSync 批量同步持久化数据
func BatchPersistSync[IDType any](requests []*PersistenceRequest[IDType]) (*BatchPersistenceResponse, error) {
	req := &BatchPersistenceRequest[IDType]{
		Requests: requests,
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

	response, ok := result.(*BatchPersistenceResponse)
	if !ok {
		return nil, ErrInvalidRequest
	}
	return response, nil
}
