package mme_agent

import "gitee.com/orbit-w/orbit/internal/game/mme_logic/imodels"

// IBaseAgent Base Agent 接口
// 业务定义的接口
type IBaseAgent interface {
	GetWrapper() any // 获取 Wrapper
}

// IHeroManagerAgent HeroManager Logic 接口
// 业务定义的接口
type IHeroManagerAgent interface {
	IBaseAgent
	imodels.IHeroManagerLogic // 继承业务层接口
}

// IHeroModuleAgent HeroModule Agent 接口
// 业务定义的接口
type IHeroModuleAgent interface {
	IBaseAgent
	GetHeroMechanismModel() imodels.IHeroMechanismLogic
	GetLevelUpMechanismLogic() imodels.ILevelUpMechanismLogic
}
