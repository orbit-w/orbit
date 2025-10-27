package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"google.golang.org/protobuf/proto"
)

// Dirty bits for Module fields
const (
	HeroModuleDirtyBaseBit    int64 = 1 << 0
	HeroModuleDirtyLevelUpBit int64 = 1 << 1
)

type HeroModule struct {
	mme *mme.HeroModule
	dirtyflag.IDirtyFlag

	Base    *HeroMechanism
	LevelUp *LevelUpMechanism
}

func NewHeroModule(pt *mme.HeroModule) *HeroModule {
	if pt == nil {
		panic("pt is nil")
	}
	m := &HeroModule{
		mme:        pt,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
	}
	// Mechanism obj 需要Link到DirtyTracker
	// xmap类型，
	// if Value 是Mechanism，则需要通过LinkFactsAccessor到Father和xmap的Key

	m.Base = NewHeroMechanism(m.mme.Base)
	m.Base.Link(m.GetDirtyTracker(), HeroModuleDirtyBaseBit)

	m.LevelUp = NewLevelUpMechanism(m.mme.LevelUp)
	m.LevelUp.Link(m.GetDirtyTracker(), HeroModuleDirtyLevelUpBit)
	return m
}

func (m *HeroModule) Name() string {
	return "HeroModule"
}

func (m *HeroModule) DeepCopy() *mme.HeroModule {
	if m == nil {
		return nil
	}
	return proto.Clone(m.mme).(*mme.HeroModule)
}

func (m *HeroModule) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if m == nil {
		return
	}

	if m.Base != nil && m.IsDirty(HeroModuleDirtyBaseBit) {
		m.Base.BuildMongoUpdate(builder, prefix+".base")
	}

	if m.LevelUp != nil && m.IsDirty(HeroModuleDirtyLevelUpBit) {
		m.LevelUp.BuildMongoUpdate(builder, prefix+".level_up")
	}
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroModule) ToIncrementalProto() proto.Message {
	if m == nil {
		return nil
	}

	incremental := &mme.HeroModule{}

	if m.Base != nil && m.IsDirty(HeroModuleDirtyBaseBit) {
		pb := m.Base.ToIncrementalProto()
		v, ok := pb.(*mme.HeroMechanism)
		if ok {
			incremental.Base = v
		}
	}

	if m.LevelUp != nil && m.IsDirty(HeroModuleDirtyLevelUpBit) {
		pb := m.LevelUp.ToIncrementalProto()
		v, ok := pb.(*mme.LevelUpMechanism)
		if ok {
			incremental.LevelUp = v
		}
	}

	return incremental
}
