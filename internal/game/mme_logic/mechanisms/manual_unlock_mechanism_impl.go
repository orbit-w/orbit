package mechanisms

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic/imodels"
)

// ManualUnlockMechanismLogicImpl ManualUnlockMechanism Logic 实现
type ManualUnlockMechanismLogicImpl struct {
	wrapper *mme.ManualUnlockMechanismWrapper
}

// NewManualUnlockMechanismLogic 创建 ManualUnlockMechanism Logic
func NewManualUnlockMechanismLogic(wrapper *mme.ManualUnlockMechanismWrapper) imodels.IManualUnlockMechanismLogic {
	return &ManualUnlockMechanismLogicImpl{
		wrapper: wrapper,
	}
}

// GetWrapper 获取 Wrapper
func (m *ManualUnlockMechanismLogicImpl) GetWrapper() any {
	return m.wrapper
}

// OnLoad 加载回调
func (m *ManualUnlockMechanismLogicImpl) OnLoad(new bool) error {
	return nil
}

// OnSave 保存回调
func (m *ManualUnlockMechanismLogicImpl) OnSave() error {
	return nil
}

// OnLogin 登录回调
func (m *ManualUnlockMechanismLogicImpl) OnLogin() error {
	return nil
}

// OnLogout 登出回调
func (m *ManualUnlockMechanismLogicImpl) OnLogout() error {
	return nil
}

// GetUnlockMap 获取解锁 Map
func (m *ManualUnlockMechanismLogicImpl) GetUnlockMap() map[int32]bool {
	result := make(map[int32]bool)
	m.wrapper.GetUnlockMapAccessor().Range(func(key int32, value bool) bool {
		result[key] = value
		return true
	})
	return result
}

// Unlock 解锁指定 id
func (m *ManualUnlockMechanismLogicImpl) Unlock(id int32) {
	m.wrapper.GetUnlockMapAccessor().Set(id, true)
}

// IsUnlocked 判断是否已解锁
func (m *ManualUnlockMechanismLogicImpl) IsUnlocked(id int32) bool {
	v, ok := m.wrapper.GetUnlockMapAccessor().Get(id)
	return ok && v
}
