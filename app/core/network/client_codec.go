package network

import (
	"errors"

	"github.com/orbit-w/meteor/modules/net/packet"
)

// ClientCodec 客户端编解码器
type ClientCodec struct{}

// NewClientCodec 创建新的客户端编解码器
func NewClientCodec() *ClientCodec {
	return &ClientCodec{}
}

// Encode 编码消息
// [seq（4byte）｜协议号（4byte）｜消息长度（4byte）｜消息内容（bytes）]
func (c *ClientCodec) Encode(data []byte, seq uint32, pid uint32) packet.IPacket {
	totalSize := HeaderSize + len(data)
	w := packet.WriterP(totalSize)
	// 写入协议号
	w.WriteUint32(pid)
	// 写入序列号
	w.WriteUint32(seq)
	// 写入消息内容
	w.WriteBytes32(data)

	return w
}

// Decode 解码消息
// [协议号（4byte）｜seq（4byte，optional）｜消息长度（4byte）｜消息内容（bytes）]...
func (c *ClientCodec) Decode(in []byte, check func(pid uint32) bool) ([]Message, error) {
	// 检查数据长度
	if len(in) < 4 {
		return nil, errors.New("insufficient data length")
	}
	// 创建读取器
	r := packet.ReaderP(in)
	defer packet.Return(r)

	var messages []Message
	for len(r.Remain()) > 0 {
		// 读取协议号
		curPid, err := r.ReadUint32()
		if err != nil {
			return nil, errors.New("failed to read protocol id")
		}

		// 初始化当前消息
		curMsg := Message{
			Pid: curPid,
		}

		// 判断是否需要读取seq
		hasSeq := check(curPid)
		if hasSeq {
			curSeq, err := r.ReadUint32()
			if err != nil {
				return nil, errors.New("failed to read sequence number")
			}
			curMsg.Seq = curSeq
		}

		// 读取消息内容
		curMsg.Data, err = r.ReadBytes32()
		if err != nil {
			return nil, errors.New("failed to read message content")
		}

		// 将当前消息添加到消息列表
		messages = append(messages, curMsg)
	}

	return messages, nil
}
