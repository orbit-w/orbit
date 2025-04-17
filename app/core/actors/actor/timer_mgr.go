package actor

import (
	"sync/atomic"
	"time"

	"gitee.com/orbit-w/meteor/bases/container/heap"
)

// Timer 表示一个定时器
type Timer struct {
	uuid       string
	repeat     bool
	duration   time.Duration
	expiration time.Time // 到期时间
	msg        any
}

func (t *Timer) GetUuid() string {
	return t.uuid
}

func (t *Timer) Equal(other *Timer) bool {
	return t.uuid == other.uuid
}

func (t *Timer) expired() bool {
	return !time.Now().Before(t.expiration)
}

// TimerMgr 使用最小堆管理Timer对象
type TimerMgr struct {
	id        atomic.Int64
	items     map[string]*heap.Item[*Timer, int64]
	timerHeap *heap.Heap[*Timer, int64] // 使用最小堆存储Timer
	timer     *time.Timer
}

// NewTimerMgr 创建一个新的TimerMgr
func NewTimerMgr(callback func()) *TimerMgr {
	mgr := &TimerMgr{
		timerHeap: &heap.Heap[*Timer, int64]{},
		timer:     time.AfterFunc(0, callback),
	}

	mgr.id.Store(0)
	return mgr
}

// AddTimer 添加一个新的定时器
func (t *TimerMgr) addTimer(key string, duration time.Duration, msg any, repeat bool) *Timer {
	timer := &Timer{
		uuid:       key,
		expiration: time.Now().Add(duration),
		msg:        msg,
		repeat:     repeat,
		duration:   duration,
	}

	var (
		item   *heap.Item[*Timer, int64]
		inited bool
	)

	if item = t.items[key]; item != nil {
		item.Value = timer
		item.Priority = timer.expiration.Unix()
		t.update(item)
	} else {
		item = &heap.Item[*Timer, int64]{
			Value:    timer,
			Priority: timer.expiration.Unix(),
		}
		t.push(item)
		inited = true
	}

	head := t.peek()
	if inited && !head.Equal(timer) {
		return timer
	}

	t.resetTimer(duration)

	return timer
}

func (t *TimerMgr) AddTimerRepeat(key string, duration time.Duration, msg any) *Timer {
	if duration <= 0 {
		return nil
	}
	return t.addTimer(key, duration, msg, true)
}

func (t *TimerMgr) AddTimerOnce(key string, duration time.Duration, msg any) *Timer {
	return t.addTimer(key, duration, msg, false)
}

// RemoveTimer 从管理器中移除一个定时器
func (t *TimerMgr) RemoveTimer(key string) {
	if item := t.items[key]; item != nil {
		head := t.peek()
		if head.Equal(item.Value) {
			t.pop()
			head = t.peek()
			t.resetTimer(time.Until(head.expiration))
		} else {
			t.remove(item)
		}
	}
}

// Stop 停止定时器管理器
func (t *TimerMgr) Stop() {
	t.stopSystemTimer()
}

func (t *TimerMgr) push(item *heap.Item[*Timer, int64]) {
	t.timerHeap.Push(item)
	t.items[item.Value.uuid] = item
}

func (t *TimerMgr) update(item *heap.Item[*Timer, int64]) {
	t.timerHeap.Fix(item.Index)
}

func (t *TimerMgr) peek() *Timer {
	return t.timerHeap.Peek().Value
}

func (t *TimerMgr) pop() *Timer {
	item := t.timerHeap.Pop()
	delete(t.items, item.Value.uuid)
	return item.Value
}

func (t *TimerMgr) remove(item *heap.Item[*Timer, int64]) {
	t.timerHeap.Delete(item.Index)
	delete(t.items, item.Value.uuid)
}

func (t *TimerMgr) resetTimer(duration time.Duration) {
	if t.timer == nil {
		t.timer = time.NewTimer(duration)
	}

	t.stopSystemTimer()
	t.timer.Reset(duration)
}

func (t *TimerMgr) stopSystemTimer() {
	if !t.timer.Stop() {
		<-t.timer.C
	}
}

func (t *TimerMgr) Process(delegate func(msg any)) {
	var (
		repeated []*Timer
		process  bool
	)

	for {
		item := t.timerHeap.Peek()
		if item == nil {
			break
		}

		timer := item.Value

		if !timer.expired() {
			break
		}
		process = true
		delegate(timer.msg)

		if timer.repeat {
			timer.expiration = timer.expiration.Add(timer.duration)
			repeated = append(repeated, timer)
		}
		t.pop()
	}

	for i := range repeated {
		timer := repeated[i]
		t.addTimer(timer.uuid, timer.duration, timer.msg, timer.repeat)
	}

	if process {
		if head := t.peek(); head != nil {
			t.resetTimer(time.Until(head.expiration))
		}
	}
}
