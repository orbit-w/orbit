package imodels

// IManualUnlockMechanismLogic ManualUnlockMechanism Logic 接口
// 业务定义的接口
type IManualUnlockMechanismLogic interface {
	IBaseLogic

	GetUnlockMap() map[int32]bool

	// 写方法
	Unlock(id int32)            // 解锁指定 id
	IsUnlocked(id int32) bool   // 判断是否已解锁
}
