package network

import (
	"sync/atomic"

	"github.com/orbit-w/mux-go"
)

type Session struct {
	id      int64
	downSeq atomic.Uint32 // 下行seq
	stream  mux.IServerConn
}

func NewSession(id int64, stream mux.IServerConn) *Session {
	return &Session{
		id:     id,
		stream: stream,
	}
}

func (s *Session) Seq() uint32 {
	return s.downSeq.Add(1)
}

func (s *Session) Send(data []byte) error {
	return s.stream.Send(data)
}

func (s *Session) Close() error {
	return nil
}
