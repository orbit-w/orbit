package unipue_task_exec

import "context"

// TaskRunner 处理单个任务的执行和结果等待
type TaskRunner struct {
	done chan any
}

// NewTaskRunner 创建一个新的TaskRunner实例
func NewTaskRunner() *TaskRunner {
	return &TaskRunner{done: make(chan any, 1)}
}

// Execute 执行任务并存储其结果
func (t *TaskRunner) Execute(do func() any) {
	result := do()
	t.Done(result)
}

// Done 设置任务的完成结果
func (t *TaskRunner) Done(result any) {
	select {
	case t.done <- result:
	default:
	}
}

// Wait 等待任务完成并返回结果
// 如果ctx被取消，则返回ctx.Err()
func (t *TaskRunner) Wait(ctx context.Context) any {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case result := <-t.done:
		return result
	}
}
