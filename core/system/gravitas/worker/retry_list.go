package worker

// RetryList 重试列表：FIFO 顺序的 AnchorID 队列，带去重（文档 Section 10）。
//
// Worker 内部数据结构，单线程访问（Actor 单线程模型保证），无需同步。
// 使用 方案 A（简单列表），游戏服务器场景下 RetryList 通常较短。
//
// 一致性不变式 (INV-1)：
//
//	StagingArea[A] 非空 ⟺ A ∈ RetryList ∨ LoadingAnchors[A] = true
type RetryList struct {
	items []int64
	set   map[int64]struct{}
}

func newRetryList() *RetryList {
	return &RetryList{
		items: make([]int64, 0, 16),
		set:   make(map[int64]struct{}),
	}
}

// Add 添加 AnchorID 到重试列表（去重，已存在则跳过）。
func (r *RetryList) Add(anchorID int64) {
	if _, ok := r.set[anchorID]; ok {
		return
	}
	r.set[anchorID] = struct{}{}
	r.items = append(r.items, anchorID)
}

// Remove 从重试列表移除 AnchorID（幂等，不存在则跳过）。
func (r *RetryList) Remove(anchorID int64) {
	if _, ok := r.set[anchorID]; !ok {
		return
	}
	delete(r.set, anchorID)
	for i, id := range r.items {
		if id == anchorID {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return
		}
	}
}

// Contains 检查 AnchorID 是否在重试列表中。
func (r *RetryList) Contains(anchorID int64) bool {
	_, ok := r.set[anchorID]
	return ok
}

// Snapshot 返回当前列表的快照副本。
// RetryProcess 遍历期间可能触发 Add/Remove，使用快照避免遍历中修改问题。
func (r *RetryList) Snapshot() []int64 {
	snapshot := make([]int64, len(r.items))
	copy(snapshot, r.items)
	return snapshot
}

// Len 返回列表长度（监控用）。
func (r *RetryList) Len() int {
	return len(r.items)
}
