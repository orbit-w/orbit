package supervisor

import (
	"github.com/asynkron/protoactor-go/actor"
	cmap "github.com/orbit-w/meteor/bases/container/map/concurrent_map"
)

var (
	cache = cmap.ConcurrentMap[string, *actor.PID]{}
)

func GetActor(id string) *actor.PID {
	pid, ok := cache.Get(id)
	if !ok {
		return nil
	}
	return pid
}

func SetActor(id string, pid *actor.PID) {
	cache.Set(id, pid)
}

func RemoveActor(id string) {
	cache.Remove(id)
}

func RangeActors(callback func(id string, pid *actor.PID)) {
	cache.IterCb(func(key string, value *actor.PID) {
		callback(key, value)
	})
}
