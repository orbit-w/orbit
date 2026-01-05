package modules

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"

	"gitee.com/orbit-w/orbit/internal/game/mme_logic/imodels"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic/mechanisms"
)

// HeroModuleLogicImpl HeroModule Logic 实现
type HeroModuleAgentImpl struct {
	wrapper *mme.HeroModuleWrapper

	baseLogic    imodels.IHeroMechanismModel
	levelUpLogic imodels.ILevelUpMechanismLogic
}

// NewHeroModuleLogic 创建 HeroModule Logic
func NewHeroModuleLogic(wrapper *mme.HeroModuleWrapper) imodels.IHeroModuleLogic {
	return &HeroModuleAgentImpl{
		wrapper: wrapper,
	}
}

// GetWrapper 获取 Wrapper
func (m *HeroModuleAgentImpl) GetWrapper() any {
	return m.wrapper
}

// OnLoad 加载回调
func (m *HeroModuleAgentImpl) OnLoad(new bool) error {
	if err := m.GetHeroMechanismModel().OnLoad(new); err != nil {
		return err
	}
	if err := m.GetLevelUpMechanismLogic().OnLoad(new); err != nil {
		return err
	}
	return nil
}

// OnSave 保存回调
func (m *HeroModuleAgentImpl) OnSave() error {
	if err := m.GetHeroMechanismModel().OnSave(); err != nil {
		return err
	}
	if err := m.GetLevelUpMechanismLogic().OnSave(); err != nil {
		return err
	}
	return nil
}

// OnLogin 登录回调
func (m *HeroModuleAgentImpl) OnLogin() error {
	if err := m.GetHeroMechanismModel().OnLogin(); err != nil {
		return err
	}
	if err := m.GetLevelUpMechanismLogic().OnLogin(); err != nil {
		return err
	}
	return nil
}

// OnLogout 登出回调
func (m *HeroModuleAgentImpl) OnLogout() error {
	if err := m.GetHeroMechanismModel().OnLogout(); err != nil {
		return err
	}
	if err := m.GetLevelUpMechanismLogic().OnLogout(); err != nil {
		return err
	}
	return nil
}

func (m *HeroModuleAgentImpl) GetHeroMechanismModel() imodels.IHeroMechanismModel {
	if m.baseLogic == nil {
		m.baseLogic = mechanisms.NewHeroMechanismLogic(m.wrapper.GetBase())
	}
	return m.baseLogic
}

func (m *HeroModuleAgentImpl) GetLevelUpMechanismLogic() imodels.ILevelUpMechanismLogic {
	if m.levelUpLogic == nil {
		m.levelUpLogic = mechanisms.NewLevelUpMechanismLogic(m.wrapper.GetLevelUp())
	}
	return m.levelUpLogic
}
