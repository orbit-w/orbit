//go:build amd64 || arm64 || 386 || riscv64
// +build amd64 arm64 386 riscv64

package reentrantlock

//go:noescape
func spin64(l *uint64)
