package dispatch

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

type Router struct {
	funcMap map[uint32]func(data []byte) (proto.Message, string, error)
}

func NewRouter() *Router {
	return &Router{
		funcMap: make(map[uint32]func(data []byte) (proto.Message, string, error)),
	}
}

func (r *Router) Register(pid uint32, router func(data []byte) (proto.Message, string, error)) {
	if _, ok := r.funcMap[pid]; ok {
		panic(fmt.Sprintf("pid %d already registered", pid))
	}

	r.funcMap[pid] = router
}

func (r *Router) Dispatch(pid uint32, data []byte) (proto.Message, string, error) {
	router, ok := r.funcMap[pid]
	if !ok {
		return nil, "", fmt.Errorf("no callback found for pid %d", pid)
	}

	return router(data)
}
