//go:build amd64
// +build amd64

#include "textflag.h"

#define MAXSPIN 1800  // 适中的自旋次数，平衡延迟和CPU使用

// func spin64(l *uint64)
TEXT ·spin64(SB), NOSPLIT, $0-8
    MOVQ ptr+0(FP), BP    // 使用BP寄存器（更符合调用约定）
    MOVL $MAXSPIN, DX
spin:
    MOVQ 0(BP), CX        // 读取锁值
    TESTQ CX, CX
    JZ acquire            // 锁可用，跳出
    DECL DX
    PAUSE                 // 在循环末尾PAUSE更高效
    JNZ spin
acquire:
    // 轻量级内存屏障：只在获取锁时使用
    LFENCE                // 比MFENCE更轻量，只保证读操作顺序
    RET

