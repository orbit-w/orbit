package supervisor

import (
	"github.com/asynkron/protoactor-go/actor"
)

// Factory function type that accepts an actor ID
type ActorFactory func(actorName string) actor.Actor

var factories = make(map[string]ActorFactory)

// RegFactory registers a factory for a specific actor type
func RegFactory(name string, factory ActorFactory) {
	if _, ok := factories[name]; ok {
		panic("factory already registered: " + name)
	}
	factories[name] = factory
}

// RegFactories registers multiple factories at once
func RegFactories(factoryMap map[string]ActorFactory) {
	for name, factory := range factoryMap {
		RegFactory(name, factory)
	}
}

// Dispatch returns a factory function for creating actors with the given ID
func Dispatch(name string) ActorFactory {
	factory, ok := factories[name]
	if !ok {
		panic("factory not found: " + name)
	}

	return factory
}

// CreateActorWithID creates an actor with a specific ID
func CreateActorWithID(pattern string, actorName string) actor.Actor {
	factory := Dispatch(pattern)
	return factory(actorName)
}
