package manager

import (
	"errors"
	"sync"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

const (
	ManagerName = "actor-manager"
)

var (
	ManagerFacade *ActorFacade
	once          sync.Once
)

func Init() {
	once.Do(func() {
		ManagerFacade = NewActorFacade(actor.NewActorSystem())
		actorsCache = NewActorsCache()
	})
}

// ActorFacade provides a simplified interface for managing actors
type ActorFacade struct {
	actorSystem *actor.ActorSystem
	managers    []*actor.PID
}

// NewActorFacade creates a new instance of ActorFacade
func NewActorFacade(actorSystem *actor.ActorSystem) *ActorFacade {
	af := &ActorFacade{
		actorSystem: actorSystem,
		managers:    make([]*actor.PID, LevelMaxLimit),
	}

	for lv := LevelNormal; lv < LevelMaxLimit; lv++ {
		af.managers[lv] = newManager(actorSystem, lv)
	}

	return af
}

func newManager(actorSystem *actor.ActorSystem, level Level) *actor.PID {
	props := actor.PropsFromProducer(func() actor.Actor {
		return NewActorManager(actorSystem, level)
	})

	managerPID, err := actorSystem.Root.SpawnNamed(props, GenManagerName(level))
	if err != nil {
		panic(err) // In a real application, handle this error appropriately
	}

	return managerPID
}

// GetOrStartActor gets an existing actor or starts a new one
func GetOrStartActor(actorName, pattern string) (*actor.PID, error) {
	// First check if manager already has this actor
	if pid, exists := actorsCache.Get(actorName); exists {
		return pid, nil
	}

	level := GetLevelByPattern(pattern)

	system := ManagerFacade.actorSystem
	future := actor.NewFuture(system, ManagerStartActorFutureTimeout)
	mPid := ManagerFacade.managers[level]
	rf := system.Root.RequestFuture(mPid, &StartActorRequest{
		ActorName: actorName,
		Pattern:   pattern,
		Future:    future.PID(),
	}, StartActorTimeout)

	result, err := waitFuture(rf)
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case *actor.PID:
		return v, nil
	case nil:
		result, err = waitFuture(future)
		if err != nil {
			return nil, err
		}
		return result.(*actor.PID), nil
	default:
		return nil, errors.New("unknown result type")
	}
}

// StopActor stops the actor with the given ID
func StopActor(actorName, pattern string) error {
	result, err := ManagerFacade.RequestFuture(actorName, pattern, &StopActorMessage{
		ActorName: actorName,
	}, StopActorTimeout)
	if err != nil {
		return err
	}

	if err, ok := result.(error); ok {
		return err
	}

	return nil
}

func (f *ActorFacade) RequestFuture(actorName, pattern string, msg any, timeout time.Duration) (any, error) {
	// Send message to manager to start the actor
	level := GetLevelByPattern(pattern)
	mPid := ManagerFacade.managers[level]
	future := f.actorSystem.Root.RequestFuture(mPid, msg, timeout)

	result, err := future.Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func waitFuture(future *actor.Future) (any, error) {
	result, err := future.Result()
	if err != nil {
		return nil, err
	}

	if err, ok := result.(error); ok {
		return nil, err
	}

	return result, nil
}
