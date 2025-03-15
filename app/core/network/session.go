package network

import (
	"sync/atomic"

	"github.com/orbit-w/meteor/modules/net/packet"
	"github.com/orbit-w/mux-go"
)

var globalSessionId atomic.Int64

type Session struct {
	id     int64 // 全局唯一自增ID
	uid    int64 // 用户ID
	stream mux.IServerConn
	codec  *Codec
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

// Send 发送原始数据
func (s *Session) Send(data []byte) error {
	return s.stream.Send(data)
}

// SendMessage 发送需要编码的消息
func (s *Session) SendMessage(data []byte, seq uint32, pid uint32) error {
	pack, err := s.codec.Encode(data, seq, pid)
	if err != nil {
		return err
	}
	defer packet.Return(pack)
	return s.stream.Send(pack.Data())
}

func (s *Session) SendMessageBatch(msgs []Message) error {
	pack, err := s.codec.EncodeBatch(msgs)
	if err != nil {
		return err
	}
	defer packet.Return(pack)
	return s.stream.Send(pack.Data())
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

func (s *Session) Decode(data []byte) (uint32, uint32, []byte, error) {
	return s.codec.Decode(data)
}

// Close 关闭会话
func (s *Session) Close() error {
	return nil
}
