package g

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

var (
	nextGoroutineID uint64
	gOffset         uintptr = getGoroutineIDOffset()
)

// ID 高性能获取 goroutine ID
func ID() uint64 {
	return getGoroutineID()
}

// getGoroutineID 通过汇编获取 goroutine ID
func getGoroutineID() uint64 {
	// 首先尝试通过 g 结构体直接获取
	if gOffset > 0 {
		gid := getGIDFromG()
		if gid > 0 {
			return gid
		}
	}

	// 回退到堆栈解析方法
	return getGIDFromStack()
}

// getGIDFromG 通过 g 结构体获取 goroutine ID（汇编实现）
//
//go:nosplit
func getGIDFromG() uint64

// getGIDFromStack 通过堆栈解析获取 goroutine ID（备用方案）
func getGIDFromStack() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	if n == 0 {
		return generateGoroutineID()
	}

	return fastParseGoroutineID(buf[:n])
}

// fastParseGoroutineID 快速解析 goroutine ID
func fastParseGoroutineID(buf []byte) uint64 {
	// 检查前缀 "goroutine "
	if len(buf) < 10 ||
		buf[0] != 'g' || buf[1] != 'o' || buf[2] != 'r' || buf[3] != 'o' ||
		buf[4] != 'u' || buf[5] != 't' || buf[6] != 'i' || buf[7] != 'n' ||
		buf[8] != 'e' || buf[9] != ' ' {
		return generateGoroutineID()
	}

	// 手动解析数字
	var gid uint64
	i := 10
	for i < len(buf) && buf[i] >= '0' && buf[i] <= '9' {
		gid = gid*10 + uint64(buf[i]-'0')
		i++
	}

	if i == 10 || gid == 0 {
		return generateGoroutineID()
	}

	return gid
}

// getGoroutineIDOffset 获取 goroutine ID 在 g 结构体中的偏移量
func getGoroutineIDOffset() uintptr {
	// 通过反射和运行时信息确定偏移量
	// 这个方法在不同 Go 版本中可能需要调整
	switch runtime.Version() {
	case "go1.21.0", "go1.21.1", "go1.21.2", "go1.21.3", "go1.21.4", "go1.21.5", "go1.21.6", "go1.21.7", "go1.21.8", "go1.21.9", "go1.21.10", "go1.21.11", "go1.21.12":
		return 152 // Go 1.21.x 的偏移量
	case "go1.22.0", "go1.22.1", "go1.22.2", "go1.22.3", "go1.22.4", "go1.22.5", "go1.22.6", "go1.22.7", "go1.22.8", "go1.22.9", "go1.22.10", "go1.22.11", "go1.22.12":
		return 152 // Go 1.22.x 的偏移量
	case "go1.23.0", "go1.23.1", "go1.23.2", "go1.23.3", "go1.23.4", "go1.23.5", "go1.23.6", "go1.23.7", "go1.23.8", "go1.23.9", "go1.23.10", "go1.23.11", "go1.23.12":
		return 152 // Go 1.23.x 的偏移量
	default:
		// 对于未知版本，尝试动态检测
		return detectGoroutineIDOffset()
	}
}

// getG 获取当前 goroutine 的 g 指针（汇编实现）
//
//go:nosplit
func getG() uintptr

// detectGoroutineIDOffset 动态检测 goroutine ID 偏移量
func detectGoroutineIDOffset() uintptr {
	// 通过比较堆栈解析的 ID 和内存中的值来确定偏移量
	stackGID := getGIDFromStack()
	if stackGID == 0 {
		return 0
	}

	g := getG()
	if g == 0 {
		return 0
	}

	// 在合理范围内搜索匹配的偏移量
	for offset := uintptr(0); offset < 512; offset += 8 {
		if *(*uint64)(unsafe.Pointer(g + offset)) == stackGID {
			return offset
		}
	}

	return 0
}

func generateGoroutineID() uint64 {
	return atomic.AddUint64(&nextGoroutineID, 1)
}
