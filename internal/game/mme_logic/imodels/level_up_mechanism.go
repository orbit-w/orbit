package imodels

type ILevelUpMechanismLogic interface {
	IBaseLogic

	GetConfId() int32
	GetCurExp() int32
	GetNextLevel() int32

	AddExp(exp int32)
}
