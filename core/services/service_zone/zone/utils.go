package servicezone

import (
	"fmt"
	"time"
)

func GenActorId(zoneId string) string {
	return fmt.Sprintf("zone-%s", zoneId)
}

func parseTimeout(timeout ...time.Duration) time.Duration {
	if len(timeout) > 0 {
		return timeout[0]
	}
	return 5 * time.Second
}

func ZoneIdToDatabase(zoneId string) string {
	return fmt.Sprintf("zone_%s", zoneId)
}
