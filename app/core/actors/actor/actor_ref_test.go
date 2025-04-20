package actor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 校验停止Actor是否正确
func Test_StopActor(t *testing.T) {
	const testPattern = "test-stopping-pattern"
	// Setup
	service := setup(testPattern)

	// Register a test pattern
	RegFactory(testPattern, func(actorName string) Behavior {
		return &StoppingBehavior{
			actorName: actorName,
		}
	})

	// Create our test actor
	actorName := "test-actor-stopping"
	actorRef := NewActorRef(NewProps(), actorName, testPattern)
	err := actorRef.Send("initial-message")
	assert.NoError(t, err)
	// Stop the actor which will trigger the stopping phase
	err = StopActor(actorName, testPattern)
	assert.NoError(t, err)
	time.Sleep(time.Second * 10)
	service.Stop(context.Background())
}

// 校验停止Actor时，消息是否丢失
// 场景：
//  1. 由于向正在停止过程中的 Actor 发送消息而导致消息丢失
//  2. 由于向已经停止的 Actor 发送消息而导致消息丢失
func Test_MessageLossDuringStopping(t *testing.T) {
	const testPattern = "test-stopping-pattern"
	// Setup
	service := setup(testPattern)

	// Create channels to track message reception
	messageSent := make(chan bool, 1)

	// Register a test pattern
	RegFactory(testPattern, func(actorName string) Behavior {
		return &StoppingBehavior{
			actorName: actorName,
		}
	})

	// Create our test actor
	actorName := "test-actor-stopping"
	actorRef := NewActorRef(NewProps(), actorName, testPattern)
	actorRef.Send("initial-message")

	// Start a goroutine that will send messages during the stopping phase
	go func() {
		// Wait a bit before sending the message to allow the stop process to begin
		time.Sleep(10 * time.Millisecond)

		// Send a message while the actor is stopping
		res, err := actorRef.RequestFuture("stopping-phase-message")
		assert.Error(t, err)
		assert.Nil(t, res)
		messageSent <- true
	}()

	// Stop the actor which will trigger the stopping phase
	err := StopActor(actorName, testPattern)
	assert.NoError(t, err)

	// Wait for the message to be sent
	<-messageSent

	service.Stop(context.Background())
}

// 校验Props传递的参数是否正确
func Test_ActorRefPropsContent(t *testing.T) {
	pattern := "content-pattern"
	service := setup(pattern)

	// Register a test pattern
	RegFactory(pattern, func(actorName string) Behavior {
		return &ContentBehavior{
			actorName: actorName,
		}
	})

	name := "test-actor"
	meta := NewMeta(name, pattern, "1", nil)
	actorRef := NewActorRef(NewProps(), name, pattern, WithMeta(meta))
	err := actorRef.Send("initial-message")
	assert.NoError(t, err)
	err = actorRef.Send("second-message")
	assert.NoError(t, err)
	actorRef.Stop()
	time.Sleep(time.Second * 10)
	_ = service.Stop(context.Background())
}

// 校验ActorRef的Start和Stop方法是否正确
func Test_ActorRefStartAndStop(t *testing.T) {
	// 测试配置
	const (
		pattern       = "start-and-stop-pattern"
		numGoroutines = 1000
		msgPerRoutine = 10
		totalMsg      = numGoroutines * msgPerRoutine
	)

	// 设置测试环境
	service := setup(pattern)
	var count atomic.Int32

	// 注册测试行为
	RegFactory(pattern, func(actorName string) Behavior {
		return &CountBehavior{
			actorName: actorName,
			count:     &count,
		}
	})

	// 创建测试Actor
	name := "test-actor"
	meta := NewMeta(name, pattern, "1", nil)
	actorRef := NewActorRef(NewProps(), name, pattern, WithMeta(meta))

	// 准备测试数据
	wg := sync.WaitGroup{}
	var msgCount atomic.Int32
	var errCount atomic.Int32

	// 模拟并发向一个正在停止中的Actor发送消息
	// 验证是否所有消息都被正确处理或记录错误
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < msgPerRoutine; j++ {
				if err := actorRef.Send(fmt.Sprintf("message-%d", msgCount.Add(1))); err != nil {
					errCount.Add(1)
				}
			}
		}()
	}

	// 模拟在消息发送过程中停止Actor
	wg.Add(1)
	go func() {
		defer wg.Done()
		actorRef.Stop()
	}()

	// 等待所有操作完成
	wg.Wait()
	time.Sleep(time.Second * 2) // 等待消息处理完成

	// 结果验证
	successes := count.Load()
	failures := errCount.Load()
	total := successes + failures

	fmt.Printf("成功处理: %d, 发送失败: %d, 总计: %d/%d\n",
		successes, failures, total, totalMsg)

	// 验证所有消息都被正确处理或记录为错误
	assert.Equal(t, int32(totalMsg), total,
		"所有消息应该被处理或记录为错误")

	_ = service.Stop(context.Background())
}

func Test_ActorTimerForFree(t *testing.T) {
	// 测试配置
	const (
		pattern = "timer_for_free"
	)

	// 设置测试环境
	service := setup(pattern)
	defer service.Stop(context.Background())
	ntf := make(chan struct{}, 1)

	// 注册测试行为
	RegFactory(pattern, func(actorName string) Behavior {
		return &CheckAliveBehavior{
			ntf: ntf,
		}
	})

	start := time.Now()
	actorRef := NewActorRef(NewProps(), pattern, pattern, WithAliveTimeout(time.Minute))
	actorRef.Send("initial-message")
	<-ntf

	fmt.Printf("test_ActorTimerForFree done, cost: %s\n", time.Since(start))
}

type ContentBehavior struct {
	actorName string
}

func (b *ContentBehavior) HandleRequest(ctx IContext, msg any) (any, error) {
	v, ok := msg.(string)
	if !ok {
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
	fmt.Printf("HandleCall message: %s\n", v)
	return v, nil
}

func (b *ContentBehavior) HandleSend(ctx IContext, msg any) {
	v, ok := msg.(string)
	if !ok {
		return
	}
	fmt.Printf("HandleCast message: %s\n", v)
	return
}

func (b *ContentBehavior) HandleForward(ctx IContext, _ any) {
}

func (b *ContentBehavior) HandleInit(ctx IContext) error {
	fmt.Printf("Initializing actor with ID: %s, serverId: %s\n", b.actorName, ctx.GetServerId())
	return nil
}

func (b *ContentBehavior) HandleStopping(ctx IContext) error {
	fmt.Printf("Stopping actor with ID: %s, serverId: %s\n", b.actorName, ctx.GetServerId())
	return nil
}

func (b *ContentBehavior) HandleStopped(ctx IContext) error {
	fmt.Printf("Stopped actor with ID: %s, serverId: %s\n", b.actorName, ctx.GetServerId())
	return nil
}

type CountBehavior struct {
	actorName string
	count     *atomic.Int32
}

func (b *CountBehavior) HandleRequest(ctx IContext, msg any) (any, error) {
	v := msg.(string)
	//fmt.Printf("HandleCall message: %s\n", v)
	b.count.Add(1)
	return v, nil
}

func (b *CountBehavior) HandleSend(ctx IContext, msg any) {
	// v := msg.(string)
	// fmt.Printf("HandleCast message: %s\n", v)
	b.count.Add(1)
}

func (b *CountBehavior) HandleForward(ctx IContext, _ any) {
}

func (b *CountBehavior) HandleInit(ctx IContext) error {
	return nil
}

func (b *CountBehavior) HandleStopping(ctx IContext) error {
	return nil
}

func (b *CountBehavior) HandleStopped(ctx IContext) error {
	fmt.Printf("Stopped actor with ID: %s, serverId: %s, count: %d\n", b.actorName, ctx.GetServerId(), b.count.Load())
	return nil
}

type CheckAliveBehavior struct {
	ntf chan struct{}
}

func (b *CheckAliveBehavior) HandleRequest(ctx IContext, msg any) (any, error) {
	//fmt.Printf("HandleCall message: %s\n", v)
	return nil, nil
}

func (b *CheckAliveBehavior) HandleSend(ctx IContext, msg any) {
	// v := msg.(string)
	// fmt.Printf("HandleCast message: %s\n", v)
}

func (b *CheckAliveBehavior) HandleForward(ctx IContext, _ any) {
}

func (b *CheckAliveBehavior) HandleInit(ctx IContext) error {
	return nil
}

func (b *CheckAliveBehavior) HandleStopping(ctx IContext) error {
	return nil
}

func (b *CheckAliveBehavior) HandleStopped(ctx IContext) error {
	close(b.ntf)
	return nil
}
