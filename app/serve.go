package app

import (
	"fmt"
	stream "github.com/orbit-w/orbit/app/core/services/agent_stream"
	"github.com/orbit-w/orbit/app/modules/config"
	"github.com/orbit-w/orbit/app/modules/logger"
	"github.com/orbit-w/orbit/app/modules/service"
	"os"
	"os/signal"
	"syscall"
)

/*
   @Author: orbit-w
   @File: serve
   @2024 4月 周日 17:36
*/

func Serve(nodeId string) {
	cfg := config.GetConfig()
	// Init logger
	logger.InitLogger(cfg.Server.Stage)

	// Init services
	services := service.NewServices()

	// Register services
	RegServices(services)

	if err := services.Start(); err != nil {
		panic(fmt.Sprintf("services start error: %v", err))
	}

	// Wait for exit signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	services.Stop()
	logger.ZLogger().Info("orbit service exit")
	logger.StopLogger()
}

func RegServices(services *service.Services) {
	services.Reg(new(stream.AgentStream))
}
