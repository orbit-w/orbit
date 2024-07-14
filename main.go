package main

import (
	"flag"
	"github.com/orbit-w/orbit/app/common/config"
	"github.com/orbit-w/orbit/app/game"
	"os"
	"os/signal"
	"syscall"
)

/*
   @Author: orbit-w
   @File: main
   @2024 4月 周日 17:13
*/

var (
	configPath = flag.String("config", "./", "config file path")
	nodeId     = flag.String("node_id", "game_nd00", "node id")
)

func main() {
	flag.Parse()

	config.ParseConfig(*configPath)

	game.Serve(*nodeId)

	// Wait for exit signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
