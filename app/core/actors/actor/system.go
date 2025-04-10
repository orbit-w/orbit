package actor

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

const (
	ManagerName = "system-actor-supervision"
)

var (
	System              *ActorSystem
	once                sync.Once
	ActorFacadeStopping atomic.Bool
)

func Init() {
	once.Do(func() {
		actorsCache = NewActorsCache()
	})
}

// ActorSystem provides a simplified interface for managing actors
type ActorSystem struct {
	actorSystem *actor.ActorSystem
	supervisors []*actor.PID
}

func (af *ActorSystem) Start() error {
	system := actor.NewActorSystem()
	af.actorSystem = system
	af.supervisors = make([]*actor.PID, LevelMaxLimit)
	for lv := LevelNormal; lv < LevelMaxLimit; lv++ {
		af.supervisors[lv] = newSupervisor(system, lv)
	}
	System = af
	return nil
}

func (af *ActorSystem) Stop() error {
	ActorFacadeStopping.Store(true)
	for lv := LevelNormal; lv < LevelMaxLimit; {
		completed, err := af.stopSupervisor(lv)
		if err != nil {
			return err
		}
		if completed {
			lv++
		}
	}
	return nil
}

func (af *ActorSystem) ActorSystem() *actor.ActorSystem {
	return af.actorSystem
}

func (af *ActorSystem) stopSupervisor(lv Level) (bool, error) {
	result, err := retry(func() (any, error) {
		future := af.actorSystem.Root.RequestFuture(af.supervisors[lv], &StopAllRequest{}, 30*time.Second)
		result, err := future.Result()
		return result, err
	}, 10)
	if err != nil {
		return false, err
	}

	resp := result.(*StopAllResponse)
	return resp.Complete, nil
}

// NewActorFacade creates a new instance of ActorFacade
func NewActorFacade(actorSystem *actor.ActorSystem) *ActorSystem {
	af := &ActorSystem{
		actorSystem: actorSystem,
		supervisors: make([]*actor.PID, LevelMaxLimit),
	}

	for lv := LevelNormal; lv < LevelMaxLimit; lv++ {
		af.supervisors[lv] = newSupervisor(actorSystem, lv)
	}

	return af
}

func newSupervisor(actorSystem *actor.ActorSystem, level Level) *actor.PID {
	decider := func(reason any) actor.Directive {
		return actor.ResumeDirective
	}
	supervisor := actor.NewOneForOneStrategy(10, 1000, decider)
	producer := func() actor.Actor {
		return NewActorSupervision(actorSystem, level)
	}
	props := actor.PropsFromProducer(producer, actor.WithSupervisor(supervisor))

	managerPID, err := actorSystem.Root.SpawnNamed(props, GenManagerName(level))
	if err != nil {
		panic(err) // In a real application, handle this error appropriately
	}

	return managerPID
}

// GetOrStartActor 获取一个就绪的Actor对象
func GetOrStartActor(actorName, pattern string) (*actor.PID, error) {
	// First check if manager already has this actor
	if pid, exists := actorsCache.Get(actorName); exists {
		return pid, nil
	}

	system := System.actorSystem
	future := actor.NewFuture(system, ManagerStartActorFutureTimeout)
	mPid := System.supervisorByPattern(pattern)
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
	case *StartActorWait:
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
	result, err := System.RequestFuture(actorName, pattern, &StopActorMessage{
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

func (f *ActorSystem) RequestFuture(actorName, pattern string, msg any, timeout time.Duration) (any, error) {
	// Send message to manager to start the actor
	mPid := f.supervisorByPattern(pattern)
	future := f.actorSystem.Root.RequestFuture(mPid, msg, timeout)

	result, err := future.Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (f *ActorSystem) supervisorByPattern(pattern string) *actor.PID {
	level := GetLevelByPattern(pattern)
	return f.supervisorByLevel(level)
}

func (f *ActorSystem) supervisorByLevel(level Level) *actor.PID {
	return f.supervisors[level]
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

func retry(fn func() (any, error), retryCount int) (any, error) {
	var lastErr error
	for i := 0; i < retryCount; i++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, errors.New("max retry attempts reached: " + lastErr.Error())
}
