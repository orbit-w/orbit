package worker

import "time"

// StagingArea 暂存区：按 AnchorID 管理消息队列（文档 Section 9）。
//
// Worker 内部数据结构，单线程访问（Actor 单线程模型保证），无需同步。
//
// 生命周期：
//   - 创建：LockBusy/NeedIO 时由 PushHead 创建
//   - 追加：新消息到达且队列已存在时由 PushTail 追加（FIFO）
//   - 消费：RetryProcess/AsyncComplete 时由 PopHead 取出
//   - 销毁：PopHead 后队列为空时自动删除
type StagingArea struct {
	queues map[int64]*messageQueue
}

type messageQueue struct {
	messages []*WorkerMessage
	createAt time.Time
}

func newStagingArea() *StagingArea {
	return &StagingArea{
		queues: make(map[int64]*messageQueue),
	}
}

// PushHead 消息入队首。
// 用于 LockBusy/NeedIO 回滚：该消息"正在执行但被阻塞"，优先级高于后续到达的消息。
func (sa *StagingArea) PushHead(anchorID int64, msg *WorkerMessage) {
	queue := sa.getOrCreate(anchorID)
	queue.messages = append([]*WorkerMessage{msg}, queue.messages...)
}

// PushTail 消息入队尾，保持 FIFO 顺序。
// 用于新消息追加（已有堆积时）和错误重试入队。
func (sa *StagingArea) PushTail(anchorID int64, msg *WorkerMessage) {
	queue := sa.getOrCreate(anchorID)
	queue.messages = append(queue.messages, msg)
}

// PopHead 取出队首消息。队列变空时自动删除。
func (sa *StagingArea) PopHead(anchorID int64) (*WorkerMessage, bool) {
	queue := sa.queues[anchorID]
	if queue == nil || len(queue.messages) == 0 {
		return nil, false
	}
	msg := queue.messages[0]
	queue.messages = queue.messages[1:]
	if len(queue.messages) == 0 {
		delete(sa.queues, anchorID)
	}
	return msg, true
}

// Exists 检查指定 Anchor 的队列是否存在且非空。
func (sa *StagingArea) Exists(anchorID int64) bool {
	queue := sa.queues[anchorID]
	return queue != nil && len(queue.messages) > 0
}

// Len 返回暂存区中活跃的 Anchor 数量（监控用）。
func (sa *StagingArea) Len() int {
	return len(sa.queues)
}

func (sa *StagingArea) getOrCreate(anchorID int64) *messageQueue {
	queue := sa.queues[anchorID]
	if queue == nil {
		queue = &messageQueue{
			messages: make([]*WorkerMessage, 0, 4),
			createAt: time.Now(),
		}
		sa.queues[anchorID] = queue
	}
	return queue
}
