package imodels

type IHeroManagerLogic interface {
	IBaseLogic

	GetSingleHeroModule() IHeroModuleLogic
}
