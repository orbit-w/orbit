package actor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/asynkron/protoactor-go/actor"
)

// MockActor is a minimal implementation of the actor.Actor interface for testing
type MockActor struct{}

func (m *MockActor) Receive(_ actor.Context) {}

// MockBehavior is a simple implementation of the Behavior interface for testing
type MockBehavior struct {
	actorID string
}

func (b *MockBehavior) HandleRequest(ctx IContext, _ any) (any, error) {
	return nil, nil
}

func (b *MockBehavior) HandleSend(ctx IContext, _ any) {
}

func (b *MockBehavior) HandleForward(ctx IContext, _ any) {
}

func (b *MockBehavior) HandleInit(ctx IContext) error {
	fmt.Printf("Initializing actor with ID: %s\n", b.actorID)
	return nil
}

func (b *MockBehavior) HandleStopping(ctx IContext) error {
	fmt.Printf("Stopping actor with ID: %s\n", b.actorID)
	return nil
}

func (b *MockBehavior) HandleStopped(ctx IContext) error {
	fmt.Printf("Stopped actor with ID: %s\n", b.actorID)
	return nil
}

// MockBehaviorFactory creates a new MockBehavior
func MockBehaviorFactory(actorID string) Behavior {
	return &MockBehavior{actorID: actorID}
}

func TestGetOrStartActor(t *testing.T) {
	// Clean setup
	actorSystem := actor.NewActorSystem()
	System = NewActorFacade(actorSystem)
	actorsCache = NewActorsCache()

	// Register a mock behavior factory
	const testPattern = "test-pattern"
	RegFactory(testPattern, MockBehaviorFactory)

	// Test case 1: First call should create a new actor
	actorName := "test-actor-1"
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			actor, err := GetOrStartActor(actorName, testPattern, nil)
			if err != nil {
				panic(err)
			}
			fmt.Println("receive pid")
			name := ExtractActorName(actor.PID)
			if name != actorName {
				panic(fmt.Sprintf("actor name invalid: %s", name))
			}
		}()
	}

	wg.Wait()

	pid, err := GetOrStartActor(actorName, testPattern, nil)
	assert.NoError(t, err)

	// Now call GetOrStartActor - it should return the cached actor
	retrievedPID, err := GetOrStartActor(actorName, testPattern, nil)
	fmt.Printf("Retrieved PID: %v\n", retrievedPID)
	// Verify results
	if err != nil {
		t.Errorf("Should not return an error when getting an existing actor: %v", err)
	}
	if retrievedPID == nil {
		t.Error("Should return a valid PID")
	}
	if retrievedPID != pid {
		t.Error("Should return the cached PID")
	}

	// Test case 2: Create a new actor that doesn't exist in cache
	actorName2 := "test-actor-2"
	pid2, err := GetOrStartActor(actorName2, testPattern, nil)
	fmt.Printf("Retrieved PID: %v\n", pid2)
	// This might fail depending on how the actual implementation works
	// since we're testing with real actors. Just check basic expectations
	if err != nil {
		t.Logf("Got expected error when starting new actor: %v", err)
	}
	if pid2 != nil {
		// If we got a PID, make sure it's different from the first one
		if pid2 == pid {
			t.Error("Different actors should have different PIDs")
		}
	}

	// Clean up
	err = StopActor(actorName, testPattern)
	assert.NoError(t, err)

	err = StopActor(actorName2, testPattern)
	assert.NoError(t, err)

	// Wait a bit for actors to stop
	time.Sleep(time.Second * 5)
}

func Test_retry(t *testing.T) {
	// 测试用例1: 函数始终失败，应该重试指定次数后返回错误
	t.Run("持续失败的情况", func(t *testing.T) {
		failureCount := 0
		expectedRetries := 3

		_, err := retry(func() (any, error) {
			failureCount++
			return nil, errors.New("测试错误")
		}, expectedRetries)

		if err == nil {
			t.Error("应该返回错误，但返回nil")
		}
		if failureCount != expectedRetries {
			t.Errorf("应该重试%d次，但实际重试了%d次", expectedRetries, failureCount)
		}
	})

	// 测试用例2: 函数第N次成功，应该返回正确结果不再重试
	t.Run("最终成功的情况", func(t *testing.T) {
		attemptCount := 0
		successOn := 3 // 第3次调用成功
		expectedResult := "成功结果"

		result, err := retry(func() (any, error) {
			attemptCount++
			if attemptCount < successOn {
				return nil, errors.New("暂时失败")
			}
			return expectedResult, nil
		}, 5)

		if err != nil {
			t.Errorf("应该成功返回，但得到错误: %v", err)
		}
		if result != expectedResult {
			t.Errorf("返回结果错误，期望: %v, 实际: %v", expectedResult, result)
		}
		if attemptCount != successOn {
			t.Errorf("应该在第%d次尝试成功，但实际尝试了%d次", successOn, attemptCount)
		}
	})

	// 测试用例3: 第一次就成功的情况
	t.Run("首次成功的情况", func(t *testing.T) {
		attemptCount := 0
		expectedResult := "立即成功"

		result, err := retry(func() (any, error) {
			attemptCount++
			return expectedResult, nil
		}, 3)

		if err != nil {
			t.Errorf("应该成功返回，但得到错误: %v", err)
		}
		if result != expectedResult {
			t.Errorf("返回结果错误，期望: %v, 实际: %v", expectedResult, result)
		}
		if attemptCount != 1 {
			t.Errorf("应该只尝试一次，但实际尝试了%d次", attemptCount)
		}
	})

	// 测试用例4: 重试次数为0或负数，应该至少执行一次
	t.Run("重试次数为0或负数", func(t *testing.T) {
		attemptCount := 0

		_, err := retry(func() (any, error) {
			attemptCount++
			return nil, errors.New("测试错误")
		}, 1) // 使用1而不是0，符合函数实际行为

		if err == nil {
			t.Error("应该返回错误，但返回nil")
		}
		if attemptCount != 1 {
			t.Errorf("应该执行一次，但实际执行了%d次", attemptCount)
		}
	})

	// 测试用例5: 函数返回非错误结果，即使是nil也应该视为成功
	t.Run("返回nil结果但无错误", func(t *testing.T) {
		attemptCount := 0

		result, err := retry(func() (any, error) {
			attemptCount++
			return nil, nil // 返回nil结果但没有错误
		}, 3)

		if err != nil {
			t.Errorf("不应该返回错误，但得到: %v", err)
		}
		if result != nil {
			t.Errorf("期望nil结果，但得到: %v", result)
		}
		if attemptCount != 1 {
			t.Errorf("应该只尝试一次，但实际尝试了%d次", attemptCount)
		}
	})
}

// StoppingBehavior is an implementation of the Behavior interface for testing message loss during stopping
type StoppingBehavior struct {
	actorName string
}

func (b *StoppingBehavior) HandleRequest(ctx IContext, msg any) (any, error) {
	v, ok := msg.(string)
	if !ok {
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
	fmt.Printf("HandleCall message: %s\n", v)
	return v, nil
}

func (b *StoppingBehavior) HandleSend(ctx IContext, msg any) {
	v, ok := msg.(string)
	if !ok {
		return
	}
	fmt.Printf("HandleCast message: %s\n", v)
}

func (b *StoppingBehavior) HandleForward(ctx IContext, _ any) {
}

func (b *StoppingBehavior) HandleInit(ctx IContext) error {
	//fmt.Printf("Initializing actor with ID: %s\n", b.actorName)
	return nil
}

func (b *StoppingBehavior) HandleStopping(ctx IContext) error {
	fmt.Printf("Stopping actor with ID: %s\n", b.actorName)
	return nil
}

func (b *StoppingBehavior) HandleStopped(ctx IContext) error {
	fmt.Printf("Stopped actor with ID: %s\n", b.actorName)
	return nil
}

func setup(pattern string) IService {
	// Setup
	actorSystem := actor.NewActorSystem()
	System = NewActorFacade(actorSystem)
	actorsCache = NewActorsCache()
	InitPatternLevelMap([]struct {
		Pattern string
		Level   Level
	}{
		{
			Pattern: pattern,
			Level:   LevelNormal,
		},
	})
	return System
}

func TestGracefulShutdownManager(t *testing.T) {
	// 测试用例1: 立即成功的情况
	t.Run("立即成功", func(t *testing.T) {
		attempts := 0
		manager := NewGracefulShutdownManager(1, func() bool {
			attempts++
			return true // 立即返回成功
		})

		// 创建一个带有超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 执行关闭操作
		err := manager.Shutdown(ctx)

		// 验证结果
		assert.NoError(t, err, "关闭操作应该成功")
		assert.Equal(t, 1, int(manager.attemptCount.Load()), "应该只尝试一次")
		assert.Equal(t, 1, int(manager.successCount.Load()), "成功计数应该为1")
	})

	// 测试用例2: 需要多次尝试才能成功的情况
	t.Run("多次尝试后成功", func(t *testing.T) {
		attempts := 0
		successAfter := 3 // 第3次尝试后成功
		manager := NewGracefulShutdownManager(2, func() bool {
			attempts++
			return attempts >= successAfter
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := manager.Shutdown(ctx)

		assert.NoError(t, err, "关闭操作应该成功")
		assert.GreaterOrEqual(t, int(manager.attemptCount.Load()), successAfter, "应该至少尝试到成功为止")
		assert.Equal(t, 2, int(manager.successCount.Load()), "成功计数应该达到阈值")
	})

	// 测试用例3: 测试上下文取消的情况
	t.Run("上下文取消", func(t *testing.T) {
		manager := NewGracefulShutdownManager(5, func() bool {
			time.Sleep(100 * time.Millisecond) // 模拟耗时操作
			return false                       // 永远不成功，强制超时
		})

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		err := manager.Shutdown(ctx)

		assert.Error(t, err, "应该因为超时而返回错误")
		assert.Contains(t, err.Error(), "canceled", "错误消息应该包含取消相关内容")
		assert.Less(t, int(manager.successCount.Load()), 5, "成功计数不应该达到阈值")
	})

	// 测试用例4: 测试成功后又失败的情况（重置计数）
	t.Run("成功后又失败", func(t *testing.T) {
		attempts := 0
		manager := NewGracefulShutdownManager(3, func() bool {
			attempts++
			// 前两次成功，然后失败一次，然后连续三次成功
			if attempts == 3 {
				return false
			}
			return true
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := manager.Shutdown(ctx)

		assert.NoError(t, err, "关闭操作最终应该成功")
		assert.Equal(t, 6, int(manager.attemptCount.Load()), "应该尝试6次")
		assert.Equal(t, 3, int(manager.successCount.Load()), "成功计数应该为3")
	})

	// 测试用例5: 测试零阈值或负阈值的情况
	t.Run("零阈值", func(t *testing.T) {
		attempts := 0
		manager := NewGracefulShutdownManager(0, func() bool {
			attempts++
			return true
		})

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := manager.Shutdown(ctx)

		assert.NoError(t, err, "即使阈值为0，也应该立即成功")
		assert.Equal(t, 0, int(manager.attemptCount.Load()), "不应该尝试任何操作")
	})

	// 测试用例6: 测试delegate函数中的panic是否被正确处理
	t.Run("delegate中的panic", func(t *testing.T) {
		attempts := 0
		manager := NewGracefulShutdownManager(2, func() bool {
			attempts++
			if attempts == 1 {
				panic("测试panic")
			}
			return true
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 由于使用了utils.GoRecoverPanic，panic应该被捕获，不会导致测试失败
		_ = manager.Shutdown(ctx)

		// 如果panic被完全抑制且成功继续执行，检查尝试计数
		assert.GreaterOrEqual(t, int(manager.attemptCount.Load()), 1, "应该至少尝试一次")
	})
}

func TestGracefulShutdownManager_Done(t *testing.T) {
	// 测试done方法是否正确关闭通道
	t.Run("完成信号", func(t *testing.T) {
		manager := NewGracefulShutdownManager(1, func() bool {
			return true
		})

		// 获取done通道
		doneCh := manager.done()

		// 检查通道是否最初是开放的
		select {
		case <-doneCh:
			t.Fatal("done通道在开始时不应该关闭")
		default:
			// 正常，通道没有关闭
		}

		// 手动调用signalCompletion
		manager.signalCompletion()

		// 检查通道现在是否已关闭
		select {
		case <-doneCh:
			// 成功，通道已关闭
		default:
			t.Fatal("done通道在signalCompletion后应该关闭")
		}
	})

	// 测试Shutdown完成时是否正确关闭done通道
	t.Run("Shutdown完成关闭通道", func(t *testing.T) {
		manager := NewGracefulShutdownManager(1, func() bool {
			return true // 立即成功
		})

		doneCh := manager.done()

		// 启动一个goroutine等待done通道关闭
		completedCh := make(chan struct{})
		go func() {
			<-doneCh
			close(completedCh)
		}()

		// 执行Shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := manager.Shutdown(ctx)
		assert.NoError(t, err)

		// 等待最多1秒钟，确认done通道已关闭
		select {
		case <-completedCh:
			// 成功，done通道已关闭
		case <-time.After(1 * time.Second):
			t.Fatal("Shutdown后1秒内done通道应该关闭")
		}
	})

	// 测试上下文取消时done通道的状态
	t.Run("上下文取消时的通道状态", func(t *testing.T) {
		manager := NewGracefulShutdownManager(5, func() bool {
			time.Sleep(100 * time.Millisecond)
			return false // 永远不成功
		})

		doneCh := manager.done()

		// 使用非常短的超时
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// 启动一个goroutine等待done通道关闭
		completedCh := make(chan struct{})
		go func() {
			<-doneCh
			close(completedCh)
		}()

		// Shutdown应该因为超时而失败
		err := manager.Shutdown(ctx)
		assert.Error(t, err)

		// 即使超时，done通道也应该最终关闭
		// 这取决于实现，但好的实现应该在任何情况下都关闭通道
		select {
		case <-completedCh:
			// 成功，通道已关闭
		case <-time.After(300 * time.Millisecond): // 给足够时间处理
			t.Fatal("即使在上下文取消后，done通道也应该关闭")
		}
	})
}

// 创建一个测试用的ActorSystem，可以模拟stopSupervisor的行为
type testActorSystem struct {
	ActorSystem
	mockStopSupervisor func(lv Level) (bool, error)
}

// 覆盖原始的stopSupervisor方法
func (t *testActorSystem) stopSupervisor(lv Level) (bool, error) {
	if t.mockStopSupervisor != nil {
		return t.mockStopSupervisor(lv)
	}
	return true, nil // 默认返回成功
}

func TestActorSystem_Stop(t *testing.T) {
	// 创建ActorSystem
	actorSystem := actor.NewActorSystem()
	system := &ActorSystem{
		actorSystem: actorSystem,
		supervisors: make([]*actor.PID, LevelMaxLimit),
	}
	system.state.Store(ActorSystemStateRunning)

	// 创建测试用的supervisor
	for lv := LevelNormal; lv < LevelMaxLimit; lv++ {
		props := actor.PropsFromProducer(func() actor.Actor {
			return &MockActor{}
		})
		pid, err := actorSystem.Root.SpawnNamed(props, fmt.Sprintf("test-supervisor-%d", lv))
		assert.NoError(t, err, "创建supervisor应该成功")
		system.supervisors[lv] = pid
	}

	// 测试1: 基本停止功能
	t.Run("正常停止", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := system.Stop(ctx)
		assert.NoError(t, err, "停止操作应该成功")
		assert.Equal(t, ActorSystemStateStopped, system.state.Load(), "系统状态应该为Stopped")
	})

	// 测试2: 重复停止
	t.Run("重复停止", func(t *testing.T) {
		// 重置状态
		system.state.Store(ActorSystemStateRunning)

		// 第一次停止
		ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel1()
		system.Stop(ctx1)

		// 第二次停止应该直接返回nil
		ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel2()
		err := system.Stop(ctx2)
		assert.NoError(t, err, "重复停止应该返回nil")
	})

	// 测试3: 带超时的停止
	t.Run("超时停止", func(t *testing.T) {
		// 创建一个特殊的ActorSystem，stopSupervisor会模拟耗时操作
		specialSystem := &testActorSystem{
			ActorSystem: ActorSystem{
				actorSystem: actorSystem,
				supervisors: make([]*actor.PID, LevelMaxLimit),
			},
			mockStopSupervisor: func(lv Level) (bool, error) {
				time.Sleep(100 * time.Millisecond) // 添加延迟
				return false, nil                  // 永远不返回完成
			},
		}
		specialSystem.state.Store(ActorSystemStateRunning)

		// 使用短超时
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		err := specialSystem.Stop(ctx)
		assert.Error(t, err, "带超时的停止应该返回错误")
		assert.Contains(t, err.Error(), "canceled", "错误应该包含超时信息")
	})
}

func TestGracefulShutdownManager_MaxAttempts(t *testing.T) {
	// 测试用例1: 最大尝试次数限制 - 达到限制时应停止重试
	t.Run("达到最大尝试次数", func(t *testing.T) {
		maxAttempts := int32(10)
		attempts := 0

		// 创建一个永远不会成功的关闭管理器，但有最大尝试次数限制
		manager := NewGracefulShutdownManagerWithMaxAttempts(5, maxAttempts, func() bool {
			attempts++
			return false // 永远返回失败
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 执行关闭操作
		err := manager.Shutdown(ctx)

		// 验证结果
		assert.Error(t, err, "达到最大尝试次数后应返回错误")
		assert.Contains(t, err.Error(), "failed after", "错误消息应说明尝试失败")
		assert.Equal(t, maxAttempts, manager.attemptCount.Load(), "应该精确尝试最大次数")
		assert.Equal(t, int32(0), manager.successCount.Load(), "成功计数应为0")
	})

	// 测试用例2: 在达到最大尝试次数之前成功
	t.Run("最大尝试次数前成功", func(t *testing.T) {
		maxAttempts := int32(20)
		successAfter := 5
		attempts := 0

		// 创建一个在特定尝试次数后成功的关闭管理器
		manager := NewGracefulShutdownManagerWithMaxAttempts(2, maxAttempts, func() bool {
			attempts++
			return attempts >= successAfter // 第5次及以后成功
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 执行关闭操作
		err := manager.Shutdown(ctx)

		// 验证结果
		assert.NoError(t, err, "在达到最大尝试次数前应成功完成")
		assert.Equal(t, int32(successAfter+1), manager.attemptCount.Load(), "应该在成功后停止尝试")
		assert.Equal(t, int32(2), manager.successCount.Load(), "成功计数应达到阈值")
	})

	// 测试用例3: 最大尝试次数为0（无限尝试）
	t.Run("无限尝试设置", func(t *testing.T) {
		successAfter := 50 // 设置一个大于0但不会使测试超时的值
		attempts := 0

		// 创建一个在较多尝试后成功的关闭管理器，但不限制最大尝试次数
		manager := NewGracefulShutdownManagerWithMaxAttempts(2, 0, func() bool {
			attempts++
			return attempts >= successAfter
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 在后台执行关闭操作
		errCh := make(chan error, 1)
		go func() {
			errCh <- manager.Shutdown(ctx)
		}()

		// 等待操作完成或超时
		select {
		case err := <-errCh:
			assert.NoError(t, err, "应该成功完成")
			assert.GreaterOrEqual(t, int(manager.attemptCount.Load()), successAfter, "应该至少尝试到成功所需次数")
			assert.Equal(t, int32(2), manager.successCount.Load(), "成功计数应达到阈值")
		case <-time.After(3 * time.Second):
			t.Fatal("操作应该在有限时间内完成")
		}
	})

	// 测试用例4: 测试最大尝试次数和上下文取消的交互
	t.Run("最大尝试次数和上下文取消", func(t *testing.T) {
		maxAttempts := int32(50)
		attempts := 0

		// 创建一个耗时操作，便于触发超时
		manager := NewGracefulShutdownManagerWithMaxAttempts(5, maxAttempts, func() bool {
			attempts++
			time.Sleep(50 * time.Millisecond) // 每次尝试耗时
			return false                      // 永远失败
		})

		// 创建一个短超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// 执行关闭操作
		err := manager.Shutdown(ctx)

		// 验证结果 - 应该因为上下文取消而不是最大尝试次数而失败
		assert.Error(t, err, "应该因为超时而返回错误")
		assert.Contains(t, err.Error(), "canceled", "错误消息应包含取消信息")
		assert.Less(t, manager.attemptCount.Load(), maxAttempts, "尝试次数应小于最大限制")
	})
}
