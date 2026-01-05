package imodels

type IHeroManagerLogic interface {
	IBaseLogic

	GetSingleHeroModule() IHeroModuleLogic
	HeroMap_GetModule(heroId int64) (IHeroModuleLogic, bool)
}
