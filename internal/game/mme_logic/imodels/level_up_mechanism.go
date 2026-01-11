package imodels

// ILevelUpMechanismLogic LevelUpMechanism Logic 接口
// 业务定义的接口
type ILevelUpMechanismLogic interface {
	IBaseLogic

	GetConfId() int32
	GetCurExp() int32
	GetNextLevel() int32

	AddExp(exp int32)
}
