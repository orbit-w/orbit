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

func Test_ActorRefStartAndStop(t *testing.T) {
	pattern := "start-and-stop-pattern"
	service := setup(pattern)
	var count atomic.Int32
	// Register a test pattern
	RegFactory(pattern, func(actorName string) Behavior {
		return &CountBehavior{
			actorName: actorName,
			count:     &count,
		}
	})

	name := "test-actor"
	meta := NewMeta(name, pattern, "1", nil)
	actorRef := NewActorRef(NewProps(), name, pattern, WithMeta(meta))
	wg := sync.WaitGroup{}
	var msgCount atomic.Int32
	var errCount atomic.Int32
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			if err := actorRef.Send(fmt.Sprintf("message-%d", msgCount.Add(1))); err != nil {
				errCount.Add(1)
			}
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		actorRef.Stop()
		wg.Done()
	}()

	wg.Wait()
	time.Sleep(time.Second * 5)
	fmt.Printf("errCount: %d\n", errCount.Load())
	assert.Equal(t, errCount.Load()+count.Load(), int32(1000))
	_ = service.Stop(context.Background())
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
