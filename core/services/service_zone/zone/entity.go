package servicezone

import mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"

type EntityAgent interface {
	mmeobj.IEntity
	OnLoad(new bool) error
	OnSave() error
	GetEntity() mmeobj.IEntity
}
