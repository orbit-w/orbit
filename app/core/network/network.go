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
	Data []byte
}

// Encode 包体长度（4byte）｜seq（4byte）｜协议号（4byte）｜消息长度（4byte）｜消息内容（bytes）
func (c *Codec) Encode(data []byte, seq uint32, pid uint32) (packet.IPacket, error) {
	// Calculate total length: HeaderSize + data length
	totalLength := HeaderSize + len(data)

	// Create packet with appropriate size
	w := packet.WriterP(totalLength)

	// Write package length (4 bytes)
	w.WriteUint32(uint32(totalLength))

	// Write upstream sequence number (4 bytes)
	w.WriteUint32(seq)

	// Write protocol number (4 bytes)
	w.WriteUint32(pid)

	// Write message length (4 bytes)
	w.WriteUint32(uint32(len(data)))

	// Write raw data
	w.Write(data)

	return w, nil
}

// EncodeBatch 包体长度（4byte）｜seq（4byte）｜[协议号（4byte）｜消息长度（4byte）｜消息内容（bytes）]...
func (c *Codec) EncodeBatch(seq uint32, msgs []Message) (packet.IPacket, error) {
	// HeaderSize  using 8 bytes (length + seq)

	// Calculate additional size for all messages
	messagesSize := 0
	for _, msg := range msgs {
		// 4 bytes for protocol ID + 4 bytes for message length + message data length
		messagesSize += 8 + len(msg.Data)
	}

	// Calculate total length
	totalLength := initialSize + messagesSize

	// Create packet with appropriate size
	w := packet.WriterP(totalLength)

	// Write package length (4 bytes)
	w.WriteUint32(uint32(totalLength))

	// Write sequence number (4 bytes)
	w.WriteUint32(seq)

	// Write each message directly without a count
	for _, msg := range msgs {
		// Write protocol ID (4 bytes)
		w.WriteUint32(msg.Pid)

		// Write message length (4 bytes)
		w.WriteUint32(uint32(len(msg.Data)))

		// Write message data
		w.Write(msg.Data)
	}

	return w, nil
}

// 包体长度（4byte）｜上行seq（4byte）｜协议号（4byte）｜raw（bytes）
func (c *Codec) Decode(data []byte) (uint32, []byte, error) {
	// Check if data has minimum required length (4 bytes for length + 4 bytes for seq + 4 bytes for protocol)
	if len(data) < HeaderSize {
		return 0, nil, errors.New("insufficient data length")
	}

	// Extract package length (first 4 bytes)
	packageLength := binary.BigEndian.Uint32(data[0:4])

	// Validate package length
	if uint32(len(data)) < packageLength {
		return 0, nil, errors.New("incomplete package")
	}

	// Extract sequence number (next 4 bytes)
	seq := binary.BigEndian.Uint32(data[4:8])

	// Extract protocol number (next 4 bytes)
	pid := binary.BigEndian.Uint32(data[8:12])

	// Extract raw data (remaining bytes)
	rawData := data[12:]

	// You might want to use the sequence number in your application logic
	// For now, we're just logging it
	_ = seq

	return pid, rawData, nil
}
