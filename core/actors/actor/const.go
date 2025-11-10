package actor

import "time"

const (
	StartActorTimeout              = 30 * time.Second
	StopActorTimeout               = 30 * time.Second
	ManagerStartActorFutureTimeout = 30 * time.Second

	aliveCheckTimerKey  = "system_alive_check_timer"
	AliveCheckInterval  = 30 * time.Second
	DefaultAliveTimeout = 30 * time.Minute
)
