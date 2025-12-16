package conn

import (
	"context"
	"io"
	"time"

	"net"
	"sync/atomic"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	gnetwork "gitee.com/orbit-w/meteor/modules/net/network"
	"gitee.com/orbit-w/meteor/modules/net/packet"
	"gitee.com/orbit-w/orbit/gate/app/net/onet"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"go.uber.org/zap"
)

type IConn interface {
	Send(data []byte) error
	Close() error
	OnClose() error
	HandleLoop(head, body []byte)
}

type IAgent interface {
	Close() error
	Proxy(out []byte) error
}

type Conn struct {
	state  atomic.Uint32
	conn   net.Conn
	codec  *gnetwork.Codec
	ctx    context.Context
	cancel context.CancelFunc
	agent  IAgent
}

type ConnOptions struct {
	IsGzip            bool
	MaxIncomingPacket uint32
	Ctx               context.Context
	ReadTimeout       time.Duration
}

func NewConn(_conn net.Conn, _agent IAgent, op ConnOptions) *Conn {
	ctx := op.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	cCtx, cancel := context.WithCancel(ctx)
	kc := &Conn{
		conn:   _conn,
		codec:  gnetwork.NewCodec(op.MaxIncomingPacket, op.IsGzip, op.ReadTimeout),
		ctx:    cCtx,
		cancel: cancel,
		agent:  _agent,
	}

	kc.state.Store(StatusNormal)
	return kc
}

func (c *Conn) Send(data []byte) error {
	out, err := c.codec.Encode(data, 0)
	if err != nil {
		return err
	}
	defer packet.Return(out)
	if err = c.conn.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		return err
	}
	_, err = c.conn.Write(out.Data())
	return err
}

func (c *Conn) Close() error {
	if c.state.CompareAndSwap(StatusNormal, StatusClosed) {
		if c.agent != nil {
			_ = c.agent.Close()
		}
	}
	return nil
}

func (c *Conn) OnClose() error {
	if c.state.CompareAndSwap(StatusNormal, StatusClosed) {
		if c.conn != nil {
			_ = c.conn.Close()
		}
	}
	return nil
}

func (c *Conn) HandleLoop(head, body []byte) {
	var (
		err error
	)

	defer utils.RecoverPanic()
	defer func() {
		_ = c.Close()
		if err != nil {
			if err == io.EOF || onet.IsClosedConnError(err) {
				//连接正常断开
				logger.GetLogger().Info("[Conn] connection disconnected", zap.String("remote_addr", c.conn.RemoteAddr().String()))
			} else {
				logger.GetLogger().Error("[Conn] abnormal connection disconnection", zap.Error(err), zap.String("remote_addr", c.conn.RemoteAddr().String()))
			}
		}
	}()

	for {
		in, _, err := c.codec.BlockDecodeBody(c.conn, head, body)
		if err != nil {
			return
		}
		if err = c.onData(in); err != nil {
			return
		}
	}
}

func (c *Conn) onData(in []byte) error {
	data := packet.ReaderP(in)
	defer packet.Return(data)
	for len(data.Remain()) > 0 {
		bytes, err := data.ReadBytes32()
		if err == nil {
			_ = c.agent.Proxy(bytes)
		} else {
			logger.GetLogger().Error("[Conn] read data failed", zap.Error(err))
		}
	}
	return nil
}
