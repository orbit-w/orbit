package examples

import (
	"fmt"

	"github.com/orbit-w/orbit/lib/utils"
)

// MurmurHashExample demonstrates the use of the MurmurHash3 implementation
func MurmurHashExample() {
	// Sample strings to hash
	examples := []string{
		"",
		"hello",
		"hello, world!",
		"The quick brown fox jumps over the lazy dog",
	}

	fmt.Println("MurmurHash3 Example:")
	fmt.Println("=====================")

	// Demonstrate 32-bit hashing
	fmt.Println("\n32-bit MurmurHash3 results:")
	fmt.Println("------------------------")
	for _, s := range examples {
		// Hash with seed 0
		hash := utils.MurmurHash3_x86_32([]byte(s), 0)
		fmt.Printf("Input: %q\nHash: 0x%08x\n\n", s, hash)

		// Also demonstrate using a different seed
		hashWithSeed := utils.MurmurHash3_x86_32([]byte(s), 42)
		fmt.Printf("Input: %q (with seed 42)\nHash: 0x%08x\n\n", s, hashWithSeed)
	}

	// Demonstrate 128-bit hashing (64-bit variant, returns 2 uint64s)
	fmt.Println("\n128-bit MurmurHash3 results (x64 variant):")
	fmt.Println("----------------------------------------")
	for _, s := range examples {
		// Hash with seed 0
		hash := utils.MurmurHash3_x64_128([]byte(s), 0)
		fmt.Printf("Input: %q\nHash: [0x%016x, 0x%016x]\n\n", s, hash[0], hash[1])
	}

	// Demonstrate convenience string hashing functions
	fmt.Println("\nConvenience string hashing functions:")
	fmt.Println("------------------------------------")
	for _, s := range examples {
		// String hash 32
		hash32 := utils.StringHash32(s, 0)
		fmt.Printf("Input: %q\nStringHash32: 0x%08x\n", s, hash32)

		// String hash 128
		hash128 := utils.StringHash128(s, 0)
		fmt.Printf("StringHash128: [0x%016x, 0x%016x]\n\n", hash128[0], hash128[1])
	}

	// Practical example: Using MurmurHash3 for sharding or data distribution
	fmt.Println("\nPractical Example - Data Sharding:")
	fmt.Println("--------------------------------")
	shardCount := 8 // Number of shards
	keys := []string{
		"user_123456",
		"user_789012",
		"user_345678",
		"product_12345",
		"product_67890",
	}

	fmt.Println("Distributing keys across 8 shards:")
	for _, key := range keys {
		// Use MurmurHash3_x86_32 for sharding, consistent for the same key
		hash := utils.StringHash32(key, 0)
		// Use modulo to determine shard
		shardNumber := hash % uint32(shardCount)
		fmt.Printf("Key: %q -> Shard: %d (Hash: 0x%08x)\n", key, shardNumber, hash)
	}
}
