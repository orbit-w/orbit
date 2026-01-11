package agent

import (
	"context"
	"fmt"
	"net"

	"gitee.com/orbit-w/meteor/modules/mlog"
	gnetwork "gitee.com/orbit-w/meteor/modules/net/network"
	"gitee.com/orbit-w/orbit/config"
	agent "gitee.com/orbit-w/orbit/internal/agent/app"
	multiplexers_lib "gitee.com/orbit-w/orbit/lib/module/mux"
	netutils "gitee.com/orbit-w/orbit/lib/utils/net_utils"

	"gitee.com/orbit-w/orbit/lib/module/logger"
	"go.uber.org/zap"
)

/*
   @Author: orbit-w
   @File: gateway
   @2024 3月 周日 17:54
*/

func Serve() (stopper func(ctx context.Context) error, err error) {
	InitLogger()

	host := joinHost()
	protocol := parseProtocol(config.AgentProtocol())
	service, err := agent.Serve(host, protocol)
	if err != nil {
		return nil, err
	}

	stopper = func(ctx context.Context) error {
		logger.GetLogger().Info("Gateway server stopping",
			zap.String("host", host),
			zap.String("protocol", string(protocol)))

		// Add command line output for better visibility
		fmt.Println("Gateway server stopping...")

		multiplexers_lib.CloseAll()

		if err = service.Stop(); err != nil {
			logger.GetLogger().Error("Gateway server stop failed",
				zap.Error(err),
				zap.String("host", host),
				zap.String("protocol", string(protocol)))

			// Add command line output for better visibility
			fmt.Printf("Gateway server stop failed: %v\n", err)

			return err
		}

		logger.GetLogger().Info("Gateway server stopped successfully")

		// Add command line output for better visibility
		fmt.Println("Gateway server stopped successfully")
		return nil
	}

	return stopper, nil
}

func joinHost() string {
	ip, err := netutils.GetLocalIPv4()
	if err != nil {
		panic(err)
	}
	port := config.GetAgentPort()
	return net.JoinHostPort(ip, port)
}

func parseProtocol(p string) gnetwork.Protocol {
	switch p {
	case config.ProtocolTCP:
		return gnetwork.TCP
	case config.ProtocolUDP:
		return gnetwork.UDP
	case config.ProtocolKCP:
		return gnetwork.KCP
	default:
		return gnetwork.TCP
	}
}

func InitLogger() {
	log := mlog.NewFileLogger(mlog.WithLevel("info"),
		mlog.WithFormat("console"),
		mlog.WithRotation(500, 7, 3, false),
		mlog.WithInitialFields(map[string]any{"app": "orbit-agent"}),
		mlog.WithOutputPaths("logs/orbit-agent.log"))

	logger.SetLogger(log)
}
