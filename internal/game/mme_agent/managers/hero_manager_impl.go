package managers

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/imodels"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/modules"
)

// HeroManagerLogicImpl HeroManager Logic 实现
type HeroManagerAgent struct {
	wrapper *mme.HeroManagerWrapper

	heroModuleLogics      map[int64]imodels.IHeroModuleLogic
	singleHeroModuleLogic imodels.IHeroModuleLogic
}

// NewHeroManagerLogic 创建 HeroManager Logic
func NewHeroManagerLogic(wrapper *mme.HeroManagerWrapper) imodels.IHeroManagerLogic {
	return &HeroManagerAgent{
		wrapper: wrapper,
	}
}

// GetWrapper 获取 Wrapper
func (m *HeroManagerAgent) GetWrapper() any {
	return m.wrapper
}

func (m *HeroManagerAgent) GetHeroMap(key int64) imodels.IHeroModuleLogic {
	if m.heroModuleLogics == nil {
		m.heroModuleLogics = make(map[int64]imodels.IHeroModuleLogic)
	}
	if logic, exists := m.heroModuleLogics[key]; !exists {
		wrapper := m.wrapper.HeroMap_Get(key)
		if wrapper == nil {
			return nil
		}
		logic = modules.NewHeroModuleLogic(wrapper)
		m.heroModuleLogics[key] = logic
	}
	return m.heroModuleLogics[key]
}

func (m *HeroManagerAgent) GetSingleHeroModule() imodels.IHeroModuleLogic {
	if m.singleHeroModuleLogic == nil {
		wrapper := m.wrapper.GetSingleHeroModule()
		if wrapper == nil {
			return nil
		}
		m.singleHeroModuleLogic = modules.NewHeroModuleLogic(wrapper)
	}
	return m.singleHeroModuleLogic
}
