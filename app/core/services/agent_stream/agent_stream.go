package agent_stream

import (
	"github.com/orbit-w/meteor/bases/packet"
	"github.com/orbit-w/meteor/modules/net/agent_stream"
	"github.com/orbit-w/orbit/app/modules/config"
	"github.com/orbit-w/orbit/app/modules/logger"
	"go.uber.org/zap"
	"net"
)

/*
   @Author: orbit-w
   @File: agent_stream
   @2024 7月 周二 23:32
*/

var streamHandle = func(stream agent_stream.IStream) error {
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

type AgentStream struct {
	server *agent_stream.Server
}

func (a *AgentStream) Start() error {
	host := streamHost()
	a.server = new(agent_stream.Server)
	if err := a.server.Serve(host, streamHandle); err != nil {
		panic(err)
	}

	logger.ZLogger().Info("AgentStream server listened...", zap.String("Host", host))
	return nil
}

func (a *AgentStream) Stop() error {
	if a.server != nil {
		return a.server.Stop()
	}
	return nil
}

func streamHost() string {
	cfg := config.GetConfig()
	ipAddr := net.ParseIP(cfg.Server.Host)
	return net.JoinHostPort(ipAddr.String(), cfg.Server.Port)
}
