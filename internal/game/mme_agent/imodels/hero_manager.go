package imodels

type IHeroManagerLogic interface {
	IBaseLogic

	GetHeroMap(key int64) IHeroModuleLogic
	GetSingleHeroModule() IHeroModuleLogic
}
