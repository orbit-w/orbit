package main

import (
	"flag"

	"gitee.com/orbit-w/orbit/app"
	"gitee.com/orbit-w/orbit/app/modules/config"
)

/*
   @Author: orbit-w
   @File: main
   @2024 4月 周日 17:13
*/

var (
	configPath = flag.String("config", "configs/config.toml", "path to config file")
	serverId   = flag.String("server_id", "1", "server id")
)

func main() {
	flag.Parse()

	config.InitConfig(*configPath)

	app.Serve(*serverId)
}
