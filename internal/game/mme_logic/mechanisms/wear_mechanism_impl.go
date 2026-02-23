package mechanisms

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic/imodels"
)

// WearMechanismLogicImpl WearMechanism Logic 实现
type WearMechanismLogicImpl struct {
	wrapper *mme.WearMechanismWrapper
}

// NewWearMechanismLogic 创建 WearMechanism Logic
func NewWearMechanismLogic(wrapper *mme.WearMechanismWrapper) imodels.IWearMechanismLogic {
	return &WearMechanismLogicImpl{
		wrapper: wrapper,
	}
}

// GetWrapper 获取 Wrapper
func (m *WearMechanismLogicImpl) GetWrapper() any {
	return m.wrapper
}

// OnLoad 加载回调
func (m *WearMechanismLogicImpl) OnLoad(new bool) error {
	return nil
}

// OnSave 保存回调
func (m *WearMechanismLogicImpl) OnSave() error {
	return nil
}

// OnLogin 登录回调
func (m *WearMechanismLogicImpl) OnLogin() error {
	return nil
}

// OnLogout 登出回调
func (m *WearMechanismLogicImpl) OnLogout() error {
	return nil
}

// GetWearMap 获取穿戴 Map（槽位 -> 物品 id）
func (m *WearMechanismLogicImpl) GetWearMap() map[int32]int32 {
	result := make(map[int32]int32)
	m.wrapper.GetWearMapAccessor().Range(func(key int32, value int32) bool {
		result[key] = value
		return true
	})
	return result
}

// GetConfId 获取 ConfId
func (m *WearMechanismLogicImpl) GetConfId() int32 {
	return m.wrapper.GetConfId()
}

// Wear 在指定槽位穿戴物品
func (m *WearMechanismLogicImpl) Wear(slotId int32, itemId int32) {
	m.wrapper.GetWearMapAccessor().Set(slotId, itemId)
}

// UnWear 卸下指定槽位的物品
func (m *WearMechanismLogicImpl) UnWear(slotId int32) {
	m.wrapper.GetWearMapAccessor().Delete(slotId)
}

// IsWearing 判断指定槽位是否有穿戴物品
func (m *WearMechanismLogicImpl) IsWearing(slotId int32) bool {
	_, ok := m.wrapper.GetWearMapAccessor().Get(slotId)
	return ok
}
