package managers

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/imodels"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/modules"
)

// HeroManagerLogicImpl HeroManager Logic 实现
type HeroManagerAgentImpl struct {
	wrapper *mme.HeroManagerWrapper

	heroModuleLogics *ModuleMapContainer[int64, *mme.HeroModule, *mme.HeroModuleWrapper, imodels.IHeroModuleLogic]
	singleHeroModule imodels.IHeroModuleLogic
}

// NewHeroManagerLogic 创建 HeroManager Logic
func NewHeroManagerLogic(wrapper *mme.HeroManagerWrapper) imodels.IHeroManagerLogic {
	ins := &HeroManagerAgentImpl{
		wrapper: wrapper,
		heroModuleLogics: NewModuleMapContainer(
			wrapper.GetHeroMap(),
			modules.NewHeroModuleLogic,
		),
	}

	return ins
}

// GetWrapper 获取 Wrapper
func (m *HeroManagerAgentImpl) GetWrapper() any {
	return m.wrapper
}

func (m *HeroManagerAgentImpl) OnLoad(new bool) error {
	var err error
	m.heroModuleLogics.Range(func(key int64, logic imodels.IHeroModuleLogic) bool {
		if err = logic.OnLoad(new); err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return err
	}

	if err = m.GetSingleHeroModule().OnLoad(new); err != nil {
		return err
	}
	return nil
}

func (m *HeroManagerAgentImpl) OnSave() error {
	var err error
	m.heroModuleLogics.Range(func(key int64, logic imodels.IHeroModuleLogic) bool {
		if err = logic.OnSave(); err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return err
	}
	if err = m.GetSingleHeroModule().OnSave(); err != nil {
		return err
	}
	return nil
}

func (m *HeroManagerAgentImpl) HeroMap_Get(heroId int64) (imodels.IHeroModuleLogic, bool) {
	logic, exists := m.heroModuleLogics.Get(heroId)
	return logic, exists
}

func (m *HeroManagerAgentImpl) GetSingleHeroModule() imodels.IHeroModuleLogic {
	if m.singleHeroModule == nil {
		wrapper := m.wrapper.GetSingleHeroModule()
		if wrapper == nil {
			return nil
		}
		m.singleHeroModule = modules.NewHeroModuleLogic(wrapper)
	}
	return m.singleHeroModule
}
