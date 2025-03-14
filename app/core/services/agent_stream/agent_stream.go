package agent_stream

import (
	"net"

	"github.com/orbit-w/meteor/modules/net/packet"
	"github.com/orbit-w/mux-go"
	"github.com/orbit-w/orbit/app/core/network"
	"github.com/orbit-w/orbit/app/modules/config"
	"github.com/orbit-w/orbit/lib/logger"
	"go.uber.org/zap"
)

/*
   @Author: orbit-w
   @File: agent_stream
   @2024 7月 周二 23:32
*/

var streamHandle = func(stream mux.IServerConn) error {
	var (
		log = logger.GetLogger()
	)
	log.Info("agent_stream server start")
	ctx := stream.Context()
	session := network.NewSession(0, stream)
	for {
		in, err := stream.Recv(ctx)
		if err != nil {
			log.Error("conn read stream failed", zap.Error(err))
			break
		}
		log.Info("agent_stream server recv", zap.String("Data", string(in)))

		w := packet.WriterP(256)
		w.WriteInt8(0)
		w.WriteString("hello, client")
		err = stream.Send(w.Data())
		if err != nil {
			log.Error("gent_stream server send failed", zap.Error(err))
		}
		packet.Return(w)
	}
	return nil
}

type AgentStream struct {
	server *mux.Server
}

func (a *AgentStream) Start() error {
	host := streamHost()

	server := new(mux.Server)
	err := server.Serve(host, streamHandle)
	if err != nil {
		panic(err)
	}

	a.server = server

	logger.GetLogger().Info("AgentStream server listened...", zap.String("Host", host))
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
