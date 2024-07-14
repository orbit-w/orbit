package game

import (
	"fmt"
	"github.com/orbit-w/meteor/bases/packet"
	"github.com/orbit-w/meteor/modules/net/agent_stream"
	"github.com/orbit-w/orbit/app/modules/config"
	"github.com/orbit-w/orbit/app/modules/logger"
	"github.com/orbit-w/orbit/app/modules/service"
	"go.uber.org/zap"
)

/*
   @Author: orbit-w
   @File: serve
   @2024 4月 周日 17:36
*/

func Serve(nodeId string) {
	logger.InitLogger()
	services := service.NewServices()

	RegServices(services)

	if err := services.Start(); err != nil {
		panic(fmt.Sprintf("services start error: %v", err))
	}

}

func RegServices(services *service.Services) {
	agentStream(services)
}

func agentStream(services *service.Services) {
	streamHandle := func(stream agent_stream.IStream) error {
		var (
			log = logger.ZLogger()
		)
		log.Info("agent_stream server start")
		for {
			in, err := stream.Recv()
			if err != nil {
				break
			}
			log.Info("agent_stream server recv", zap.String("Data", string(in)))

			w := packet.Writer()
			w.WriteInt8(0)
			w.WriteString("hello, client")
			err = stream.Send(w.Data())
			if err != nil {
				log.Error("gent_stream server send failed", zap.Error(err))
			}
			w.Return()
		}
		return nil
	}

	gs := new(agent_stream.Server)
	services.Reg(service.Wrapper("AgentStream").
		WrapStart(func() error {
			host := config.StreamHost()
			if err := gs.Serve(host, streamHandle); err != nil {
				panic(err)
			}

			logger.ZLogger().Info("AgentStream server listened...", zap.String("Host", host))
			return nil
		}).
		WrapStop(func() error {
			return gs.Stop()
		}).
		WrapLogger(logger.ZLogger()))
}

type IService interface {
	Start() error
	Stop() error
}
