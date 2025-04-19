package actor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

const (
	ManagerName = "system-actor-supervision"

	ActorSystemStateRunning  = 0
	ActorSystemStateStopping = 1
	ActorSystemStateStopped  = 2
)

var (
	System              *ActorSystem
	once                sync.Once
	ActorFacadeStopping atomic.Bool
)

func init() {
	once.Do(func() {
		actorsCache = NewActorsCache()
	})
}

type IService interface {
	Start() error
	Stop(ctx context.Context) error
}

// ActorSystem provides a simplified interface for managing actors
type ActorSystem struct {
	state       atomic.Int32
	actorSystem *actor.ActorSystem
	supervisors []*actor.PID
}

func (af *ActorSystem) Start() error {
	system := actor.NewActorSystem()
	af.actorSystem = system
	af.supervisors = make([]*actor.PID, LevelMaxLimit)
	for lv := LevelNormal; lv < LevelMaxLimit; lv++ {
		af.supervisors[lv] = newSupervisor(system, lv)
	}
	System = af
	return nil
}

// Stop 开始ActorSystem的优雅关闭流程
// 参数:
//   - ctx: 用于控制停止操作超时和取消的上下文
//
// 返回:
//   - 如果系统已在关闭或关闭成功则返回nil，如果发生错误则返回错误
//
// 说明:
//   - 此方法只有在系统处于运行状态时才会进行实际关闭
//   - 使用CAS操作确保只有一个goroutine能够触发系统关闭
func (af *ActorSystem) Stop(ctx context.Context) error {
	if af.state.CompareAndSwap(ActorSystemStateRunning, ActorSystemStateStopping) {
		return af.stop(ctx)
	}
	return nil
}

// stop 负责处理ActorSystem的优雅关闭流程
// 参数:
//   - ctx: 用于控制停止操作超时和取消的上下文
//
// 返回:
//   - 如果停止操作成功完成则返回nil，如果发生错误则返回错误
//
// 实现细节:
//   - 使用GracefulShutdownManager确保所有supervisor正确停止
//   - 要求所有supervisor连续5次成功报告完成才视为停止成功
//   - 设置最大尝试次数为100，防止永远无法达到成功阈值时的无限重试
//   - 当所有supervisor停止后，将系统状态设置为Stopped
func (af *ActorSystem) stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ActorFacadeStopping.Store(true)

	// 创建一个有最大尝试次数限制的关闭管理器
	const (
		successThreshold = 5   // 需要连续成功的次数
		maxAttempts      = 100 // 最大尝试次数，防止无限循环
	)

	shutdownManager := NewGracefulShutdownManagerWithMaxAttempts(successThreshold, maxAttempts, func() bool {
		for lv := LevelNormal; lv < LevelMaxLimit; lv++ {
			completed, err := af.stopSupervisor(lv)
			if err != nil {
				logger.GetLogger().Error("Failed to stop supervisor",
					zap.Int32("level", int32(lv)),
					zap.Error(err))
				return false
			}

			if !completed {
				logger.GetLogger().Info("supervisor not completed",
					zap.Int32("level", int32(lv)))
				return false
			}
		}
		logger.GetLogger().Info("All supervisors stopped complete")
		return true
	})

	err := shutdownManager.Shutdown(ctx)

	// 无论结果如何，都将状态设置为Stopped
	af.state.CompareAndSwap(ActorSystemStateStopping, ActorSystemStateStopped)

	return err
}

func (af *ActorSystem) ActorSystem() *actor.ActorSystem {
	return af.actorSystem
}

func (af *ActorSystem) stopSupervisor(lv Level) (bool, error) {
	result, err := retry(func() (any, error) {
		future := af.actorSystem.Root.RequestFuture(af.supervisors[lv], &StopAllRequest{}, 30*time.Second)
		result, err := future.Result()
		return result, err
	}, 10)
	if err != nil {
		return false, err
	}

	resp := result.(*StopAllResponse)
	return resp.Complete, nil
}

// NewActorFacade creates a new instance of ActorFacade
func NewActorFacade(actorSystem *actor.ActorSystem) *ActorSystem {
	af := &ActorSystem{
		actorSystem: actorSystem,
		supervisors: make([]*actor.PID, LevelMaxLimit),
	}

	for lv := LevelNormal; lv < LevelMaxLimit; lv++ {
		af.supervisors[lv] = newSupervisor(actorSystem, lv)
	}

	return af
}

func newSupervisor(actorSystem *actor.ActorSystem, level Level) *actor.PID {
	decider := func(reason any) actor.Directive {
		return actor.ResumeDirective
	}
	supervisor := actor.NewOneForOneStrategy(10, 1000, decider)
	producer := func() actor.Actor {
		return NewActorSupervision(actorSystem, level)
	}
	props := actor.PropsFromProducer(producer, actor.WithSupervisor(supervisor))

	managerPID, err := actorSystem.Root.SpawnNamed(props, GenManagerName(level))
	if err != nil {
		panic(err) // In a real application, handle this error appropriately
	}

	return managerPID
}

// GetOrStartActor 获取一个就绪的Actor对象
func GetOrStartActor(actorName, pattern string, props *Props) (*Process, error) {
	// First check if manager already has this actor
	if actor, exists := actorsCache.Get(actorName); exists {
		if !actor.IsStopped() {
			return actor, nil
		}
	}

	system := System.actorSystem
	future := actor.NewFuture(system, ManagerStartActorFutureTimeout)
	mPid := System.supervisorByPattern(pattern)
	rf := system.Root.RequestFuture(mPid, &StartActorRequest{
		ActorName: actorName,
		Pattern:   pattern,
		Future:    future.PID(),
		Props:     props,
	}, StartActorTimeout)

	result, err := waitFuture(rf)
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case *Process:
		return v, nil
	case *StartActorWait:
		result, err = waitFuture(future)
		if err != nil {
			return nil, err
		}
		switch v := result.(type) {
		case *Process:
			return v, nil
		default:
			return nil, errors.New("unknown result type")
		}
	default:
		return nil, errors.New("unknown result type")
	}
}

// StopActor stops the actor with the given ID
func StopActor(actorName, pattern string) error {
	result, err := System.RequestFuture(pattern, &StopActorMessage{
		ActorName: actorName,
		Pattern:   pattern,
	}, StopActorTimeout)
	if err != nil {
		return err
	}

	if err, ok := result.(error); ok {
		return err
	}

	return nil
}

func (f *ActorSystem) RequestFuture(pattern string, msg any, timeout time.Duration) (any, error) {
	// Send message to manager to start the actor
	mPid := f.supervisorByPattern(pattern)
	future := f.actorSystem.Root.RequestFuture(mPid, msg, timeout)

	result, err := future.Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (f *ActorSystem) supervisorByPattern(pattern string) *actor.PID {
	level := GetLevelByPattern(pattern)
	return f.supervisorByLevel(level)
}

func (f *ActorSystem) supervisorByLevel(level Level) *actor.PID {
	return f.supervisors[level]
}

func waitFuture(future *actor.Future) (any, error) {
	result, err := future.Result()
	if err != nil {
		return nil, err
	}

	if err, ok := result.(error); ok {
		return nil, err
	}

	return result, nil
}

func retry(fn func() (any, error), retryCount int) (any, error) {
	var lastErr error
	for i := 0; i < retryCount; i++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, errors.New("max retry attempts reached: " + lastErr.Error())
}

// StopWithTimeout is a convenience method that stops the actor system with a timeout
// It creates a context with the specified timeout and calls Stop
func (af *ActorSystem) StopWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return af.Stop(ctx)
}

// StopWithDefaultTimeout is a convenience method that stops the actor system with a default timeout of 60 seconds
func (af *ActorSystem) StopWithDefaultTimeout() error {
	return af.StopWithTimeout(60 * time.Second)
}

// GracefulShutdownManager 负责管理优雅关闭过程，通过反复尝试特定操作直到连续成功达到阈值次数
// 用于确保系统组件可以安全、可靠地关闭，即使在分布式或并发环境中
type GracefulShutdownManager struct {
	// 需要连续成功的次数，达到此阈值表示操作完全成功
	successThreshold int32
	// 当前连续成功的次数
	successCount atomic.Int32
	// 总尝试次数，用于监控和调试
	attemptCount atomic.Int32
	// 最大尝试次数，超过此次数后不再重试，0表示无限重试
	maxAttempts int32
	// 关闭操作的实际执行函数，返回操作是否成功
	shutdownOperation func() bool
	// 通知通道，用于发出关闭完成的信号
	completionSignal chan struct{}
}

// NewGracefulShutdownManager 创建一个新的优雅关闭管理器
// 参数:
//   - successThreshold: 确认操作完成所需的连续成功次数
//   - shutdownOperation: 实际执行关闭操作并返回是否成功的函数
func NewGracefulShutdownManager(successThreshold int32, shutdownOperation func() bool) *GracefulShutdownManager {
	return NewGracefulShutdownManagerWithMaxAttempts(successThreshold, 0, shutdownOperation)
}

// NewGracefulShutdownManagerWithMaxAttempts 创建一个有最大尝试次数限制的优雅关闭管理器
// 参数:
//   - successThreshold: 确认操作完成所需的连续成功次数
//   - maxAttempts: 最大尝试次数，0表示无限重试
//   - shutdownOperation: 实际执行关闭操作并返回是否成功的函数
func NewGracefulShutdownManagerWithMaxAttempts(successThreshold int32, maxAttempts int32, shutdownOperation func() bool) *GracefulShutdownManager {
	return &GracefulShutdownManager{
		successThreshold:  successThreshold,
		successCount:      atomic.Int32{},
		attemptCount:      atomic.Int32{},
		maxAttempts:       maxAttempts,
		shutdownOperation: shutdownOperation,
		completionSignal:  make(chan struct{}, 1),
	}
}

// done 返回一个接收通道，在关闭完成时会被关闭
// 可用于等待关闭过程完成
func (g *GracefulShutdownManager) done() <-chan struct{} {
	return g.completionSignal
}

// signalCompletion 关闭完成信号通道，表示关闭过程已完成
func (g *GracefulShutdownManager) signalCompletion() {
	close(g.completionSignal)
}

// Shutdown 开始并管理关闭过程
// 参数:
//   - ctx: 用于控制关闭超时和取消的上下文
//
// 返回:
//   - 如果关闭成功完成则返回nil，否则返回错误
//
// 注意:
//   - 此方法会阻塞直到关闭完成或上下文被取消
//   - 内部使用goroutine执行反复尝试的关闭操作，避免阻塞主线程
func (g *GracefulShutdownManager) Shutdown(ctx context.Context) error {
	// 创建一个衍生的 context，在函数退出时可以取消，确保 goroutine 不会泄漏
	shutdownCtx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc() // 确保在函数退出时取消 context

	utils.GoRecoverPanic(func() {
		defer g.signalCompletion()

		for g.successCount.Load() < g.successThreshold {
			// 检查最大尝试次数
			if g.maxAttempts > 0 && g.attemptCount.Load() >= g.maxAttempts {
				logger.GetLogger().Warn("reached maximum attempt count without success",
					zap.Int32("maxAttempts", g.maxAttempts),
					zap.Int32("successCount", g.successCount.Load()),
					zap.Int32("successThreshold", g.successThreshold))
				return
			}

			// 检查 context 是否已取消
			select {
			case <-shutdownCtx.Done():
				return
			default:
				// 继续执行
			}

			g.attemptCount.Add(1)

			// 执行实际的关闭操作
			operationSucceeded := g.shutdownOperation()

			if operationSucceeded {
				// 如果操作成功，增加连续成功计数
				g.successCount.Add(1)
			} else {
				// 如果操作失败，重置连续成功计数
				g.successCount.Store(0)
				// 短暂等待后再次尝试，给系统一些恢复时间
				select {
				case <-shutdownCtx.Done():
					return
				case <-time.After(50 * time.Millisecond):
					// 继续尝试
				}
			}
		}
	})

	// 等待关闭完成或上下文取消
	select {
	case <-g.done():
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return fmt.Errorf("shutdown operation canceled: %w", ctx.Err())
		default:
		}

		// 检查是否达到了成功阈值或是因为达到最大尝试次数而退出
		if g.successCount.Load() >= g.successThreshold {
			logger.GetLogger().Info("shutdown completed successfully",
				zap.Int32("attempts", g.attemptCount.Load()),
				zap.Int32("successCount", g.successCount.Load()))
			return nil
		} else {
			return fmt.Errorf("shutdown operation failed after %d attempts, only reached %d/%d consecutive successes",
				g.attemptCount.Load(), g.successCount.Load(), g.successThreshold)
		}
	case <-ctx.Done():
		return fmt.Errorf("shutdown operation canceled: %w", ctx.Err())
	}
}
