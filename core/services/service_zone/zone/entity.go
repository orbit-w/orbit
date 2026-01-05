package servicezone

import (
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EntityAgent interface {
	OnLoad(raw bson.Raw, new bool) error
	OnSave() error
	GetEntity() mmeobj.IEntity
}
