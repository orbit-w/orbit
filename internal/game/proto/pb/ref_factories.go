package pb

import (
	"fmt"

	"gitee.com/orbit-w/orbit/internal/game/proto/core"
	"gitee.com/orbit-w/orbit/internal/game/proto/mme"
	"google.golang.org/protobuf/proto"
)

type ExtractRefsFromRequest func(req proto.Message) ([]*mme.EntityRef, error)

var (
	extractRefsFromRequests = make(map[uint32]ExtractRefsFromRequest)
)

func init() {
	RegisterRefExtract(PID_Request_LoginRequest, func(request proto.Message) ([]*mme.EntityRef, error) {
		req := request.(*core.Request_LoginRequest)
		refs := make([]*mme.EntityRef, 0)
		if req.PlayerEntityRef != nil {
			refs = append(refs, req.PlayerEntityRef)
		}
		return refs, nil
	})

	RegisterRefExtract(PID_Request_AskLevelUp, func(request proto.Message) ([]*mme.EntityRef, error) {
		return nil, nil
	})
}

func RegisterRefExtract(pid uint32, factory ExtractRefsFromRequest) {
	if _, ok := extractRefsFromRequests[pid]; ok {
		panic(fmt.Sprintf("ref factory already registered for pid: %d", pid))
	}
	extractRefsFromRequests[pid] = factory
}

// 获取Ref提取器
func GetExtract(pid uint32) ExtractRefsFromRequest {
	return extractRefsFromRequests[pid]
}
