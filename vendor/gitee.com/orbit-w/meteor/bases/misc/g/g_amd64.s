//go:build amd64

#include "textflag.h"

// func getG() uintptr
TEXT ·getG(SB), NOSPLIT, $0-8
    // 获取当前 goroutine 的 g 指针
    // 在 AMD64 上，g 指针存储在 TLS 中
    MOVQ (TLS), AX
    MOVQ AX, ret+0(FP)
    RET

// func getGIDFromG() uint64
TEXT ·getGIDFromG(SB), NOSPLIT, $0-8
    // 获取 g 指针
    MOVQ (TLS), AX
    
    // 检查 g 指针是否有效
    TESTQ AX, AX
    JZ fallback
    
    // 加载全局偏移量变量
    MOVQ ·gOffset(SB), BX
    
    // 检查偏移量是否有效
    TESTQ BX, BX
    JZ fallback
    
    // 从 g 结构体中读取 goroutine ID
    MOVQ (AX)(BX*1), CX
    
    // 检查 ID 是否有效（非零）
    TESTQ CX, CX
    JZ fallback
    
    // 返回 goroutine ID
    MOVQ CX, ret+0(FP)
    RET

fallback:
    // 返回 0 表示失败，将使用备用方案
    MOVQ $0, ret+0(FP)
    RET 
