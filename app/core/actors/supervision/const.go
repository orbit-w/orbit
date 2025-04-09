package supervision

import "time"

const (
	StartActorTimeout              = 5 * time.Second
	StopActorTimeout               = 30 * time.Second
	ManagerStartActorFutureTimeout = 30 * time.Second
)
