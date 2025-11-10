package actor

import (
	"sync/atomic"
	"time"

	"gitee.com/orbit-w/meteor/bases/container/heap"
)

// Timer 表示一个定时器
type Timer struct {
	key        string
	system     bool
	idx        int64
	duration   time.Duration
	expiration time.Time // 到期时间
	msg        any
}

func NewTimer(key string, idx int64, duration time.Duration, msg any) *Timer {
	return &Timer{
		key:        key,
		idx:        idx,
		duration:   duration,
		expiration: time.Now().Add(duration),
		msg:        msg,
	}
}

func NewSystemTimer(key string, idx int64, duration time.Duration, msg any) *Timer {
	return &Timer{
		key:        key,
		idx:        idx,
		duration:   duration,
		expiration: time.Now().Add(duration),
		msg:        msg,
		system:     true,
	}
}

func (t *Timer) GetKey() string {
	return t.key
}

func (t *Timer) Equal(other *Timer) bool {
	return t.key == other.key && t.idx == other.idx
}

func (t *Timer) expired() bool {
	return !time.Now().Before(t.expiration)
}

func (t *Timer) IsSystem() bool {
	return t.system
}

func (t *Timer) GetDuration() time.Duration {
	return t.duration
}

// TimerMgr 使用最小堆管理Timer对象
//
//	1:所有接口不允许并发操作
//	2:管理两种定时器
//		Timer: 普通定时器，不支持自动续约
//			支持操作类型：Insert，Remove，Update
//		SystemTimer: 系统定时器，支持自动续约
//			支持操作类型：Insert，Remove
//	3: 所有接口不允许并发操作
type TimerMgr struct {
	id        atomic.Int64
	timer     *time.Timer
	systemMap map[string]*Timer // 系统定时器映射
	items     map[string]*heap.Item[*Timer, int64]
	timerHeap *heap.Heap[*Timer, int64] // 使用最小堆存储Timer
}

// NewTimerMgr 创建一个新的TimerMgr
func NewTimerMgr(callback func()) *TimerMgr {
	mgr := &TimerMgr{
		timerHeap: &heap.Heap[*Timer, int64]{},
		timer:     time.AfterFunc(0, callback),
		systemMap: make(map[string]*Timer),
		items:     make(map[string]*heap.Item[*Timer, int64]),
	}

	mgr.id.Store(0)
	return mgr
}

// AddSystemTimer 添加一个系统定时器
// SystemTimer：
//
//	1: 系统定时器会自动续约
//	2: 系统定时器不支持更新操作，如果需要更新，请使用RemoveTimer和AddSystemTimer
func (t *TimerMgr) AddSystemTimer(key string, duration time.Duration, msg any) *Timer {
	if duration <= 0 {
		return nil
	}

	//如果系统定时器已经存在，不允许重复添加
	if _, exist := t.systemMap[key]; exist {
		return nil
	}

	timer := NewSystemTimer(key, t.id.Add(1), duration, msg)
	t.addTimer(timer)
	t.systemMap[key] = timer
	return timer
}

// AddTimerOnce 添加一个一次性定时器
// Timer:
//
//	1: 一次性定时器,支持插入/更新/删除操作
//	2: 一次性定时器不支持自动续约
func (t *TimerMgr) AddTimerOnce(key string, duration time.Duration, msg any) *Timer {
	if duration <= 0 {
		return nil
	}

	timer := NewTimer(key, t.id.Add(1), duration, msg)
	return t.addTimer(timer)
}

// RemoveTimer 从管理器中移除一个定时器
func (t *TimerMgr) RemoveTimer(key string) {
	delete(t.systemMap, key)
	if item := t.items[key]; item != nil {
		t.remove(item)
		t.schedule()
	}
}

// renewalTimer 续租一个系统定时器
func (t *TimerMgr) renewalSystemTimer(timer *Timer) {
	if !timer.IsSystem() {
		return
	}

	//如果定时器已经被删除，不会在续约
	if remain := t.systemMap[timer.key]; remain != nil {
		if timer.Equal(remain) {
			timer.expiration = time.Now().Add(timer.duration)
			t.initTimer(timer)
		}
	}
}

// AddTimer 添加一个新的定时器
func (t *TimerMgr) addTimer(timer *Timer) *Timer {
	if item := t.items[timer.key]; item != nil {
		t.updateTimer(item, timer)
	} else {
		t.initTimer(timer)
	}

	t.schedule()

	return timer
}

func (t *TimerMgr) initTimer(timer *Timer) {
	item := &heap.Item[*Timer, int64]{
		Value:    timer,
		Priority: timer.expiration.Unix(),
	}
	t.push(item)
}

func (t *TimerMgr) updateTimer(item *heap.Item[*Timer, int64], timer *Timer) {
	item.Value = timer
	item.Priority = timer.expiration.Unix()
	t.update(item)
}

func (t *TimerMgr) schedule() {
	head := t.peek()
	t.stopSystemTimer()
	if head != nil {
		t.timer.Reset(time.Until(head.expiration))
	}
}

// Stop 停止定时器管理器
func (t *TimerMgr) Stop() {
	t.stopSystemTimer()
}

func (t *TimerMgr) push(item *heap.Item[*Timer, int64]) {
	t.timerHeap.Push(item)
	t.items[item.Value.key] = item
}

func (t *TimerMgr) pop() *Timer {
	head := t.timerHeap.Pop()
	if head != nil {
		delete(t.items, head.Value.key)
		return head.Value
	}
	return nil
}

func (t *TimerMgr) update(item *heap.Item[*Timer, int64]) {
	t.timerHeap.Fix(item.Index)
}

func (t *TimerMgr) peek() *Timer {
	head := t.timerHeap.Peek()
	if head == nil {
		return nil
	}
	return head.Value
}

func (t *TimerMgr) remove(item *heap.Item[*Timer, int64]) {
	t.timerHeap.Delete(item.Index)
	timer := item.Value
	delete(t.items, timer.key)
}

func (t *TimerMgr) stopSystemTimer() {
	t.timer.Stop()
}

func (t *TimerMgr) Process(delegate func(msg any)) {
	var (
		repeated []*Timer
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

		t.pop()
		delegate(timer.msg)

		if timer.IsSystem() {
			repeated = append(repeated, timer)
		}
	}

	for i := range repeated {
		timer := repeated[i]
		t.renewalSystemTimer(timer)
	}

	t.schedule()
}
