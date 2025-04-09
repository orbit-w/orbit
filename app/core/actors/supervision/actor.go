package supervision

import "time"

func Cast(actorName, pattern string, msg any) error {
	pid, err := GetOrStartActor(actorName, pattern)
	if err != nil {
		return err
	}

	Facade.ActorSystem().Root.Send(pid, msg)
	return nil
}

func Call(actorName, pattern string, msg any) (any, error) {
	pid, err := GetOrStartActor(actorName, pattern)
	if err != nil {
		return nil, err
	}

	futrue := Facade.ActorSystem().Root.RequestFuture(pid, msg, time.Second*5)
	result, err := futrue.Result()
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case error:
		return nil, v
	default:
		return v, nil
	}
}
