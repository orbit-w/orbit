package main

import (
	"encoding/binary"
	"flag"
	"github.com/orbit-w/orbit/app/game"
	"github.com/orbit-w/orbit/app/modules/config"
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

	binary.BigEndian.String()

	// Wait for exit signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
