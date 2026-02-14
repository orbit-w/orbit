//go:build arm64

#include "textflag.h"

// func getG() uintptr
TEXT ·getG(SB), NOSPLIT, $0-8
    // 获取当前 goroutine 的 g 指针
    // 在 ARM64 上，g 指针存储在 g 寄存器中
    MOVD g, R0
    MOVD R0, ret+0(FP)
    RET

// func getGIDFromG() uint64
TEXT ·getGIDFromG(SB), NOSPLIT, $0-8
    // 获取 g 指针
    MOVD g, R0
    
    // 检查 g 指针是否有效
    CBZ R0, fallback
    
    // 加载全局偏移量变量
    MOVD ·gOffset(SB), R1
    
    // 检查偏移量是否有效
    CBZ R1, fallback
    
    // 从 g 结构体中读取 goroutine ID
    MOVD (R0)(R1), R2
    
    // 检查 ID 是否有效（非零）
    CBZ R2, fallback
    
    // 返回 goroutine ID
    MOVD R2, ret+0(FP)
    RET

fallback:
    // 返回 0 表示失败，将使用备用方案
    MOVD $0, ret+0(FP)
    RET 
