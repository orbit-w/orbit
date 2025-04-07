package manager

import (
	"gitee.com/orbit-w/meteor/bases/container/priority_queue"
)

type MapSlice[K comparable, V any] struct {
	pq *priority_queue.PriorityQueue[K, []V, int64]
}

func NewMapSlice[K comparable, V any]() *MapSlice[K, V] {
	return &MapSlice[K, V]{
		pq: priority_queue.New[K, []V, int64](),
	}
}

func (m *MapSlice[K, V]) Push(key K, value V, priority int64) {
	ent, ok := m.pq.Get(key)
	if !ok {
		m.pq.Push(key, []V{value}, priority)
		return
	}
	values := ent.Value.GetValue()
	values = append(values, value)
	m.pq.UpdateValue(key, values)
}

func (m *MapSlice[K, V]) Pop(key K) ([]V, bool) {
	return m.pq.PopK(key)
}

func (m *MapSlice[K, V]) Exists(key K) bool {
	return m.pq.Exist(key)
}
