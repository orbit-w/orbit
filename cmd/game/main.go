package main

import (
	"flag"

	"gitee.com/orbit-w/orbit/config"
	"gitee.com/orbit-w/orbit/internal/game"
)

/*
   @Author: orbit-w
   @File: main
   @2024 4月 周日 17:13
*/

var (
	configPath = flag.String("config", "configs/config_center.yaml", "path to config file")
	serverId   = flag.Int64("server_id", 1, "server id")
)

func main() {
	flag.Parse()

	err := config.InitConfig(*configPath)
	if err != nil {
		panic(err)
	}

	game.Serve(int32(*serverId))
}
