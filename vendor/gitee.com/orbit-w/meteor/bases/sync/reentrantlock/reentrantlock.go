package reentrantlock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"gitee.com/orbit-w/meteor/bases/misc/g"
)

const (
	lockBits     = 16                 // 16位重入计数，支持65,535次重入
	countMask    = 0xFFFF000000000000 // 16位计数掩码，用于提取重入计数
	goIDMask     = 0x0000FFFFFFFFFFFF // 48位GOID掩码，用于提取goroutine ID
	goIDBits     = 64 - lockBits      // 48位GOID，支持281万亿个goroutine
	maxLockCount = 1 << lockBits      // 最大重入次数: 65,536

	spinlockSpinSteps      = 20               // CPU自旋步数，前20次使用CPU自旋
	spinlockSleepTime      = time.Millisecond // 调度器让出时间，1毫秒
	tryLockWithSpinTryLoop = 100              // TryLockWithSpin的最大自旋次数
	spinlockWaitSchedule   = 20000            // spinlockWaitSchedule 最大等待调度次数，20000次
)

var (
	// spinlockBusy 锁忙碌阈值，3秒，超过此时间认为可能死锁
	spinlockBusy = uint64(time.Second.Nanoseconds() * 10)

	// forceLockTimeout 强制超时阈值，15秒，超过此时间强制panic
	forceLockTimeout = spinlockBusy * 5
)

// ReentrantLock 是一个可重入锁
// 在同一个goroutine 内可以重复加锁，必须在同一goroutine内解锁。
//
// 位分配设计 (48+16):
//   - 48位GOID: 支持281万亿个goroutine，按100万/秒创建可运行8,925年
//   - 16位重入计数: 支持65,535次重入，基于以下现实递归深度分析
//
// 现实递归深度分析 (基于Go栈约束和实际使用场景):
//
//	递归类别      典型深度    最大深度    栈消耗/次    总栈需求    现实性
//	算法递归      100        1,000      512B        0.5MB      ✅ 现实
//	业务递归      50         500        1KB         0.5MB      ✅ 现实
//	深度递归      1,000      10,000     2KB         20MB       ✅ 现实
//	极限递归      10,000     65,535     4KB         256MB      ❌ 栈溢出
//
// Go栈约束分析:
//   - 初始栈: 2KB, 动态扩展(2倍增长)
//   - 理论最大栈: 1GB (64位系统)
//   - 实际建议最大: 100MB (避免OOM)
//   - 结论: 16位计数(6.5万次)覆盖99.99%现实递归需求，提供合理安全余量
type ReentrantLock struct {
	lock       uint64 // 位打包: 高16位=重入计数, 低48位=goroutine ID
	lockedTime uint64 // 加锁时间戳(纳秒), 用于死锁检测
}

// adjustReentrantCount 调整重入计数
// 参数: lock - 当前锁值, offset - 计数偏移量(通常为1或-1)
// 返回: 新的锁值 (保持GOID不变，更新重入计数)
// 注意: 当计数溢出时会panic
func adjustReentrantCount(lock uint64, offset int64) uint64 {
	offset += int64(lock >> goIDBits)
	if offset >= maxLockCount {
		panic(fmt.Sprintf("ReentrantLock lock overflow: %d", offset))
	}
	return (lock & goIDMask) | (uint64(offset) << goIDBits)
}

// lockerToGoID 从锁值中提取goroutine ID
// 参数: locker - 锁值
// 返回: goroutine ID (低48位)
func lockerToGoID(locker uint64) uint64 {
	return locker & goIDMask
}

// count 从锁值中提取重入次数
// 参数: locker - 锁值
// 返回: 重入次数 (0表示未加锁，>0表示重入次数)
func count(locker uint64) int {
	if locker == 0 {
		return 0
	}
	return int(locker>>goIDBits) + 1
}

// spin 自旋等待策略
// 参数: loop - 当前循环次数
// 逻辑: 前40次使用CPU自旋，之后使用调度器让出CPU
func (rl *ReentrantLock) spin(loop int) {
	if loop < spinlockSpinSteps {
		spin64(&rl.lock) // CPU自旋，适合短时间等待
	} else {
		// 处理了"锁很快就会被释放"的超低延迟场景。
		// 此时通过Sleep一段时间让出CPU，可以避免忙等。
		// 保证系统在高竞争下的稳定性。
		gPM.RecordConflict(LockTypeReentrant, ConflictTypeScheduler)
		time.Sleep(spinlockSleepTime)
	}
}

// checkTimeout 检查锁是否超时
// now: 当前时间戳(纳秒)
// locked: 锁定的时间戳(纳秒)
// timeoutThreshold: 超时阈值(纳秒)
// 返回: (是否超时, 错误)
func (rl *ReentrantLock) checkTimeout(now, timeoutThreshold uint64) error {
	locked := atomic.LoadUint64(&rl.lockedTime)
	if locked == 0 {
		return nil
	}
	if now > locked+timeoutThreshold {
		gPM.RecordTimeoutLock()
		cur := atomic.LoadUint64(&rl.lock)
		msg := fmt.Sprintf("ReentrantLock spinlock owner locked at:%d[%d], %s", lockerToGoID(cur), count(cur), time.Duration(now-locked).String())
		return errors.New(msg)
	}
	return nil
}

// setLockedTime 设置加锁时间戳，用于死锁检测
func (rl *ReentrantLock) setLockedTime(lTime uint64) {
	rl.lockedTime = lTime
}

// tryReentrantLock 尝试重入加锁
// 如果当前 goroutine 已经持有锁，则增加重入计数并返回 true
func (rl *ReentrantLock) tryReentrantLock(goid uint64) bool {
	cur := atomic.LoadUint64(&rl.lock)
	if lockerToGoID(cur) == goid {
		atomic.StoreUint64(&rl.lock, adjustReentrantCount(cur, 1))
		gPM.RecordSuccessfulLock()
		return true
	}
	return false
}

// tryAcquireFastPath 尝试通过快速路径获取锁。
// 这是在锁未被占用的情况下的优化。
// 如果成功获取锁，它会更新加锁时间并记录成功事件。
// 如果提供了 startTime，它还会记录等待时间。
func (rl *ReentrantLock) tryAcquireFastPath(goid uint64, startTime *time.Time) bool {
	if atomic.CompareAndSwapUint64(&rl.lock, 0, goid) {
		rl.setLockedTime(uint64(time.Now().UnixNano()))
		gPM.RecordSuccessfulLock()
		if startTime != nil {
			gPM.RecordWaitTime(time.Since(*startTime))
		}
		return true
	}
	return false
}

// Lock 阻塞式加锁，支持重入
// 加锁逻辑:
//  1. 尝试直接获取锁(无竞争情况)
//  2. 检查是否为同一goroutine的重入调用
//  3. 如果是其他goroutine持有锁，则进入自旋等待
//  4. 包含死锁检测和性能监控
//  5. 快速失败：如果超时，强制程序崩溃，需要通过堆栈跟踪（stack trace）来定位死锁问题
//
// 注意: 必须在同一goroutine内调用Unlock()
func (rl *ReentrantLock) Lock() {
	goid := uint64(g.ID())
	n := uint64(time.Now().UnixNano())

	// 快速路径: 尝试直接获取锁
	if rl.tryAcquireFastPath(goid, nil) {
		return
	}

	// 检查重入: 如果是同一goroutine，增加计数
	if rl.tryReentrantLock(goid) {
		return
	}

	// 慢速路径: 其他goroutine持有锁，进入竞争处理
	gPM.RecordConflict(LockTypeReentrant, ConflictTypeBasic)

	// 死锁检测: 检查锁持有时间是否超过强制超时
	if err := rl.checkTimeout(n, forceLockTimeout); err != nil {
		panic(err)
	}

	// 自旋等待获取锁
	for j := range spinlockWaitSchedule {
		// 并发的goroutine goid 应该是唯一的
		// 继续尝试用goid获取锁，不需要考虑重入的问题
		if rl.tryAcquireFastPath(goid, nil) {
			return
		}
		rl.spin(j)
	}

	// 自旋超时，触发panic
	rl.panicOnTimeout()
}

// TryLockWithSpin 非阻塞式加锁，支持重入，失败时会自旋一定次数
// 返回: true-成功获取锁, false-获取失败
// 特点:
//
//	1.相比TryLock会进行有限次数的自旋等待，提高成功率。
//	2.非阻塞式，不会panic，是 Lock 轻度"安全"版本
func (rl *ReentrantLock) TryLockWithSpin() bool {
	gPM.RecordLockAttempt()
	goid := uint64(g.ID())
	if rl.tryAcquireFastPath(goid, nil) {
		return true
	}

	if rl.tryReentrantLock(goid) {
		return true
	}

	gPM.RecordConflict(LockTypeReentrant, ConflictTypeBasic)
	for j := range tryLockWithSpinTryLoop {
		if rl.tryAcquireFastPath(goid, nil) {
			return true
		}
		rl.spin(j)
	}
	return false
}

// TryLock 非阻塞式加锁，支持重入，立即返回
// 返回: true-成功获取锁, false-获取失败
// 特点: 不会等待，立即返回结果，适合不希望阻塞的场景
func (rl *ReentrantLock) TryLock() bool {
	gPM.RecordLockAttempt()

	goid := uint64(g.ID())
	if rl.tryAcquireFastPath(goid, nil) {
		return true
	}

	return rl.tryReentrantLock(goid)
}

// TryLockWithTimeout 带超时的加锁，支持重入
// 参数: d - 超时时间
// 返回: nil-成功, error-失败(超时或死锁)
func (rl *ReentrantLock) TryLockWithTimeout(d time.Duration) error {
	gPM.RecordLockAttempt()

	goid := uint64(g.ID())
	if rl.tryAcquireFastPath(goid, nil) {
		return nil
	}

	if rl.tryReentrantLock(goid) {
		return nil
	}

	// 死锁检测
	// TODO: 其他goroutine 持有锁超时，此时是否需要panic？或者有没有更好的处理方式？
	last := time.Now().UnixNano()
	if err := rl.checkTimeout(uint64(last), spinlockBusy); err != nil {
		return err
	}

	return rl.tryLockWithTimeoutSlowPath(d, goid)
}

// TryLockByCtx 基于Context的加锁，支持重入和取消
// 参数: ctx - 上下文，用于取消和超时控制
// 返回: nil-成功, error-失败(取消/超时/死锁)
func (rl *ReentrantLock) TryLockByCtx(ctx context.Context) error {
	startTime := time.Now()
	gPM.RecordLockAttempt()

	goid := uint64(g.ID())
	if rl.tryAcquireFastPath(goid, &startTime) {
		return nil
	}

	if rl.tryReentrantLock(goid) {
		return nil
	}

	if err := rl.checkTimeout(uint64(time.Now().UnixNano()), spinlockBusy); err != nil {
		return err
	}

	if err := rl.tryLockWithCtxSlowPath(ctx, goid); err != nil {
		cur := atomic.LoadUint64(&rl.lock)
		gPM.RecordWaitTime(time.Since(startTime))
		lockedTime := atomic.LoadUint64(&rl.lockedTime)
		return fmt.Errorf("spinlock owner locked at:%d[%d], deadlock:%s, err:%s", lockerToGoID(cur), count(cur), time.Duration(uint64(time.Now().UnixNano())-lockedTime).String(), err.Error())
	}

	return nil
}

func (rl *ReentrantLock) tryLockWithCtxSlowPath(ctx context.Context, goid uint64) error {
	gPM.RecordConflict(LockTypeReentrant, ConflictTypeBasic)
	for i := range spinlockWaitSchedule {
		if rl.tryAcquireFastPath(goid, nil) {
			return nil
		}

		select {
		case <-ctx.Done():
			if os.IsTimeout(ctx.Err()) {
				return fmt.Errorf("ReentrantLock try lock timeout")
			} else {
				return fmt.Errorf("ReentrantLock lock error:%w", ctx.Err())
			}
		default:
			rl.spin(i)
		}
	}
	return nil
}

// tryLockWithTimeoutSlowPath 带超时加锁的内部实现
// 参数: d - 超时时间, goid - goroutine ID
// 返回: nil-成功, error-失败
func (rl *ReentrantLock) tryLockWithTimeoutSlowPath(d time.Duration, goid uint64) error {
	gPM.RecordConflict(LockTypeReentrant, ConflictTypeBasic)
	timer := time.NewTimer(d)
	defer timer.Stop()

	//TODO: 这里可以优化，如果超时时间过长，自旋次数过多，可以考虑使用信号量来控制自旋次数？或者依赖spin中sleep来控制自旋次数？
	for i := 0; ; i++ {
		if rl.tryAcquireFastPath(goid, nil) {
			return nil
		}

		// 检查超时或执行自旋
		select {
		case <-timer.C:
			locked := atomic.LoadUint64(&rl.lock)
			lockedTime := atomic.LoadUint64(&rl.lockedTime)
			gPM.RecordTimeoutLock()
			return fmt.Errorf("ReentrantLock spinlock owner locked at:%d[%d], %s", lockerToGoID(locked), count(locked), time.Duration(uint64(time.Now().UnixNano())-lockedTime).String())
		default:
			rl.spin(i)
		}
	}
}

// Unlock 解锁，支持重入
// 解锁逻辑:
//  1. 如果重入计数为1，直接释放锁
//  2. 如果重入计数>1，减少计数但保持锁定状态
//  3. 检查调用goroutine是否为锁的持有者
//
// 注意: 必须在持有锁的同一goroutine内调用
func (rl *ReentrantLock) Unlock() {
	goid := uint64(g.ID())
	cur := atomic.LoadUint64(&rl.lock)

	ownerID := lockerToGoID(cur)
	// 安全检查：确保在正确的goroutine内解锁
	if ownerID != goid {
		if ownerID == 0 {
			panic("ReentrantLock: unlock of unlocked lock")
		}
		panic(fmt.Sprintf("ReentrantLock: unlock in different goroutine (caller: %d, owner: %d)", goid, ownerID))
	}

	// 检查是否为单次加锁(重入计数为1)
	if cur == goid {
		// 单次加锁，直接释放
		rl.setLockedTime(0)
		atomic.StoreUint64(&rl.lock, 0)
	} else {
		// 多次重入，减少计数
		cur = adjustReentrantCount(cur, -1)
		atomic.StoreUint64(&rl.lock, cur)
	}
}

// IsLocked 检查当前goroutine是否持有锁
// 返回: true-当前goroutine持有锁, false-未持有
func (rl *ReentrantLock) IsLocked() bool {
	goid := uint64(g.ID())
	cur := atomic.LoadUint64(&rl.lock)
	return lockerToGoID(cur) == goid
}

// IsLockedByOther 检查锁是否被其他goroutine持有
// 返回: true-被其他goroutine持有, false-未被持有或被当前goroutine持有
func (rl *ReentrantLock) IsLockedByOther() bool {
	goid := uint64(g.ID())
	cur := lockerToGoID(atomic.LoadUint64(&rl.lock))
	return cur != 0 && cur != goid
}

func (rl *ReentrantLock) panicOnTimeout() {
	gPM.RecordTimeoutLock()
	cur := atomic.LoadUint64(&rl.lock)
	panic(fmt.Sprintf("【Spinlock】 lock timeout, owner: %d, count: %d", lockerToGoID(cur), count(cur)))
}
