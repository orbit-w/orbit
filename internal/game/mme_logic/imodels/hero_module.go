package imodels

// IHeroModuleLogic HeroModule Logic 接口
// 业务定义的接口
type IHeroModuleLogic interface {
	IBaseLogic

	GetHeroMechanismModel() IHeroMechanismModel
	GetLevelUpMechanismLogic() ILevelUpMechanismLogic
}
