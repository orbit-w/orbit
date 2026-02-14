package reentrantlock

import (
	"fmt"
	"sync/atomic"
	"time"
)

// LockType 锁类型枚举
type LockType int

const (
	LockTypeSpinlock LockType = iota
	LockTypeReentrant
)

// ConflictType 冲突类型枚举
type ConflictType int

const (
	ConflictTypeBasic     ConflictType = iota // 基础冲突
	ConflictTypeScheduler                     // 调度器级别冲突
)

// PerformanceMonitor 性能监控器
// 统一管理锁的性能指标和冲突统计
type PerformanceMonitor struct {
	// 基础冲突计数器
	spinlockConflicts          int64 // 普通自旋锁冲突
	reentrantSpinlockConflicts int64 // 可重入锁冲突

	// 调度器级别冲突计数器
	spinlockSchedulerConflicts  int64 // 普通自旋锁调度器冲突
	reentrantSchedulerConflicts int64 // 可重入锁调度器冲突

	// 性能统计
	totalLockAttempts int64 // 总加锁尝试次数
	successfulLocks   int64 // 成功加锁次数
	timeoutLocks      int64 // 超时加锁次数

	// 时间统计
	totalWaitTime int64 // 总等待时间(纳秒)
	maxWaitTime   int64 // 最大等待时间(纳秒)

	// 启动时间
	startTime time.Time
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		startTime: time.Now(),
	}
}

// RecordConflict 记录冲突
func (pm *PerformanceMonitor) RecordConflict(lockType LockType, conflictType ConflictType) {
	switch lockType {
	case LockTypeSpinlock:
		switch conflictType {
		case ConflictTypeBasic:
			atomic.AddInt64(&pm.spinlockConflicts, 1)
		case ConflictTypeScheduler:
			atomic.AddInt64(&pm.spinlockSchedulerConflicts, 1)
		}
	case LockTypeReentrant:
		switch conflictType {
		case ConflictTypeBasic:
			atomic.AddInt64(&pm.reentrantSpinlockConflicts, 1)
		case ConflictTypeScheduler:
			atomic.AddInt64(&pm.reentrantSchedulerConflicts, 1)
		}
	}
}

// RecordLockAttempt 记录加锁尝试
func (pm *PerformanceMonitor) RecordLockAttempt() {
	atomic.AddInt64(&pm.totalLockAttempts, 1)
}

// RecordSuccessfulLock 记录成功加锁
func (pm *PerformanceMonitor) RecordSuccessfulLock() {
	atomic.AddInt64(&pm.successfulLocks, 1)
}

// RecordTimeoutLock 记录超时加锁
func (pm *PerformanceMonitor) RecordTimeoutLock() {
	atomic.AddInt64(&pm.timeoutLocks, 1)
}

// RecordWaitTime 记录等待时间
func (pm *PerformanceMonitor) RecordWaitTime(waitTime time.Duration) {
	waitNanos := waitTime.Nanoseconds()
	atomic.AddInt64(&pm.totalWaitTime, waitNanos)

	// 更新最大等待时间
	for {
		current := atomic.LoadInt64(&pm.maxWaitTime)
		if waitNanos <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.maxWaitTime, current, waitNanos) {
			break
		}
	}
}

// GetConflictStats 获取冲突统计
func (pm *PerformanceMonitor) GetConflictStats() ConflictStats {
	return ConflictStats{
		SpinlockConflicts:           atomic.LoadInt64(&pm.spinlockConflicts),
		ReentrantSpinlockConflicts:  atomic.LoadInt64(&pm.reentrantSpinlockConflicts),
		SpinlockSchedulerConflicts:  atomic.LoadInt64(&pm.spinlockSchedulerConflicts),
		ReentrantSchedulerConflicts: atomic.LoadInt64(&pm.reentrantSchedulerConflicts),
	}
}

// GetPerformanceStats 获取性能统计
func (pm *PerformanceMonitor) GetPerformanceStats() PerformanceStats {
	totalAttempts := atomic.LoadInt64(&pm.totalLockAttempts)
	successfulLocks := atomic.LoadInt64(&pm.successfulLocks)
	timeoutLocks := atomic.LoadInt64(&pm.timeoutLocks)
	totalWaitTime := atomic.LoadInt64(&pm.totalWaitTime)
	maxWaitTime := atomic.LoadInt64(&pm.maxWaitTime)

	var successRate float64
	var avgWaitTime time.Duration

	if totalAttempts > 0 {
		successRate = float64(successfulLocks) / float64(totalAttempts) * 100
	}

	if successfulLocks > 0 {
		avgWaitTime = time.Duration(totalWaitTime / successfulLocks)
	}

	return PerformanceStats{
		TotalLockAttempts: totalAttempts,
		SuccessfulLocks:   successfulLocks,
		TimeoutLocks:      timeoutLocks,
		SuccessRate:       successRate,
		TotalWaitTime:     time.Duration(totalWaitTime),
		AverageWaitTime:   avgWaitTime,
		MaxWaitTime:       time.Duration(maxWaitTime),
		Uptime:            time.Since(pm.startTime),
	}
}

// Reset 重置所有计数器
func (pm *PerformanceMonitor) Reset() {
	atomic.StoreInt64(&pm.spinlockConflicts, 0)
	atomic.StoreInt64(&pm.reentrantSpinlockConflicts, 0)
	atomic.StoreInt64(&pm.spinlockSchedulerConflicts, 0)
	atomic.StoreInt64(&pm.reentrantSchedulerConflicts, 0)
	atomic.StoreInt64(&pm.totalLockAttempts, 0)
	atomic.StoreInt64(&pm.successfulLocks, 0)
	atomic.StoreInt64(&pm.timeoutLocks, 0)
	atomic.StoreInt64(&pm.totalWaitTime, 0)
	atomic.StoreInt64(&pm.maxWaitTime, 0)
	pm.startTime = time.Now()
}

// ConflictStats 冲突统计数据
type ConflictStats struct {
	SpinlockConflicts           int64 // 普通自旋锁冲突次数
	ReentrantSpinlockConflicts  int64 // 可重入锁冲突次数
	SpinlockSchedulerConflicts  int64 // 普通自旋锁调度器冲突次数
	ReentrantSchedulerConflicts int64 // 可重入锁调度器冲突次数
}

// String 格式化输出冲突统计
func (cs ConflictStats) String() string {
	return fmt.Sprintf(
		"ConflictStats{Spinlock: %d, ReentrantSpinlock: %d, SpinlockScheduler: %d, ReentrantScheduler: %d}",
		cs.SpinlockConflicts,
		cs.ReentrantSpinlockConflicts,
		cs.SpinlockSchedulerConflicts,
		cs.ReentrantSchedulerConflicts,
	)
}

// TotalConflicts 总冲突次数
func (cs ConflictStats) TotalConflicts() int64 {
	return cs.SpinlockConflicts + cs.ReentrantSpinlockConflicts +
		cs.SpinlockSchedulerConflicts + cs.ReentrantSchedulerConflicts
}

// PerformanceStats 性能统计数据
type PerformanceStats struct {
	TotalLockAttempts int64         // 总加锁尝试次数
	SuccessfulLocks   int64         // 成功加锁次数
	TimeoutLocks      int64         // 超时加锁次数
	SuccessRate       float64       // 成功率(%)
	TotalWaitTime     time.Duration // 总等待时间
	AverageWaitTime   time.Duration // 平均等待时间
	MaxWaitTime       time.Duration // 最大等待时间
	Uptime            time.Duration // 运行时间
}

// String 格式化输出性能统计
func (ps PerformanceStats) String() string {
	return fmt.Sprintf(
		"PerformanceStats{Attempts: %d, Success: %d, Timeout: %d, SuccessRate: %.2f%%, "+
			"AvgWait: %v, MaxWait: %v, Uptime: %v}",
		ps.TotalLockAttempts,
		ps.SuccessfulLocks,
		ps.TimeoutLocks,
		ps.SuccessRate,
		ps.AverageWaitTime,
		ps.MaxWaitTime,
		ps.Uptime,
	)
}

// 全局性能监控器实例
var gPM = NewPerformanceMonitor()

func RecordReentrantConflict() {
	gPM.RecordConflict(LockTypeReentrant, ConflictTypeBasic)
}

func RecordReentrantSchedulerConflict() {
	gPM.RecordConflict(LockTypeReentrant, ConflictTypeScheduler)
}

func RecordSpinlockConflict() {
	gPM.RecordConflict(LockTypeSpinlock, ConflictTypeBasic)
}

func RecordSpinlockSchedulerConflict() {
	gPM.RecordConflict(LockTypeSpinlock, ConflictTypeScheduler)
}
