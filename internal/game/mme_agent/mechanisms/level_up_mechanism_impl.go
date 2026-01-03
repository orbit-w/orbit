package mechanisms

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/imodels"
)

// LevelUpMechanismLogicImpl LevelUpMechanism Logic 实现
type LevelUpMechanismLogicImpl struct {
	wrapper *mme.LevelUpMechanismWrapper
}

// NewLevelUpMechanismLogic 创建 LevelUpMechanism Logic
func NewLevelUpMechanismLogic(wrapper *mme.LevelUpMechanismWrapper) imodels.ILevelUpMechanismLogic {
	return &LevelUpMechanismLogicImpl{
		wrapper: wrapper,
	}
}

// GetWrapper 获取 Wrapper
func (m *LevelUpMechanismLogicImpl) GetWrapper() any {
	return m.wrapper
}

// GetCurLevel 获取 CurLevel
func (m *LevelUpMechanismLogicImpl) GetCurLevel() int32 {
	return m.wrapper.GetCurLevel()
}

// SetCurLevel 设置 CurLevel
func (m *LevelUpMechanismLogicImpl) SetCurLevel(curLevel int32) {
	m.wrapper.SetCurLevel(curLevel)
}

// GetCurExp 获取 CurExp
func (m *LevelUpMechanismLogicImpl) GetCurExp() int32 {
	return m.wrapper.GetCurExp()
}

// SetCurExp 设置 CurExp
func (m *LevelUpMechanismLogicImpl) SetCurExp(curExp int32) {
	m.wrapper.SetCurExp(curExp)
}

// GetConfId 获取 ConfId
func (m *LevelUpMechanismLogicImpl) GetConfId() int32 {
	return m.wrapper.GetConfId()
}

// SetConfId 设置 ConfId
func (m *LevelUpMechanismLogicImpl) SetConfId(confId int32) {
	m.wrapper.SetConfId(confId)
}

// 业务方法

// GetNextLevel 获取下一级
func (m *LevelUpMechanismLogicImpl) GetNextLevel() int32 {
	return m.wrapper.GetCurLevel() + 1
}

// AddExp 添加经验
func (m *LevelUpMechanismLogicImpl) AddExp(exp int32) {
	curExp := m.wrapper.GetCurExp()
	curExp += exp
	m.wrapper.SetCurExp(curExp)

	// 如果经验值大于等于当前等级配置的升级经验值，则升级
	// if curExp >= m.wrapper.GetCurLevelConfig().Exp {
	// 	m.LevelUp()
	// }
}

// LevelUp 升级
func (m *LevelUpMechanismLogicImpl) LevelUp() {
	curLevel := m.wrapper.GetCurLevel()
	curLevel++
	m.wrapper.SetCurLevel(curLevel)
}
