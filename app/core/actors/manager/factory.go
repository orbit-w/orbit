package manager

// Factory function type that accepts an actor ID
type BehaivorFactory func(actorName string) Behavior

var factories = make(map[string]BehaivorFactory)

// RegFactory registers a factory for a specific actor type
func RegFactory(pattern string, factory BehaivorFactory) {
	if _, ok := factories[pattern]; ok {
		panic("factory already registered: " + pattern)
	}
	factories[pattern] = factory
}

// RegFactories registers multiple factories at once
func RegFactories(factories ...struct {
	pattern string
	factory BehaivorFactory
}) {
	for _, cell := range factories {
		RegFactory(cell.pattern, cell.factory)
	}
}

// Dispatch returns a factory function for creating actors with the given ID
func Dispatch(pattern string) BehaivorFactory {
	factory, ok := factories[pattern]
	if !ok {
		panic("factory not found: " + pattern)
	}

	return factory
}

// CreateBehaivorWithID creates an actor with a specific ID
func CreateBehaivorWithID(pattern string, actorName string) Behavior {
	factory := Dispatch(pattern)
	return factory(actorName)
}
