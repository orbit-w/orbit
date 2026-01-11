package agent

import (
	gnetwork "gitee.com/orbit-w/meteor/modules/net/network"
)

func Serve(host string, protocol gnetwork.Protocol) (IServer, error) {
	factory := GetFactory(protocol)
	s := factory()

	if err := s.Serve(host); err != nil {
		panic(err)
	}

	return s, nil
}

type IServer interface {
	Serve(addr string) error
	Stop() error
}

type Factory func() IServer

var factories = make(map[gnetwork.Protocol]Factory)

func regFactory(name gnetwork.Protocol, f Factory) {
	factories[name] = f
}

func GetFactory(name gnetwork.Protocol) Factory {
	return factories[name]
}
