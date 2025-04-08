package manager

import (
	"sync/atomic"
	"unsafe"
)

const (
	// Level definitions for actors
	LevelNormal Level = iota // Normal priority level
	LevelHigh                // High priority level
	LevelMaxLimit
)

type Level int

var (
	// Map to store pattern to level mappings
	patternLevelMap = make(map[string]int)
)

func InitPatternLevelMap(list []struct {
	Pattern string
	Level   int
}) {
	m := make(map[string]int)
	for _, item := range list {
		m[item.Pattern] = item.Level
	}

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&patternLevelMap)), unsafe.Pointer(&m))
}

func GetLevelByPattern(pattern string) int {
	return patternLevelMap[pattern]
}
