package network

import (
	"sync/atomic"

	"github.com/orbit-w/meteor/modules/net/packet"
	"github.com/orbit-w/mux-go"
)

const (
	PatternNone = iota
	PatternKick
)

var globalSessionId atomic.Int64

type Session struct {
	id     int64 // 全局唯一自增ID
	uid    int64 // 用户ID
	stream mux.IServerConn
	codec  *Codec
	closed atomic.Int32 // 0: not closed, 1: closed
}

// NewSession 创建新的会话，自动分配全局唯一ID
func NewSession(uid int64, stream mux.IServerConn) *Session {
	return &Session{
		id:     globalSessionId.Add(1), // 原子操作，保证唯一性
		uid:    uid,
		stream: stream,
		codec:  new(Codec),
	}
}

func (s *Session) Id() int64 {
	return s.id
}

func (s *Session) Uid() int64 {
	return s.uid
}

// Send 发送原始数据
func (s *Session) Send(data []byte) error {
	w := packet.WriterP(1 + len(data))
	w.WriteInt8(PatternNone)
	w.Write(data)
	defer packet.Return(w)
	return s.stream.Send(w.Data())
}

func (s *Session) Kick() error {
	w := packet.WriterP(1)
	w.WriteInt8(PatternKick)
	defer packet.Return(w)
	return s.stream.Send(w.Data())
}

// SendMessage 发送需要编码的消息
func (s *Session) SendMessage(data []byte, seq uint32, pid uint32) error {
	pack, err := s.codec.Encode(data, seq, pid)
	if err != nil {
		return err
	}
	defer packet.Return(pack)
	return s.Send(pack.Data())
}

func (s *Session) SendMessageBatch(msgs []Message) error {
	pack, err := s.codec.EncodeBatch(msgs)
	if err != nil {
		return err
	}
	defer packet.Return(pack)
	return s.Send(pack.Data())
}

// Push 直接发送消息，不需要序列号
func (s *Session) Push(data []byte, pid uint32) error {
	pack, err := s.codec.Encode(data, 0, pid)
	if err != nil {
		return err
	}
	defer packet.Return(pack)
	return s.stream.Send(pack.Data())
}

func (s *Session) PushBatch(msgs []Message) error {
	pack, err := s.codec.EncodeBatch(msgs)
	if err != nil {
		return err
	}
	defer packet.Return(pack)
	return s.stream.Send(pack.Data())
}

func (s *Session) Decode(data []byte) ([]Message, error) {
	return s.codec.Decode(data)
}

// Close 关闭会话，保证只执行一次
func (s *Session) Close() {
	// CompareAndSwap returns true if the swap was successful
	if s.closed.CompareAndSwap(0, 1) {
		s.stream.Close()
	}
}
