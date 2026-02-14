//go:build 386
// +build 386

#include "textflag.h"

// func spin64(l *uint64)
TEXT ·spin64(SB), NOSPLIT, $0-4
    MOVL ptr+0(FP), AX    // 将指针参数加载到AX寄存器
    
spin_loop:
    // 32位架构也支持PAUSE指令（在较新的处理器上）
    PAUSE
    
    // 在32位架构上读取64位值需要两次操作
    MOVL 0(AX), DX        // 读取低32位
    MOVL 4(AX), CX        // 读取高32位
    
    // 检查是否都为0
    ORL DX, CX            // 将两个32位值进行OR操作
    JNZ spin_loop         // 如果结果不为0，继续自旋
    
    // 使用内存屏障
    // 在某些老的32位处理器上可能不支持MFENCE，使用替代方案
    LOCK
    ADDL $0, (SP)         // 使用LOCK前缀的空操作作为内存屏障
    
    RET