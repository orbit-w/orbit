package actor

import (
	"sync/atomic"
	"unsafe"
)

const (
	// Level definitions for actors
	// 等级越高，停止的优先度越低
	LevelNormal Level = iota // Normal priority level
	LevelHigh                // High priority level

	LevelMaxLimit // Max limit level，无实际意义
)

type Level int

var (
	// Map to store pattern to level mappings
	patternLevelMap *map[string]Level
)

func InitPatternLevelMap(list []struct {
	Pattern string
	Level   Level
}) {
	m := make(map[string]Level)
	for _, item := range list {
		m[item.Pattern] = item.Level
	}

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&patternLevelMap)), unsafe.Pointer(&m))
}

func GetLevelByPattern(pattern string) Level {
	m := *patternLevelMap
	return m[pattern]
}
