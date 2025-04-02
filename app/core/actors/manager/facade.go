package manager

import (
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
	managerPID  *actor.PID
}

// NewActorFacade creates a new instance of ActorFacade
func NewActorFacade(actorSystem *actor.ActorSystem) *ActorFacade {
	// Create the manager actor
	props := actor.PropsFromProducer(func() actor.Actor {
		return NewActorManager(actorSystem)
	})

	managerPID, err := actorSystem.Root.SpawnNamed(props, "actor-manager")
	if err != nil {
		panic(err) // In a real application, handle this error appropriately
	}

	return &ActorFacade{
		actorSystem: actorSystem,
		managerPID:  managerPID,
	}
}

// GetOrStartActor gets an existing actor or starts a new one
func GetOrStartActor(actorName, pattern string) (*actor.PID, error) {
	// First check if manager already has this actor
	if pid, exists := actorsCache.Get(actorName); exists {
		return pid, nil
	}

	result, err := ManagerFacade.RequestFuture(actorName, &StartActorMessage{
		ActorName: actorName,
		Pattern:   pattern,
	}, StartActorTimeout)
	if err != nil {
		return nil, err
	}

	// Check if result is an error
	if err, ok := result.(error); ok {
		return nil, err
	}

	// Return the PID
	return result.(*actor.PID), nil
}

// StopActor stops the actor with the given ID
func StopActor(actorName string) error {
	result, err := ManagerFacade.RequestFuture(actorName, &StopActorMessage{
		ActorID: actorName,
	}, StopActorTimeout)
	if err != nil {
		return err
	}

	if err, ok := result.(error); ok {
		return err
	}

	return nil
}

func (f *ActorFacade) RequestFuture(actorName string, msg any, timeout time.Duration) (any, error) {
	// Send message to manager to start the actor
	future := f.actorSystem.Root.RequestFuture(f.managerPID, msg, timeout)

	result, err := future.Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}
