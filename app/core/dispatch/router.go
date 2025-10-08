package dispatch

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

type Router struct {
	funcMap map[uint32]func(data []byte) (proto.Message, uint32, error)
}

func NewRouter() *Router {
	return &Router{
		funcMap: make(map[uint32]func(data []byte) (proto.Message, uint32, error)),
	}
}

func (r *Router) Register(pid uint32, callback func(data []byte) (proto.Message, uint32, error)) {
	if _, ok := r.funcMap[pid]; ok {
		panic(fmt.Sprintf("pid %d already registered", pid))
	}

	r.funcMap[pid] = callback
}

func (r *Router) Dispatch(pid uint32, data []byte) (proto.Message, uint32, error) {
	callback, ok := r.funcMap[pid]
	if !ok {
		return nil, 0, fmt.Errorf("no callback found for pid %d", pid)
	}

	return callback(data)
}
