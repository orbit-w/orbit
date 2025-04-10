package actor

import "time"

func parseTimeout(ops ...time.Duration) time.Duration {
	timeout := time.Second * 5
	if len(ops) > 0 && ops[0] > 0 {
		timeout = ops[0]
	}
	return timeout
}
