package imodels

type IHeroModuleLogic interface {
	IBaseLogic

	GetHeroMechanismModel() IHeroMechanismModel
	GetLevelUpMechanismLogic() ILevelUpMechanismLogic
}
