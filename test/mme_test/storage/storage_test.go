package storage

import (
	"context"
	"testing"
	"time"

	mme "gitee.com/orbit-w/orbit/app/mme_v2"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	"github.com/stretchr/testify/assert"
)

func Test_Storage(t *testing.T) {
	persistence.InitSingletonPersistenceWithFile("mongodb.toml")

	data := &mme.HeroManager{
		HeroMap: make(map[int64]*mme.HeroModule),
	}
	accessor := mme.NewHeroManagerWrapper(data)
	_ = accessor.HeroMap_Set(1, &mme.HeroModule{
		Base: &mme.HeroMechanism{
			Id:         2,
			CreateTime: time.Now().Unix(),
			UseTimes:   1,
			Skills:     make(map[int32]int32),
		},
		LevelUp: &mme.LevelUpMechanism{
			CurLevel: 1,
			CurExp:   0,
			ConfId:   10011,
		},
	})

	builder := mgo_builder.NewMongoUpdateBuilder()
	path := mgo_builder.NewNestedPath()
	accessor.BuildMongoUpdate(builder, path.Field("hero_manager"))

	persistence.PersistSync(context.Background(), "test", "player", 1, builder.Build())

	accessor.HeroMap_Set(2, &mme.HeroModule{
		Base: &mme.HeroMechanism{
			Id:         150001,
			CreateTime: time.Now().Unix(),
			UseTimes:   1,
			Skills:     make(map[int32]int32),
		},
		LevelUp: &mme.LevelUpMechanism{
			CurLevel: 2,
			CurExp:   500,
			ConfId:   10012,
		},
	})

	builder = mgo_builder.NewMongoUpdateBuilder()
	path = mgo_builder.NewNestedPath()
	accessor.BuildMongoUpdate(builder, path.Field("hero_manager"))

	persistence.PersistSync(context.Background(), "test", "player", 1, builder.Build())

	err := persistence.SingletonGracefulStop()
	assert.NoError(t, err)
}
