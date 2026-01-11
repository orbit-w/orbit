package managers

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic/imodels"

	agent_managers "gitee.com/orbit-w/orbit/internal/game/mme_agent/managers"
)

// HeroManagerLogicImpl HeroManager Logic 实现
type HeroManagerLogicImpl struct {
	agent_managers.IHeroManagerAgent
	wrapper *mme.HeroManagerWrapper
}

// NewHeroManagerLogic 创建 HeroManager Logic
func NewHeroManagerLogic(wrapper *mme.HeroManagerWrapper) imodels.IHeroManagerLogic {
	ins := &HeroManagerLogicImpl{
		wrapper:           wrapper,
		IHeroManagerAgent: agent_managers.NewHeroManagerLogic(wrapper),
	}
	return ins
}
