package main

import (
	"flag"

	"gitee.com/orbit-w/orbit/app"
	"gitee.com/orbit-w/orbit/app/modules/config_v2"
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

	err := config_v2.InitConfig(*configPath)
	if err != nil {
		panic(err)
	}

	app.Serve(*serverId)
}
