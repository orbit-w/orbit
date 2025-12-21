package mmeobj

import (
	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/pb_gen/net_message"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

type Mechanism struct {
	*MMEObject
	Requests []*net_message.NetMessage
	Notifies []*net_message.NetMessage
}

func NewMechanism() *Mechanism {
	return &Mechanism{
		MMEObject: NewMMEObject(mmeobject.ObjectTypeMechanism),
		Requests:  make([]*net_message.NetMessage, 0),
		Notifies:  make([]*net_message.NetMessage, 0),
	}
}
