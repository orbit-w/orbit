package main

import (
	"flag"

	"github.com/orbit-w/orbit/app"
	"github.com/orbit-w/orbit/app/modules/config"
)

/*
   @Author: orbit-w
   @File: main
   @2024 4月 周日 17:13
*/

var (
	configPath = flag.String("config", "configs/config.toml", "path to config file")
	nodeId     = flag.String("node_id", "game_nd00", "node id")
)

func main() {
	flag.Parse()

	config.LoadConfig(*configPath)

	app.Serve(*nodeId)
}
