package agent_stream

import "gitee.com/orbit-w/orbit/app/core/network"

var (
	requestHandler func(session *network.Session, data []byte, seq, pid uint32) error
)

func RegisterRequestHandler(handler func(session *network.Session, data []byte, seq, pid uint32) error) {
	requestHandler = handler
}
