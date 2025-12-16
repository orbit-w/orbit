package conn

import (
	"net"

	"gitee.com/orbit-w/meteor/modules/net/network"
)

func init() { RegFactory(network.KCP, NewKcpConn) }

type KcpConn struct {
	IConn
}

func NewKcpConn(_conn net.Conn, _agent IAgent, op ConnOptions) IConn {
	return &KcpConn{IConn: NewConn(_conn, _agent, op)}
}
