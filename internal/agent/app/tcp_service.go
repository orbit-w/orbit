package agent

import (
	"context"
	"net"
	"sync/atomic"

	gnetwork "gitee.com/orbit-w/meteor/modules/net/network"
	"gitee.com/orbit-w/orbit/internal/agent/app/agent"
	netconn "gitee.com/orbit-w/orbit/internal/agent/net/conn"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"go.uber.org/zap"
)

func init() {
	regFactory(gnetwork.TCP, func() IServer {
		return new(TcpServer)
	})
}

type TcpServer struct {
	idIncr  atomic.Uint64
	stopper Stopper
}

func (s *TcpServer) Serve(addr string) error {
	// Create a TCP listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	server := new(gnetwork.Server)
	server.Serve(gnetwork.TCP, listener, func(ctx context.Context, generic net.Conn, head, body []byte,
		op *gnetwork.AcceptorOptions) {
		idx := s.Idx()
		oAgent := agent.NewAgent(idx, generic)

		conn := netconn.GetFactory(gnetwork.TCP)(generic, oAgent, netconn.ConnOptions{
			Ctx:               ctx,
			MaxIncomingPacket: MaxInPacketSize,
			ReadTimeout:       gnetwork.ReadTimeout,
		})
		defer func() {
			_ = conn.Close()
		}()

		oAgent.BindSender(conn)
		logger.GetLogger().Info("new tcp connection, binding agent", zap.Uint64("AgentId", idx), zap.String("RemoteAddr", generic.RemoteAddr().String()))
		conn.HandleLoop(head, body)
	}, &gnetwork.AcceptorOptions{
		MaxIncomingPacket: MaxInPacketSize,
		IsGzip:            false,
	})

	s.stopper = server
	return nil
}

func (s *TcpServer) Stop() error {
	return s.stopper.Stop()
}

func (s *TcpServer) Idx() uint64 {
	return s.idIncr.Add(1)
}
