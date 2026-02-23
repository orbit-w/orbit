package gravitas

import (
	entityloader "gitee.com/orbit-w/orbit/core/system/gravitas/entity_loader"
	entitymgr "gitee.com/orbit-w/orbit/core/system/gravitas/entity_mgr"
	"gitee.com/orbit-w/orbit/core/system/gravitas/worker"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
)

var inst *GravitasImpl

// Start 初始化并启动 Gravitas 单例。
//
// 进程生命周期内只应调用一次；重复调用前必须先调用 Stop。
func Start(cfg Config, handler worker.MessageHandler) error {
	g := &GravitasImpl{
		cfg:     cfg,
		handler: handler,
	}
	if err := g.init(); err != nil {
		return err
	}
	inst = g
	return nil
}

// Stop 停止并重置 Gravitas 单例。
func Stop() error {
	if inst == nil {
		return nil
	}
	err := inst.Stop()
	inst = nil
	return err
}

// GetInst 返回 Gravitas 单例实例。
func GetInst() *GravitasImpl {
	return inst
}

// Dispatch 将消息分发到对应的 Worker（单例转发）。
func Dispatch(refs []entityloader.EntityRef) {
	inst.Dispatch(refs)
}

// DispatchFromProto 从 proto EntityRef 列表分发消息（单例转发）。
func DispatchFromProto(refs []*mme.EntityRef) {
	inst.DispatchFromProto(refs)
}

// GetEntityManager 返回单例的 EntityManager（单例转发）。
func GetEntityManager() *entitymgr.EntityManager {
	return inst.EntityManager()
}
