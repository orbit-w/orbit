package actor

import "time"

// Send 单向投递消息给 Actor
func Send(actorName, pattern string, msg any, timeout ...time.Duration) error {
	pid, err := GetOrStartActor(actorName, pattern)
	if err != nil {
		return err
	}

	Facade.ActorSystem().Root.Send(pid, msg)
	return nil
}

// Request 与 Actor 进行请求-响应模式通信，默认超时5秒
func RequestFuture(actorName, pattern string, msg any, timeout ...time.Duration) (any, error) {
	pid, err := GetOrStartActor(actorName, pattern)
	if err != nil {
		return nil, err
	}

	future := Facade.ActorSystem().Root.RequestFuture(pid, msg, parseTimeout(timeout...))
	result, err := future.Result()
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

func parseTimeout(ops ...time.Duration) time.Duration {
	timeout := time.Second * 5
	if len(ops) > 0 && ops[0] > 0 {
		timeout = ops[0]
	}
	return timeout
}
