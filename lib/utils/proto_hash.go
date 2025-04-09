package utils

import (
	"strings"

	"gitee.com/orbit-w/orbit/lib/logger"
)

const (
	// DefaultProtoHashSeed is the default seed used for hashing proto message names
	DefaultProtoHashSeed uint32 = 0x123456
)

// MessageNameMap stores a map of message name to their hash values
var MessageNameMap = make(map[string]uint32)

// HashProtoMessage generates a hash for a proto message name
// Format: packageName.messageName (e.g., "Core.Request_SearchBook")
// Returns a uint32 hash that can be used as a protocol ID (pid)
func HashProtoMessage(fullName string) uint32 {
	// Check if we've already hashed this message name
	if hash, ok := MessageNameMap[fullName]; ok {
		return hash
	}

	// Generate the hash using MurmurHash3
	hash := StringHash32(fullName, DefaultProtoHashSeed)

	// Store in the map for future use
	MessageNameMap[fullName] = hash

	// Log it for debugging purposes
	logger.GetLogger().Debugf("HashProtoMessage: name=%s, hash=0x%08x", fullName, hash)

	return hash
}

// HashProtoMessageBulk hashes multiple proto message names at once
// Returns a map of message name to hash value
func HashProtoMessageBulk(names []string) map[string]uint32 {
	result := make(map[string]uint32, len(names))

	for _, name := range names {
		result[name] = HashProtoMessage(name)
	}

	return result
}

// GetProtoMessageNameParts extracts package name and message name from a fully qualified proto message name
// Input: "Core.Request_SearchBook"
// Output: "Core", "Request_SearchBook"
func GetProtoMessageNameParts(fullName string) (packageName, messageName string) {
	parts := strings.SplitN(fullName, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// If no dot separator found, consider the whole string as the message name
	return "", fullName
}
