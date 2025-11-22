package servicezone

import (
	"testing"
)

func TestZone(t *testing.T) {
	Cast("test", &Request{
		ZoneId: "1",
	})
}
