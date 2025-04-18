package actor

import "errors"

var (
	ErrActorNotFound      = errors.New("actor not found")
	ErrActorStopped       = errors.New("actor is stopped")
	ErrSupervisionStopped = errors.New("supervision stopped")
)
