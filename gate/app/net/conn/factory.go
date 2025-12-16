package conn

import (
	"net"

	"gitee.com/orbit-w/meteor/modules/net/network"
)

var factories = make(map[network.Protocol]Factory)

type Factory func(_conn net.Conn, _agent IAgent, op ConnOptions) IConn

func RegFactory(name network.Protocol, f Factory) {
	factories[name] = f
}

func GetFactory(name network.Protocol) Factory {
	if f, ok := factories[name]; ok {
		return f
	}
	return nil
}
