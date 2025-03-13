package orbit_test

import (
	"testing"

	"github.com/orbit-w/orbit/app"
	"github.com/orbit-w/orbit/app/modules/config"
)

func Setup(nodeId string) {
	config.LoadConfig("../configs/config.toml")

	app.Serve(nodeId)
}

func Test_orbit(t *testing.T) {
	Setup("game_nd00")
}
