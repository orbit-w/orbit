package modules

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/imodels"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/mechanisms"
)

// HeroModuleLogicImpl HeroModule Logic 实现
type HeroModuleAgent struct {
	wrapper *mme.HeroModuleWrapper

	baseLogic    imodels.IHeroMechanismModel
	levelUpLogic imodels.ILevelUpMechanismLogic
}

// NewHeroModuleLogic 创建 HeroModule Logic
func NewHeroModuleLogic(wrapper *mme.HeroModuleWrapper) imodels.IHeroModuleLogic {
	return &HeroModuleAgent{
		wrapper: wrapper,
	}
}

// GetWrapper 获取 Wrapper
func (m *HeroModuleAgent) GetWrapper() any {
	return m.wrapper
}

func (m *HeroModuleAgent) GetHeroMechanismModel() imodels.IHeroMechanismModel {
	if m.baseLogic == nil {
		m.baseLogic = mechanisms.NewHeroMechanismLogic(m.wrapper.GetBase())
	}
	return m.baseLogic
}

func (m *HeroModuleAgent) GetLevelUpMechanismLogic() imodels.ILevelUpMechanismLogic {
	if m.levelUpLogic == nil {
		m.levelUpLogic = mechanisms.NewLevelUpMechanismLogic(m.wrapper.GetLevelUp())
	}
	return m.levelUpLogic
}
