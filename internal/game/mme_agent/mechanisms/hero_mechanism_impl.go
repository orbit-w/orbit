package mechanisms

import (
	"time"

	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/imodels"
)

// HeroMechanismLogicImpl HeroMechanism Logic 实现
type HeroMechanismAgent struct {
	wrapper *mme.HeroMechanismWrapper
}

// NewHeroMechanismLogic 创建 HeroMechanism Logic
func NewHeroMechanismLogic(wrapper *mme.HeroMechanismWrapper) imodels.IHeroMechanismModel {
	return &HeroMechanismAgent{
		wrapper: wrapper,
	}
}

// GetWrapper 获取 Wrapper
func (m *HeroMechanismAgent) GetWrapper() any {
	return m.wrapper
}

func (m *HeroMechanismAgent) OnLoad(new bool) error {
	return nil
}

func (m *HeroMechanismAgent) OnSave() error {
	return nil
}

func (m *HeroMechanismAgent) InitHero(confId int32) {
	m.wrapper.SetConfId(confId)
	m.wrapper.SetCreateTime(time.Now().Unix())
	m.wrapper.GetSkillsAccessor().Set(10001, 1)
	m.wrapper.GetSkillsAccessor().Set(10002, 1)
	m.wrapper.GetSkillsAccessor().Set(10003, 1)
	m.wrapper.GetSkillsAccessor().Set(10004, 1)
}

func (m *HeroMechanismAgent) UnlockSkill(confId int32) {

}

// GetId 获取 Id
func (m *HeroMechanismAgent) GetId() int64 {
	return m.wrapper.GetId()
}

// SetId 设置 Id
func (m *HeroMechanismAgent) SetId(id int64) {
	m.wrapper.SetId(id)
}

// GetConfId 获取 ConfId
func (m *HeroMechanismAgent) GetConfId() int32 {
	return m.wrapper.GetConfId()
}

// SetConfId 设置 ConfId
func (m *HeroMechanismAgent) SetConfId(confId int32) {
	m.wrapper.SetConfId(confId)
}

// GetCreateTime 获取 CreateTime
func (m *HeroMechanismAgent) GetCreateTime() int64 {
	return m.wrapper.GetCreateTime()
}

// SetCreateTime 设置 CreateTime
func (m *HeroMechanismAgent) SetCreateTime(createTime int64) {
	m.wrapper.SetCreateTime(createTime)
}

// GetUseTimes 获取 UseTimes
func (m *HeroMechanismAgent) GetUseTimes() int32 {
	return m.wrapper.GetUseTimes()
}

// SetUseTimes 设置 UseTimes
func (m *HeroMechanismAgent) SetUseTimes(useTimes int32) {
	m.wrapper.SetUseTimes(useTimes)
}

// GetSkills 获取 Skills
func (m *HeroMechanismAgent) GetSkills() map[int32]int32 {
	result := make(map[int32]int32)
	m.wrapper.GetSkillsAccessor().Range(func(key int32, value int32) bool {
		result[key] = value
		return true
	})
	return result
}

// SetSkills 设置 Skills
func (m *HeroMechanismAgent) SetSkills(skills map[int32]int32) {
	for key, value := range skills {
		m.wrapper.GetSkillsAccessor().Set(key, value)
	}
}

// 业务方法示例

// AddSkill 添加技能
func (m *HeroMechanismAgent) AddSkill(skillId int32, level int32) {
	m.wrapper.GetSkillsAccessor().Set(skillId, level)
}

// RemoveSkill 移除技能
func (m *HeroMechanismAgent) RemoveSkill(skillId int32) {
	m.wrapper.GetSkillsAccessor().Delete(skillId)
}

// GetSkillLevel 获取技能等级
func (m *HeroMechanismAgent) GetSkillLevel(skillId int32) (int32, bool) {
	return m.wrapper.GetSkillsAccessor().Get(skillId)
}

// IncrementUseTimes 增加使用次数
func (m *HeroMechanismAgent) IncrementUseTimes() {
	m.wrapper.SetUseTimes(m.wrapper.GetUseTimes() + 1)
}
