//go:build riscv64
// +build riscv64

#include "textflag.h"

// func spin64(l *uint64)
TEXT ·spin64(SB), NOSPLIT, $0-8
    MOV ptr+0(FP), X10    // 将指针参数加载到X10寄存器
    
spin_loop:
    // RISC-V没有专门的暂停指令，使用NOP或轻量级指令
    // 使用fence指令来提供一些延迟
    FENCE
    
    // 使用原子加载指令
    MOV (X10), X11        // 加载64位值到X11
    
    // 检查值是否为0
    BNE X11, ZERO, spin_loop  // 如果不为0，继续自旋
    
    // 使用内存屏障确保内存操作顺序
    FENCE
    
    RET