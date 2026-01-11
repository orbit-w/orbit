package servicezone_behavior

import (
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	"gitee.com/orbit-w/orbit/pkg/proto/core"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (ab *ZoneActorBehavior) HandleSystemRequest_SetEntity(ctx actor.Context, clientRequest *ClientRequest) {
	var err error
	defer func() {
		if err != nil {
			ab.logger.Error("ZoneActor error", zap.Error(err), zap.Uint32("Pid", clientRequest.GetPid()))
			ResponseError(clientRequest, err.Error())
		} else {
			ResponseOK(clientRequest)
		}
	}()

	req := new(core.Request_SetEntityRequest)
	if err = proto.Unmarshal(clientRequest.GetIn(), req); err != nil {
		ab.logger.Error("ZoneActor unmarshal error", zap.Error(err), zap.Uint32("Pid", clientRequest.GetPid()))
		return
	}

	ref := req.GetEntityRef()
	if ref == nil {
		ab.logger.Error("ZoneActor received unknown message", zap.Uint32("Pid", clientRequest.GetPid()))
		err = ErrEntityRefNil
		return
	}

	entity, err := ab.ctx.Load(ref.GetEntityId(), mme.EntityType(ref.GetEntityType()))
	if err != nil {
		ab.logger.Error("ZoneActor load error", zap.Error(err), zap.Uint32("Pid", clientRequest.GetPid()))
		err = ErrEntityLoadFailed
		return
	}

	pbFactory := mmeobj.GetEntityDataFactory(mme.EntityType(ref.GetEntityType()))
	if pbFactory == nil {
		ab.logger.Error("ZoneActor received unknown message", zap.Uint32("Pid", clientRequest.GetPid()))
		err = ErrEntityDataFactoryNotFound
		return
	}
	pb := pbFactory()
	if err = proto.Unmarshal(req.GetData(), pb); err != nil {
		ab.logger.Error("ZoneActor unmarshal error", zap.Error(err), zap.Uint32("Pid", clientRequest.GetPid()))
		err = ErrEntityDataUnmarshalFailed
		return
	}

	entity.GetEntityWrapper().FromProto(pb)

	//ab.ctx.SetEntity(entity)

	ab.Persist([]mme_agent.IEntity{entity})
}
