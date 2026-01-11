package managers

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic/imodels"
)

// HeroManagerLogicImpl HeroManager Logic 实现
type HeroManagerLogicImpl struct {
	wrapper *mme.HeroManagerWrapper
}

// NewHeroManagerLogic 创建 HeroManager Logic
func NewHeroManagerLogic(wrapper *mme.HeroManagerWrapper) imodels.IHeroManagerLogic {
	ins := &HeroManagerLogicImpl{
		wrapper: wrapper,
	}
	return ins
}
