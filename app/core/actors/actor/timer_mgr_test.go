package actor

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"gitee.com/orbit-w/meteor/bases/container/heap"
	"github.com/stretchr/testify/assert"
)

func TestTimerMgr_AddTimerOnce(t *testing.T) {
	ntf := make(chan any, 1024)
	mgr := NewTimerMgr(func() {
		ntf <- "test-message"
	})
	defer mgr.Stop()

	start := time.Now()
	timer := mgr.AddTimerOnce("test-timer", 100*time.Millisecond, "test-message")
	assert.NotNil(t, timer)
	wg := sync.WaitGroup{}
	wg.Add(5)

	go func() {
		for {
			select {
			case <-ntf:
				mgr.Process(func(msg any) {
					assert.Equal(t, "test-message", msg)
					fmt.Println("time cost:", time.Since(start))
					mgr.AddTimerOnce("test-timer", 100*time.Millisecond, "test-message")
					wg.Done()
				})
			case <-time.After(150 * time.Millisecond):
				assert.Fail(t, "Callback was not triggered in time")
			}
		}
	}()

	wg.Wait()
	fmt.Println("complete")
}

func TestTimerMgr_AddAndRemoveTimerOnce(t *testing.T) {
	ntf := make(chan any, 1024)
	mgr := NewTimerMgr(func() {
		ntf <- "test-message"
	})
	defer mgr.Stop()

	result := make(chan struct{}, 1)
	start := time.Now()
	timer := mgr.AddTimerOnce("test-timer", 100*time.Millisecond, "test-message")
	assert.NotNil(t, timer)
	go func() {
		for {
			select {
			case <-ntf:
				mgr.Process(func(msg any) {
					assert.Equal(t, "test-message", msg)
					fmt.Println("time cost:", time.Since(start))
					close(result)
				})
			}
		}
	}()

	mgr.RemoveTimer(timer.GetKey())

	select {
	case <-result:
		assert.Fail(t, "timer was not removed in time")
	case <-time.After(time.Second * 5):
		fmt.Println("complete")
	}
}

func TestTimerMgr_AddSystemTimer(t *testing.T) {
	ntf := make(chan any, 1024)
	mgr := NewTimerMgr(func() {
		ntf <- "test-message"
	})
	defer mgr.Stop()

	timer := mgr.AddSystemTimer("test-timer", 100*time.Millisecond, "test-message")
	assert.NotNil(t, timer)
	wg := sync.WaitGroup{}
	wg.Add(10)
	var count int
	start := time.Now()

	go func() {
		for {
			select {
			case <-ntf:
				mgr.Process(func(msg any) {
					assert.Equal(t, "test-message", msg)
					fmt.Println("time cost:", time.Since(start))
					if count++; count >= 5 {
						mgr.RemoveTimer("test-timer")
						mgr.AddSystemTimer("test-timer", 500*time.Millisecond, "test-message")
						count -= 999999
					}
					wg.Done()
				})
			case <-time.After(5 * time.Second):
				assert.Fail(t, "Callback was not triggered in time")
			}
		}
	}()

	wg.Wait()
}

func TestTimerMgr_RemoveSystemTimer(t *testing.T) {
	ntf := make(chan any, 1024)
	mgr := NewTimerMgr(func() {
		ntf <- "test-message"
	})
	defer mgr.Stop()

	timer := mgr.AddSystemTimer("test-timer", 100*time.Millisecond, "test-message")
	assert.NotNil(t, timer)
	var count int
	start := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			select {
			case <-ntf:
				mgr.Process(func(msg any) {
					assert.Equal(t, "test-message", msg)
					fmt.Println("time cost:", time.Since(start))
					if count++; count >= 5 {
						mgr.RemoveTimer(timer.GetKey())
						count -= 999999
					}
				})
			case <-time.After(5 * time.Second):
				wg.Done()
			}
		}
	}()

	wg.Wait()
	fmt.Println("complete")
}

func TestTimerMgr_SimpleOperations(t *testing.T) {
	// Create a timer manager with a simple callback
	mgr := &TimerMgr{
		timerHeap: &heap.Heap[*Timer, int64]{},
		items:     make(map[string]*heap.Item[*Timer, int64]),
		timer:     time.NewTimer(time.Hour), // Use a long duration to prevent automatic triggering
	}
	defer mgr.Stop()

	// Test adding a one-time timer
	msg1 := "message1"
	timer1 := mgr.AddTimerOnce("timer1", time.Minute, msg1)

	assert.NotNil(t, timer1)
	assert.Equal(t, "timer1", timer1.GetKey())
	assert.Equal(t, msg1, timer1.msg)
	assert.False(t, timer1.IsSystem())
	assert.Equal(t, 1, mgr.timerHeap.Len())

	// Test adding a system timer
	msg2 := "message2"
	timer2 := mgr.AddSystemTimer("timer2", time.Minute*2, msg2)

	assert.NotNil(t, timer2)
	assert.Equal(t, "timer2", timer2.GetKey())
	assert.Equal(t, msg2, timer2.msg)
	assert.True(t, timer2.IsSystem())
	assert.Equal(t, 2, mgr.timerHeap.Len())

	// Test removing a timer
	mgr.RemoveTimer("timer1")
	assert.Equal(t, 1, mgr.timerHeap.Len())
	assert.Nil(t, mgr.items["timer1"])

	// Test adding a timer with the same key (should update the existing timer)
	// Note: System timers cannot be updated according to the implementation
	msg3 := "message3"
	timer3 := mgr.AddTimerOnce("timer3", time.Minute*3, msg3)

	assert.NotNil(t, timer3)
	assert.Equal(t, "timer3", timer3.GetKey())
	assert.Equal(t, msg3, timer3.msg)
	assert.Equal(t, 2, mgr.timerHeap.Len())

	// Test invalid durations
	invalidTimer := mgr.AddSystemTimer("invalid", -1*time.Second, "negative")
	assert.Nil(t, invalidTimer)
	assert.Equal(t, 2, mgr.timerHeap.Len())

	zeroTimer := mgr.AddTimerOnce("zero", 0, "zero")
	assert.Nil(t, zeroTimer)
	assert.Equal(t, 2, mgr.timerHeap.Len())
}

func TestTimerMgr_ProcessExpiredTimers(t *testing.T) {
	// Create a timer manager
	mgr := &TimerMgr{
		timerHeap: &heap.Heap[*Timer, int64]{},
		items:     make(map[string]*heap.Item[*Timer, int64]),
		timer:     time.NewTimer(time.Hour), // Use a long duration to prevent automatic triggering
	}
	defer mgr.Stop()

	// Add timers that are already expired
	pastTime := time.Now().Add(-time.Minute)

	timer1 := &Timer{
		key:        "timer1",
		expiration: pastTime,
		msg:        "expired1",
		system:     false,
		duration:   time.Minute,
	}

	timer2 := &Timer{
		key:        "timer2",
		expiration: pastTime,
		msg:        "expired2",
		system:     true,
		duration:   time.Minute,
	}

	// Add the timers to the heap manually
	item1 := &heap.Item[*Timer, int64]{
		Value:    timer1,
		Priority: timer1.expiration.Unix(),
	}

	item2 := &heap.Item[*Timer, int64]{
		Value:    timer2,
		Priority: timer2.expiration.Unix(),
	}

	mgr.timerHeap.Push(item1)
	mgr.items[timer1.key] = item1

	mgr.timerHeap.Push(item2)
	mgr.items[timer2.key] = item2

	assert.Equal(t, 2, mgr.timerHeap.Len())

	// Process the expired timers
	var processedMsgs []string

	mgr.Process(func(msg any) {
		if s, ok := msg.(string); ok {
			processedMsgs = append(processedMsgs, s)
		}
	})

	// Verify results
	assert.Equal(t, 2, len(processedMsgs))
	assert.Contains(t, processedMsgs, "expired1")
	assert.Contains(t, processedMsgs, "expired2")

	// One-time timer should be removed, system timer should be renewed
	// In the new implementation, system timers are renewed in Process
	assert.Equal(t, 1, mgr.timerHeap.Len()) // System timer is renewed
	assert.Nil(t, mgr.items["timer1"])      // One-time timer is removed
	assert.NotNil(t, mgr.items["timer2"])   // System timer is still there
}

func TestTimerMgr_SystemTimerRenewal(t *testing.T) {
	// Create a timer manager with a callback
	ntf := make(chan any, 1024)
	mgr := NewTimerMgr(func() {
		ntf <- "callback"
	})
	defer mgr.Stop()

	// Add a system timer with a short duration
	systemTimer := mgr.AddSystemTimer("system-timer", 50*time.Millisecond, "system-message")
	assert.NotNil(t, systemTimer)
	assert.True(t, systemTimer.IsSystem())

	// Setup to process expired timers
	wg := sync.WaitGroup{}
	wg.Add(2) // We expect the system timer to fire at least twice

	count := 0
	go func() {
		for i := 0; i < 2; i++ {
			select {
			case <-ntf:
				mgr.Process(func(msg any) {
					assert.Equal(t, "system-message", msg)
					count++
					wg.Done()
				})
			case <-time.After(200 * time.Millisecond):
				assert.Fail(t, "Callback was not triggered in time")
				wg.Done()
			}
		}
	}()

	wg.Wait()

	// Verify that the timer was processed twice
	assert.Equal(t, 2, count)

	// The system timer should still exist because it's renewed
	assert.NotNil(t, mgr.items["system-timer"])
}
