package imodels

// IHeroMechanismModel HeroMechanism Model 接口
// 业务定义的接口
type IHeroMechanismModel interface {
	IBaseLogic

	GetId() int64
	GetUseTimes() int32
	GetSkills() map[int32]int32
	GetConfId() int32
	GetCreateTime() int64

	// 写方法
	InitHero(confId int32)          // 初始化英雄
	SetCreateTime(createTime int64) // 设置创建时间
	SetUseTimes(useTimes int32)     // 设置使用次数
	UnlockSkill(confId int32)       // 解锁技能
}
