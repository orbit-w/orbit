package agent

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// GenerateUniqueId generates a unique int64 ID using a combination of timestamp and random bits
func GenerateUniqueId() int64 {
	// Use top 41 bits for timestamp
	timestamp := time.Now().UnixNano() / 1e6 // millisecond precision
	timestampBits := (timestamp & ((1 << 41) - 1)) << 22

	// Use remaining 22 bits for random data
	b := make([]byte, 8)
	rand.Read(b)
	randomBits := int64(binary.BigEndian.Uint64(b) & ((1 << 22) - 1))

	// Combine timestamp and random bits
	return timestampBits | randomBits
}
