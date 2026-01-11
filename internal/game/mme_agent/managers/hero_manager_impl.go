package managers

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic/imodels"
	"gitee.com/orbit-w/orbit/lib/container"

	"gitee.com/orbit-w/orbit/internal/game/mme_agent/modules"
)

type IHeroManagerAgent interface {
	GetWrapper() any
	HeroMap_GetModule(heroId int64) (imodels.IHeroModuleLogic, bool)
	GetSingleHeroModule() imodels.IHeroModuleLogic
}

type HeroManagerAgentImpl struct {
	wrapper *mme.HeroManagerWrapper

	heroModuleLogics *container.ModuleMapContainer[int64, *mme.HeroModule, *mme.HeroModuleWrapper, imodels.IHeroModuleLogic]
	singleHeroModule imodels.IHeroModuleLogic
}

// NewHeroManagerLogic 创建 HeroManager Logic
func NewHeroManagerLogic(wrapper *mme.HeroManagerWrapper) IHeroManagerAgent {
	ins := &HeroManagerAgentImpl{
		wrapper: wrapper,
		heroModuleLogics: container.NewModuleMapContainer(
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
		if lifecycle, ok := logic.(interface{ OnLoad(new bool) error }); ok {
			if err = lifecycle.OnLoad(new); err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}

	if lifecycle, ok := m.GetSingleHeroModule().(interface{ OnLoad(new bool) error }); ok {
		if err = lifecycle.OnLoad(new); err != nil {
			return err
		}
	}
	return nil
}

func (m *HeroManagerAgentImpl) OnSave() error {
	var err error
	m.heroModuleLogics.Range(func(key int64, logic imodels.IHeroModuleLogic) bool {
		if lifecycle, ok := logic.(interface{ OnSave() error }); ok {
			if err = lifecycle.OnSave(); err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}
	if lifecycle, ok := m.GetSingleHeroModule().(interface{ OnSave() error }); ok {
		if err = lifecycle.OnSave(); err != nil {
			return err
		}
	}
	return nil
}

func (m *HeroManagerAgentImpl) OnLogin() error {
	var err error
	m.heroModuleLogics.Range(func(key int64, logic imodels.IHeroModuleLogic) bool {
		if lifecycle, ok := logic.(interface{ OnLogin() error }); ok {
			if err = lifecycle.OnLogin(); err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}

	if lifecycle, ok := m.GetSingleHeroModule().(interface{ OnLogin() error }); ok {
		if err = lifecycle.OnLogin(); err != nil {
			return err
		}
	}
	return nil
}

func (m *HeroManagerAgentImpl) OnLogout() error {
	var err error
	m.heroModuleLogics.Range(func(key int64, logic imodels.IHeroModuleLogic) bool {
		if lifecycle, ok := logic.(interface{ OnLogout() error }); ok {
			if err = lifecycle.OnLogout(); err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}

	if lifecycle, ok := m.GetSingleHeroModule().(interface{ OnLogout() error }); ok {
		if err = lifecycle.OnLogout(); err != nil {
			return err
		}
	}
	return nil
}
func (m *HeroManagerAgentImpl) HeroMap_GetModule(heroId int64) (imodels.IHeroModuleLogic, bool) {
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
