package imodels

// IWearMechanismLogic WearMechanism Logic 接口
// 业务定义的接口
type IWearMechanismLogic interface {
	IBaseLogic

	GetWearMap() map[int32]int32
	GetConfId() int32

	// 写方法
	Wear(slotId int32, itemId int32) // 在指定槽位穿戴物品
	UnWear(slotId int32)             // 卸下指定槽位的物品
	IsWearing(slotId int32) bool     // 判断指定槽位是否有穿戴物品
}
