package sender_wrapper

import (
	"runtime/debug"

	"gitee.com/orbit-w/meteor/modules/net/packet"
	"gitee.com/orbit-w/meteor/modules/unbounded"
)

/*
   @Author: orbit-w
   @File: sender_wrapper
   @2023 11月 周日 19:52
*/

type SenderWrapper struct {
	sender func(body packet.IPacket) error
	ch     *unbounded.Unbounded[packet.IPacket]
}

func NewSender(sender func(body packet.IPacket) error) *SenderWrapper {
	ins := &SenderWrapper{
		sender: sender,
		ch:     unbounded.NewUnbounded[packet.IPacket](2048),
	}

	go func() {
		defer func() {
			if x := recover(); x != nil {
				debug.PrintStack()
			}
		}()

		ins.ch.Receive(func(msg packet.IPacket) bool {
			_ = ins.sender(msg)
			return false
		})
	}()

	return ins
}

func (ins *SenderWrapper) Send(data packet.IPacket) error {
	return ins.ch.Send(data)
}

func (ins *SenderWrapper) OnClose() {
	ins.ch.Close()
}
