//go:build 386

#include "textflag.h"

// func getG() uintptr
TEXT ·getG(SB), NOSPLIT, $0-4
    // 获取当前 goroutine 的 g 指针
    // 在 386 上，g 指针存储在 TLS 中
    MOVL (TLS), AX
    MOVL AX, ret+0(FP)
    RET

// func getGIDFromG() uint64
TEXT ·getGIDFromG(SB), NOSPLIT, $0-8
    // 获取 g 指针
    MOVL (TLS), AX
    
    // 检查 g 指针是否有效
    TESTL AX, AX
    JZ fallback
    
    // 加载全局偏移量变量
    MOVL ·gOffset(SB), BX
    
    // 检查偏移量是否有效
    TESTL BX, BX
    JZ fallback
    
    // 从 g 结构体中读取 goroutine ID (64位值)
    // 在 32 位系统上需要读取两个 32 位值
    MOVL (AX)(BX*1), CX      // 低 32 位
    MOVL 4(AX)(BX*1), DX     // 高 32 位
    
    // 检查 ID 是否有效（非零）
    TESTL CX, CX
    JNZ valid
    TESTL DX, DX
    JZ fallback

valid:
    // 返回 64 位 goroutine ID
    MOVL CX, ret+0(FP)       // 低 32 位
    MOVL DX, ret+4(FP)       // 高 32 位
    RET

fallback:
    // 返回 0 表示失败，将使用备用方案
    MOVL $0, ret+0(FP)
    MOVL $0, ret+4(FP)
    RET 
