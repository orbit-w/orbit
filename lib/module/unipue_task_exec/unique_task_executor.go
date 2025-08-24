package unipue_task_exec

import (
	"context"
	"fmt"
	"sync"
)

// UniqueTaskExecutor 确保相同key的任务只执行一次的并发执行器
type UniqueTaskExecutor struct {
	cache sync.Map
}

// NewUniqueTaskExecutor 创建一个新的UniqueTaskExecutor实例
func NewUniqueTaskExecutor() *UniqueTaskExecutor {
	return &UniqueTaskExecutor{
		cache: sync.Map{},
	}
}

// ExecuteOnce 执行一个任务，相同key的任务只会执行一次
// 无论有多少goroutine尝试执行相同key的任务，只有第一次会真正执行
// 其余调用会等待第一次执行的结果
func (c *UniqueTaskExecutor) ExecuteOnce(key string, do func() any) any {
	return c.executeWithContext(context.Background(), key, do)
}

// ExecuteOnceWithContext 带有上下文控制的任务执行函数
// 可以通过ctx控制超时或取消执行
func (c *UniqueTaskExecutor) ExecuteOnceWithContext(ctx context.Context, key string, do func() any) any {
	return c.executeWithContext(ctx, key, do)
}

// executeWithContext 内部实现，处理带有上下文的任务执行
func (c *UniqueTaskExecutor) executeWithContext(ctx context.Context, key string, do func() any) any {
	defer c.cache.Delete(key)
	v, ok := c.cache.Load(key)
	if ok {
		runner := v.(*TaskRunner)
		return runner.Wait(ctx)
	}

	v, ok = c.cache.LoadOrStore(key, NewTaskRunner())
	runner := v.(*TaskRunner)
	if !ok {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					runner.Store(fmt.Errorf("task panic: %v", r))
					runner.Done()
				}
			}()
			runner.Execute(do)
		}()
	}

	return runner.Wait(ctx)
}
