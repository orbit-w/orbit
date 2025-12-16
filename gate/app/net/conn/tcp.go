package conn

import (
	"net"

	"gitee.com/orbit-w/meteor/modules/net/network"
)

func init() { RegFactory(network.TCP, NewTcpConn) }

type TcpConn struct {
	IConn
}

func NewTcpConn(_conn net.Conn, _agent IAgent, op ConnOptions) IConn {
	return &TcpConn{IConn: NewConn(_conn, _agent, op)}
}
