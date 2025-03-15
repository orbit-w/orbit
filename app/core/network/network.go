package network

import (
	"encoding/binary"
	"errors"

	"github.com/orbit-w/meteor/modules/net/packet"
)

const (
	HeaderSize  = 12 // 包体长度（4byte） | seq（4byte） | 协议号（4byte）
	initialSize = 8
)

var (
	codec = new(Codec)
)

type Codec struct{}

type Message struct {
	Pid  uint32
	Seq  uint32
	Data []byte
}

// Encode 包体长度（4byte）｜[消息长度（4byte）｜协议号（4byte）｜seq（4byte，optional）｜消息内容（bytes）]...
func (c *Codec) Encode(data []byte, seq uint32, pid uint32) (packet.IPacket, error) {
	// Calculate message size: 4(protocol) + len(data) + optional 4(seq)
	msgSize := 4 + len(data) // 4(protocol)
	if seq != 0 {
		msgSize += 4 // Add 4 bytes for seq
	}

	// Total size: 4(package length) + 4(message length) + msgSize
	totalSize := 4 + 4 + msgSize

	// Create packet with total size
	w := packet.WriterP(totalSize)

	// Write total package length (4 bytes)
	w.WriteUint32(uint32(totalSize))

	// Write message length (4 bytes)
	w.WriteUint32(uint32(msgSize))

	// Write protocol number (4 bytes)
	w.WriteUint32(pid)

	// Write sequence number (4 bytes) if not 0
	if seq != 0 {
		w.WriteUint32(seq)
	}

	// Write message content
	w.Write(data)

	return w, nil
}

// EncodeBatch 包体长度（4byte）｜[消息长度（4byte）｜协议号（4byte）｜seq（4byte，optional）｜消息内容（bytes）]...
func (c *Codec) EncodeBatch(msgs []Message) (packet.IPacket, error) {
	// Calculate total size for all messages
	totalSize := 4 // Initial 4 bytes for package length
	for _, msg := range msgs {
		// Base size: 4(length) + 4(protocol) + len(data)
		msgSize := 8 + len(msg.Data)
		// Add 4 bytes for seq if it's not 0
		if msg.Seq != 0 {
			msgSize += 4
		}
		totalSize += 4 + msgSize // 4 for message length
	}

	// Create packet with total size
	w := packet.WriterP(totalSize)

	// Write total package length (4 bytes)
	w.WriteUint32(uint32(totalSize))

	// Write each message
	for _, msg := range msgs {
		// Calculate message block size: 4(protocol) + len(data) + optional 4(seq)
		msgSize := 8 + len(msg.Data) // 4(protocol) + 4(seq)
		if msg.Seq != 0 {
			msgSize += 4
		}

		// Write message length (4 bytes)
		w.WriteUint32(uint32(msgSize))

		// Write protocol number (4 bytes)
		w.WriteUint32(msg.Pid)

		// Write sequence number (4 bytes) if not 0
		if msg.Seq != 0 {
			w.WriteUint32(msg.Seq)
		}

		// Write message content
		w.Write(msg.Data)
	}

	return w, nil
}

// 包体长度（4byte）｜上行seq（4byte）｜协议号（4byte）｜raw（bytes）
func (c *Codec) Decode(in []byte) (pid, seq uint32, data []byte, err error) {
	// Check if data has minimum required length (4 bytes for length + 4 bytes for seq + 4 bytes for protocol)
	if len(in) < HeaderSize {
		return 0, 0, nil, errors.New("insufficient data length")
	}

	// Extract package length (first 4 bytes)
	packageLength := binary.BigEndian.Uint32(in[0:4])

	// Validate package length
	if uint32(len(in)) < packageLength {
		return 0, 0, nil, errors.New("incomplete package")
	}

	// Extract sequence number (next 4 bytes)
	seq = binary.BigEndian.Uint32(in[4:8])

	// Extract protocol number (next 4 bytes)
	pid = binary.BigEndian.Uint32(in[8:12])

	// Extract raw data (remaining bytes)
	data = in[12:]

	return
}
