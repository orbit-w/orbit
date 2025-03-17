package agent_stream

import (
	"context"
	"errors"
	"io"
	"net"

	gnetwork "github.com/orbit-w/meteor/modules/net/network"
	"github.com/orbit-w/mux-go"
	"github.com/orbit-w/mux-go/metadata"
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

const (
	keyUid = "uid"
)

var streamHandle = func(stream mux.IServerConn) error {
	var (
		log = logger.GetLogger()
	)

	// 处理连接：退出情况下，系统会自动关闭stream，不需要业务层显示处理
	session, err := newSession(stream)
	if err != nil {
		log.Error("new session failed", zap.Error(err))
		return err
	}
	defer session.Close()
	log.Info("agent_stream server start", zap.Int64("uid", session.Uid()))

	for {
		in, err := stream.Recv(context.Background())
		if err != nil {
			if gnetwork.IsClosedConnError(err) || errors.Is(err, io.EOF) {
				log.Info("stream closed", zap.Int64("uid", session.Uid()))
				break
			}
			log.Error("conn read stream failed", zap.Error(err))
			break
		}

		// TODO: 处理消息
		msgList, err := session.Decode(in)
		if err != nil {
			log.Error("decode failed", zap.Error(err))
			stream.Close()
			break
		}

		for _, msg := range msgList {
			session.SendMessage([]byte("hello, client"), msg.Seq, msg.Pid)
		}
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

func newSession(stream mux.IServerConn) (*network.Session, error) {
	ctx := stream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.GetLogger().Error("read stream metadata failed")
		return nil, errors.New("read stream metadata failed")
	}
	uid, exist := md.GetInt64(keyUid)
	if !exist {
		logger.GetLogger().Error("read stream metadata failed")
		return nil, errors.New("read stream metadata failed")
	}

	session := network.NewSession(uid, stream)
	return session, nil
}

func streamHost() string {
	cfg := config.GetConfig()
	ipAddr := net.ParseIP(cfg.Server.Host)
	return net.JoinHostPort(ipAddr.String(), cfg.Server.Port)
}
