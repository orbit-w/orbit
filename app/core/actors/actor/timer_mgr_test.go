package actor

import (
	"testing"
	"time"

	"gitee.com/orbit-w/meteor/bases/container/heap"
	"github.com/stretchr/testify/assert"
)

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
	assert.Equal(t, "timer1", timer1.uuid)
	assert.Equal(t, msg1, timer1.msg)
	assert.False(t, timer1.repeat)
	assert.Equal(t, 1, mgr.timerHeap.Len())

	// Test adding a repeating timer
	msg2 := "message2"
	timer2 := mgr.AddTimerRepeat("timer2", time.Minute*2, msg2)

	assert.NotNil(t, timer2)
	assert.Equal(t, "timer2", timer2.uuid)
	assert.Equal(t, msg2, timer2.msg)
	assert.True(t, timer2.repeat)
	assert.Equal(t, 2, mgr.timerHeap.Len())

	// Test removing a timer
	mgr.RemoveTimer("timer1")
	assert.Equal(t, 1, mgr.timerHeap.Len())
	assert.Nil(t, mgr.items["timer1"])

	// Test adding a timer with the same ID (should update the existing timer)
	msg3 := "message3"
	timer3 := mgr.AddTimerOnce("timer2", time.Minute*3, msg3)

	assert.NotNil(t, timer3)
	assert.Equal(t, "timer2", timer3.uuid)
	assert.Equal(t, msg3, timer3.msg)
	assert.False(t, timer3.repeat) // Should have changed from repeat to once
	assert.Equal(t, 1, mgr.timerHeap.Len())

	// Test invalid durations
	invalidTimer := mgr.AddTimerRepeat("invalid", -1*time.Second, "negative")
	assert.Nil(t, invalidTimer)
	assert.Equal(t, 1, mgr.timerHeap.Len())

	zeroTimer := mgr.AddTimerRepeat("zero", 0, "zero")
	assert.Nil(t, zeroTimer)
	assert.Equal(t, 1, mgr.timerHeap.Len())
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
		uuid:       "timer1",
		expiration: pastTime,
		msg:        "expired1",
		repeat:     false,
		duration:   time.Minute,
	}

	timer2 := &Timer{
		uuid:       "timer2",
		expiration: pastTime,
		msg:        "expired2",
		repeat:     true,
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
	mgr.items[timer1.uuid] = item1

	mgr.timerHeap.Push(item2)
	mgr.items[timer2.uuid] = item2

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

	// One-time timer should be removed, repeat timer should be kept
	assert.Equal(t, 1, mgr.timerHeap.Len())
	assert.Nil(t, mgr.items["timer1"])
	assert.NotNil(t, mgr.items["timer2"])
}

func TestTimerMgr_PeekAndUpdate(t *testing.T) {
	// Create a timer manager
	mgr := &TimerMgr{
		timerHeap: &heap.Heap[*Timer, int64]{},
		items:     make(map[string]*heap.Item[*Timer, int64]),
		timer:     time.NewTimer(time.Hour),
	}
	defer mgr.Stop()

	// Test peek on empty heap
	item := mgr.timerHeap.Peek()
	assert.Nil(t, item, "Peek on empty heap should return nil")

	// Add multiple timers with different expiration times
	// Make sure they're added in non-chronological order to test the heap property
	timer3 := mgr.AddTimerOnce("timer3", time.Minute*3, "msg3")
	timer1 := mgr.AddTimerOnce("timer1", time.Minute*1, "msg1")
	timer2 := mgr.AddTimerOnce("timer2", time.Minute*2, "msg2")

	// Verify all timers were added
	assert.Equal(t, 3, mgr.timerHeap.Len())

	// The peek should return the timer with the earliest expiration (timer1)
	headTimer := mgr.peek()
	assert.Equal(t, timer1.uuid, headTimer.uuid)
	assert.Equal(t, "msg1", headTimer.msg)

	// Test updating a timer's expiration
	// Update timer1 to have a later expiration than timer3
	updatedTimer := mgr.AddTimerOnce("timer1", time.Minute*4, "updated-msg1")
	assert.Equal(t, timer1.uuid, updatedTimer.uuid)
	assert.Equal(t, "updated-msg1", updatedTimer.msg)

	// Now the peek should return timer2 (the new earliest timer)
	newHeadTimer := mgr.peek()
	assert.Equal(t, timer2.uuid, newHeadTimer.uuid)
	assert.Equal(t, "msg2", newHeadTimer.msg)

	// Pop the head timer
	poppedTimer := mgr.pop()
	assert.Equal(t, timer2.uuid, poppedTimer.uuid)
	assert.Equal(t, 2, mgr.timerHeap.Len())

	// Now timer3 should be the head
	nextHeadTimer := mgr.peek()
	assert.Equal(t, timer3.uuid, nextHeadTimer.uuid)
}
