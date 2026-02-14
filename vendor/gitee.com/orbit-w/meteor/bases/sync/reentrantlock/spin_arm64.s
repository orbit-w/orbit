//go:build arm64
// +build arm64

#include "textflag.h"

#define MAXSPIN 4000  // ARM64适中的自旋次数

// func spin64(l *uint64)
TEXT ·spin64(SB), NOSPLIT, $0-8
    MOVD ptr+0(FP), R0    // 将指针参数加载到R0寄存器
    MOVW $MAXSPIN, R1     // 设置自旋计数器
    
spin:
    // 使用LDAR（Load-Acquire Register）进行原子读取
    LDAR (R0), R2
    
    // 检查值是否为0
    CBZ R2, acquire       // 如果为0，锁可用
    
    // 递减计数器并检查
    SUBW $1, R1
    YIELD                 // ARM64自旋优化指令
    CBNZ R1, spin        // 如果计数器不为0，继续自旋
    
acquire:
    // 轻量级内存屏障：只保证Load-Acquire语义
    // LDAR已经提供了acquire语义，只需要轻量屏障
    DMB $11              // DMB LD - 只同步加载操作，比DMB SY轻量
    
    RET 
