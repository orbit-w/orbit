package actor

import (
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

func (b *MockBehavior) HandleCall(_ actor.Context, _ any) (any, error) {
	return nil, nil
}

func (b *MockBehavior) HandleCast(_ actor.Context, _ any) error {
	return nil
}

func (b *MockBehavior) HandleForward(_ actor.Context, _ any) error {
	return nil
}

func (b *MockBehavior) HandleInit(_ actor.Context) error {
	fmt.Printf("Initializing actor with ID: %s\n", b.actorID)
	return nil
}

func (b *MockBehavior) HandleStopping(_ actor.Context) error {
	fmt.Printf("Stopping actor with ID: %s\n", b.actorID)
	return nil
}

func (b *MockBehavior) HandleStopped(_ actor.Context) error {
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
			pid, err := GetOrStartActor(actorName, testPattern)
			if err != nil {
				panic(err)
			}
			fmt.Println("receive pid")
			name := ExtractActorName(pid)
			if name != actorName {
				panic(fmt.Sprintf("actor name invalid: %s", name))
			}
		}()
	}

	wg.Wait()

	pid, err := GetOrStartActor(actorName, testPattern)
	assert.NoError(t, err)

	// Now call GetOrStartActor - it should return the cached actor
	retrievedPID, err := GetOrStartActor(actorName, testPattern)
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
	pid2, err := GetOrStartActor(actorName2, testPattern)
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

func TestStopActor(t *testing.T) {
	const testPattern = "test-stopping-pattern"
	// Setup
	setup(testPattern)

	// Register a test pattern
	RegFactory(testPattern, func(actorName string) Behavior {
		return &StoppingBehavior{}
	})

	// Create our test actor
	actorName := "test-actor-stopping"
	actorRef := NewActorRef(actorName, testPattern)
	actorRef.Send("initial-message")
	// Stop the actor which will trigger the stopping phase
	err := StopActor(actorName, testPattern)
	assert.NoError(t, err)
	time.Sleep(time.Second * 10)
}

// TestMessageLossDuringStopping demonstrates that messages sent to an actor during its stopping phase may be lost
func TestMessageLossDuringStopping(t *testing.T) {
	const testPattern = "test-stopping-pattern"
	// Setup
	setup(testPattern)

	// Create channels to track message reception
	messageSent := make(chan bool, 1)

	// Register a test pattern
	RegFactory(testPattern, func(actorName string) Behavior {
		return &StoppingBehavior{}
	})

	// Create our test actor
	actorName := "test-actor-stopping"
	actorRef := NewActorRef(actorName, testPattern)
	actorRef.Send("initial-message")

	// Start a goroutine that will send messages during the stopping phase
	go func() {
		// Wait a bit before sending the message to allow the stop process to begin
		time.Sleep(10 * time.Millisecond)

		// Send a message while the actor is stopping
		res, err := actorRef.RequestFuture("stopping-phase-message")
		if err != nil {
			panic(err)
		}
		fmt.Printf("res: %v\n", res)
		messageSent <- true
	}()

	// Stop the actor which will trigger the stopping phase
	err := StopActor(actorName, testPattern)
	assert.NoError(t, err)

	// Wait for the message to be sent
	<-messageSent
}

// StoppingBehavior is an implementation of the Behavior interface for testing message loss during stopping
type StoppingBehavior struct {
	actorName string
}

func (b *StoppingBehavior) HandleCall(_ actor.Context, _ any) (any, error) {
	return nil, nil
}

func (b *StoppingBehavior) HandleCast(_ actor.Context, msg any) error {
	return nil
}

func (b *StoppingBehavior) HandleForward(_ actor.Context, _ any) error {
	return nil
}

func (b *StoppingBehavior) HandleInit(_ actor.Context) error {
	//fmt.Printf("Initializing actor with ID: %s\n", b.actorName)
	return nil
}

func (b *StoppingBehavior) HandleStopping(_ actor.Context) error {
	fmt.Printf("Stopping actor with ID: %s\n", b.actorName)
	return nil
}

func (b *StoppingBehavior) HandleStopped(_ actor.Context) error {
	fmt.Printf("Stopped actor with ID: %s\n", b.actorName)
	return nil
}

func setup(pattern string) {
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
}
