package actor

import "errors"

var (
	ErrActorNotFound      = errors.New("actor not found")
	ErrSupervisionStopped = errors.New("supervision stopped")
)
