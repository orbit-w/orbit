package controller

import (
	"sync"

	"github.com/orbit-w/orbit/app/proto/pb"
)

var (
	GExampleController = &ExampleController{}
)

// Manager 消息分发管理器
type Manager struct {
	pb.CoreRequestHandler
	// 可以扩展添加其他包的处理器
}

var (
	// 全局分发管理器实例
	globalManager *Manager

	// 确保全局实例只初始化一次
	once sync.Once
)

func Init() {
	once.Do(func() {
		globalManager = NewManager()
	})
}

// GetGlobalManager 获取全局分发管理器实例
func GlobalManager() *Manager {
	return globalManager
}

func NewManager() *Manager {
	return &Manager{
		CoreRequestHandler: GExampleController,
	}
}
