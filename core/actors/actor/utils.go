package actor

import (
	"strings"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

func parseTimeout(ops ...time.Duration) time.Duration {
	timeout := time.Second * 5
	if len(ops) > 0 && ops[0] > 0 {
		timeout = ops[0]
	}
	return timeout
}

// ExtractActorName extracts the base actor name from a PID
// For example: "system-actor-supervision-level-0/test-actor-2" -> "test-actor-2"
func ExtractActorName(pid *actor.PID) string {
	if pid == nil {
		return ""
	}

	id := pid.GetId()

	// If there's no path separator, return the whole thing
	if !strings.Contains(id, "/") {
		return id
	}

	// Split by path separator and return the last part
	parts := strings.Split(id, "/")
	return parts[len(parts)-1]
}

// ExtractActorNameFromPath extracts the base actor name from a path string
// For example: "system-actor-supervision-level-0/test-actor-2" -> "test-actor-2"
func ExtractActorNameFromPath(path string) string {
	// If there's no path separator, return the whole thing
	if !strings.Contains(path, "/") {
		return path
	}

	// Split by path separator and return the last part
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
