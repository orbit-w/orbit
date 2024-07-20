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
	configPath = flag.String("config", "./", "config file path")
	configName = flag.String("config_name", "config", "config file name")
	nodeId     = flag.String("node_id", "game_nd00", "node id")
)

func main() {
	flag.Parse()

	config.LoadConfig(*configName, *configPath)

	app.Serve(*nodeId)
}
