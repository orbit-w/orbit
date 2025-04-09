package network

import (
	"errors"

	"gitee.com/orbit-w/meteor/modules/net/packet"
)

const (
	HeaderSize  = 12 // 包体长度（4byte） | seq（4byte） | 协议号（4byte）
	initialSize = 8
)

type Codec struct{}

type Message struct {
	Pid  uint32
	Seq  uint32
	Data []byte
}

// Encode [协议号（4byte）｜seq（4byte，optional）｜消息长度（4byte）｜消息内容（bytes）]...
func (c *Codec) Encode(data []byte, seq uint32, pid uint32) (packet.IPacket, error) {
	// Calculate message size: 4(protocol) + len(data) + optional 4(seq)
	msgSize := initialSize + len(data) // 4(protocol)
	if seq != 0 {
		msgSize += 4 // Add 4 bytes for seq
	}

	// Create packet with total size
	w := packet.WriterP(msgSize)

	// Write protocol number (4 bytes)
	w.WriteUint32(pid)

	// Write sequence number (4 bytes) if not 0
	if seq != 0 {
		w.WriteUint32(seq)
	}

	// Write message content
	w.WriteBytes32(data)

	return w, nil
}

// EncodeBatch 消息类型(1byte) [协议号（4byte）｜seq（4byte，optional）｜消息长度（4byte）｜消息内容（bytes）]...
func (c *Codec) EncodeBatch(msgList []Message) (packet.IPacket, error) {
	// Calculate total size for all messages
	totalSize := 1 // Initial 4 bytes for package length
	for _, msg := range msgList {
		// Base size: 4(protocol) + 4(seq) + 4(length) + len(data)
		msgSize := 8 + len(msg.Data)
		// Add 4 bytes for seq if it's not 0
		if msg.Seq != 0 {
			msgSize += 4
		}
		totalSize += msgSize // 4 for message length
	}

	// Create packet with total size
	w := packet.WriterP(totalSize)
	w.WriteInt8(PatternNone)

	// Write each message
	for _, msg := range msgList {
		// Write protocol number (4 bytes)
		w.WriteUint32(msg.Pid)

		// Write sequence number (4 bytes) if not 0
		if msg.Seq != 0 {
			w.WriteUint32(msg.Seq)
		}

		// Write message length (4 bytes)
		w.WriteBytes32(msg.Data)
	}

	return w, nil
}

// Decode [协议号（4byte）｜seq（4byte，optional）｜消息长度（4byte）｜消息内容（bytes）]...
func (c *Codec) Decode(in []byte) ([]Message, error) {
	// Check if data has minimum required length (4 bytes for length + 4 bytes for seq + 4 bytes for protocol)
	if len(in) < HeaderSize {
		return nil, errors.New("insufficient data length")
	}

	r := packet.ReaderP(in)
	defer packet.Return(r)
	msgList := make([]Message, 0)

	for len(r.Remain()) > 0 {
		// Extract protocol number (next 4 bytes)
		pid, err := r.ReadUint32()
		if err != nil {
			return nil, err
		}

		// Extract sequence number (next 4 bytes)
		seq, err := r.ReadUint32()
		if err != nil {
			return nil, err
		}

		// Extract raw data (remaining bytes)
		data, err := r.ReadBytes32()
		if err != nil {
			return nil, err
		}

		msgList = append(msgList, Message{
			Pid:  pid,
			Seq:  seq,
			Data: data,
		})
	}

	return msgList, nil
}
