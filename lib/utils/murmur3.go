// Package utils provides utility functions for the application
package utils

import (
	"encoding/binary"
	"math/bits"
)

const (
	c1_32  uint32 = 0xcc9e2d51
	c2_32  uint32 = 0x1b873593
	c1_128 uint32 = 0x239b961b
	c2_128 uint32 = 0xab0e9789
	c3_128 uint32 = 0x38b34ae5
	c4_128 uint32 = 0xa1e38b93
)

// MurmurHash3_x86_32 implements the 32-bit variant of MurmurHash3 for x86
// It is optimized for speed on x86 architectures
func MurmurHash3_x86_32(data []byte, seed uint32) uint32 {
	h1 := seed
	nblocks := len(data) / 4

	// Body
	for i := 0; i < nblocks; i++ {
		k1 := binary.LittleEndian.Uint32(data[i*4:])

		k1 *= c1_32
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_32

		h1 ^= k1
		h1 = bits.RotateLeft32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	// Tail
	tail := data[nblocks*4:]
	var k1 uint32
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1_32
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_32
		h1 ^= k1
	}

	// Finalization
	h1 ^= uint32(len(data))
	h1 = fmix32(h1)

	return h1
}

// fmix32 finalizes the hash value for the 32-bit variant
func fmix32(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

// MurmurHash3_x86_128 implements the 128-bit variant of MurmurHash3 for x86
// Returns an array containing the 128-bit hash (as 4 uint32 values)
func MurmurHash3_x86_128(data []byte, seed uint32) [4]uint32 {
	h1 := seed
	h2 := seed
	h3 := seed
	h4 := seed

	nblocks := len(data) / 16

	// Body
	for i := 0; i < nblocks; i++ {
		// Get the next 16 bytes as 4 uint32 values
		k1 := binary.LittleEndian.Uint32(data[i*16:])
		k2 := binary.LittleEndian.Uint32(data[i*16+4:])
		k3 := binary.LittleEndian.Uint32(data[i*16+8:])
		k4 := binary.LittleEndian.Uint32(data[i*16+12:])

		// Mix k1
		k1 *= c1_128
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_128
		h1 ^= k1
		h1 = bits.RotateLeft32(h1, 19)
		h1 += h2
		h1 = h1*5 + 0x561ccd1b

		// Mix k2
		k2 *= c2_128
		k2 = bits.RotateLeft32(k2, 16)
		k2 *= c3_128
		h2 ^= k2
		h2 = bits.RotateLeft32(h2, 17)
		h2 += h3
		h2 = h2*5 + 0x0bcaa747

		// Mix k3
		k3 *= c3_128
		k3 = bits.RotateLeft32(k3, 17)
		k3 *= c4_128
		h3 ^= k3
		h3 = bits.RotateLeft32(h3, 15)
		h3 += h4
		h3 = h3*5 + 0x96cd1c35

		// Mix k4
		k4 *= c4_128
		k4 = bits.RotateLeft32(k4, 18)
		k4 *= c1_128
		h4 ^= k4
		h4 = bits.RotateLeft32(h4, 13)
		h4 += h1
		h4 = h4*5 + 0x32ac3b17
	}

	// Tail
	tail := data[nblocks*16:]
	var k1, k2, k3, k4 uint32

	switch len(tail) {
	case 15:
		k4 ^= uint32(tail[14]) << 16
		fallthrough
	case 14:
		k4 ^= uint32(tail[13]) << 8
		fallthrough
	case 13:
		k4 ^= uint32(tail[12])
		k4 *= c4_128
		k4 = bits.RotateLeft32(k4, 18)
		k4 *= c1_128
		h4 ^= k4
		fallthrough
	case 12:
		k3 ^= uint32(tail[11]) << 24
		fallthrough
	case 11:
		k3 ^= uint32(tail[10]) << 16
		fallthrough
	case 10:
		k3 ^= uint32(tail[9]) << 8
		fallthrough
	case 9:
		k3 ^= uint32(tail[8])
		k3 *= c3_128
		k3 = bits.RotateLeft32(k3, 17)
		k3 *= c4_128
		h3 ^= k3
		fallthrough
	case 8:
		k2 ^= uint32(tail[7]) << 24
		fallthrough
	case 7:
		k2 ^= uint32(tail[6]) << 16
		fallthrough
	case 6:
		k2 ^= uint32(tail[5]) << 8
		fallthrough
	case 5:
		k2 ^= uint32(tail[4])
		k2 *= c2_128
		k2 = bits.RotateLeft32(k2, 16)
		k2 *= c3_128
		h2 ^= k2
		fallthrough
	case 4:
		k1 ^= uint32(tail[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1_128
		k1 = bits.RotateLeft32(k1, 15)
		k1 *= c2_128
		h1 ^= k1
	}

	// Finalization
	h1 ^= uint32(len(data))
	h2 ^= uint32(len(data))
	h3 ^= uint32(len(data))
	h4 ^= uint32(len(data))

	h1 += h2 + h3 + h4
	h2 += h1
	h3 += h1
	h4 += h1

	h1 = fmix32(h1)
	h2 = fmix32(h2)
	h3 = fmix32(h3)
	h4 = fmix32(h4)

	h1 += h2 + h3 + h4
	h2 += h1
	h3 += h1
	h4 += h1

	return [4]uint32{h1, h2, h3, h4}
}

// MurmurHash3_x64_128 implements the 128-bit variant of MurmurHash3 for x64
// Returns the 128-bit hash as two uint64 values
func MurmurHash3_x64_128(data []byte, seed uint64) [2]uint64 {
	h1 := seed
	h2 := seed

	c1 := uint64(0x87c37b91114253d5)
	c2 := uint64(0x4cf5ad432745937f)

	nblocks := len(data) / 16

	// Body
	for i := 0; i < nblocks; i++ {
		k1 := binary.LittleEndian.Uint64(data[i*16:])
		k2 := binary.LittleEndian.Uint64(data[i*16+8:])

		k1 *= c1
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= c2
		h1 ^= k1

		h1 = bits.RotateLeft64(h1, 27)
		h1 += h2
		h1 = h1*5 + 0x52dce729

		k2 *= c2
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= c1
		h2 ^= k2

		h2 = bits.RotateLeft64(h2, 31)
		h2 += h1
		h2 = h2*5 + 0x38495ab5
	}

	// Tail
	tail := data[nblocks*16:]
	var k1, k2 uint64

	switch len(tail) & 15 {
	case 15:
		k2 ^= uint64(tail[14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(tail[13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(tail[12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(tail[11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(tail[10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(tail[9]) << 8
		fallthrough
	case 9:
		k2 ^= uint64(tail[8])
		k2 *= c2
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= c1
		h2 ^= k2
		fallthrough
	case 8:
		k1 ^= uint64(tail[7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(tail[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(tail[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(tail[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(tail[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(tail[0])
		k1 *= c1
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= c2
		h1 ^= k1
	}

	// Finalization
	h1 ^= uint64(len(data))
	h2 ^= uint64(len(data))

	h1 += h2
	h2 += h1

	h1 = fmix64(h1)
	h2 = fmix64(h2)

	h1 += h2
	h2 += h1

	return [2]uint64{h1, h2}
}

// fmix64 finalizes the hash value for the 64-bit variant
func fmix64(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}

// StringHash32 is a convenience function that computes a 32-bit MurmurHash3 of a string
func StringHash32(s string, seed uint32) uint32 {
	return MurmurHash3_x86_32([]byte(s), seed)
}

// StringHash128 is a convenience function that computes a 128-bit MurmurHash3 of a string
// using the x64 variant which is faster on 64-bit systems
func StringHash128(s string, seed uint64) [2]uint64 {
	return MurmurHash3_x64_128([]byte(s), seed)
}
