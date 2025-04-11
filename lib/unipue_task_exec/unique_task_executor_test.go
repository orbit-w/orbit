package unipue_task_exec

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestUniqueTaskExecutor_ExecuteOnce(t *testing.T) {
	// 测试基本功能，确保任务只执行一次
	t.Run("执行次数测试", func(t *testing.T) {
		executor := NewUniqueTaskExecutor()
		counter := 0

		// 创建同一个key的多个并发调用
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result := executor.ExecuteOnce("test-key", func() any {
					counter++
					time.Sleep(10 * time.Millisecond) // 模拟工作
					return counter
				})

				// 所有协程应该得到相同的结果 1
				if result != 1 {
					t.Errorf("预期结果为1，实际获得: %v", result)
				}
			}()
		}
		wg.Wait()

		// 确认计数器只增加了一次
		if counter != 1 {
			t.Errorf("任务应该只执行一次，但执行了 %d 次", counter)
		}
	})

	// 测试不同key应该独立执行
	t.Run("不同key独立执行", func(t *testing.T) {
		executor := NewUniqueTaskExecutor()
		results := make(map[string]int)
		var mu sync.Mutex // 保护 results map

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key-%d", i)
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				result := executor.ExecuteOnce(k, func() any {
					time.Sleep(10 * time.Millisecond)
					mu.Lock()
					defer mu.Unlock()
					if _, exists := results[k]; !exists {
						results[k] = 1
					} else {
						results[k]++
					}
					return results[k]
				})

				if result != 1 {
					t.Errorf("key %s 预期结果为1，实际获得: %v", k, result)
				}
			}(key)
		}
		wg.Wait()

		// 验证每个key的任务只执行了一次
		if len(results) != 5 {
			t.Errorf("应该有5个不同的key，但实际有 %d 个", len(results))
		}
		for k, v := range results {
			if v != 1 {
				t.Errorf("key %s 任务应该只执行一次，但执行了 %d 次", k, v)
			}
		}
	})

	// 测试错误处理和panic恢复
	t.Run("错误处理与panic恢复", func(t *testing.T) {
		executor := NewUniqueTaskExecutor()

		// 测试返回错误
		result := executor.ExecuteOnce("error-key", func() any {
			return errors.New("测试错误")
		})

		err, ok := result.(error)
		if !ok {
			t.Errorf("预期返回error类型，得到: %T", result)
		} else if err.Error() != "测试错误" {
			t.Errorf("错误消息不匹配，预期: '测试错误'，得到: '%v'", err)
		}

		// 测试panic恢复
		result = executor.ExecuteOnce("panic-key", func() any {
			panic("测试panic")
			return nil // 永远不会执行到
		})

		err, ok = result.(error)
		if !ok {
			t.Errorf("预期返回error类型，得到: %T", result)
		} else if err.Error() != "task panic: 测试panic" {
			t.Errorf("错误消息不匹配，预期包含: 'task panic: 测试panic'，得到: '%v'", err)
		}
	})
}

func TestUniqueTaskExecutor_ExecuteOnceWithContext(t *testing.T) {
	// 测试上下文取消
	t.Run("上下文取消", func(t *testing.T) {
		executor := NewUniqueTaskExecutor()

		// 创建可取消的上下文
		ctx, cancel := context.WithCancel(context.Background())

		// 在任务开始后但完成前取消上下文
		var startedCh = make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			result := executor.ExecuteOnceWithContext(ctx, "cancel-test", func() any {
				close(startedCh)                   // 通知任务已开始
				time.Sleep(200 * time.Millisecond) // 长时间运行任务
				return "完成"
			})

			// 这里应该得到上下文取消的错误
			_, ok := result.(error)
			if !ok {
				t.Errorf("上下文取消后应返回错误，实际得到: %v", result)
			}
		}()

		// 等待任务开始
		<-startedCh
		// 取消上下文
		cancel()
		wg.Wait()
	})

	// 测试上下文超时
	t.Run("上下文超时", func(t *testing.T) {
		executor := NewUniqueTaskExecutor()

		// 创建一个很短超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		result := executor.ExecuteOnceWithContext(ctx, "timeout-test", func() any {
			time.Sleep(200 * time.Millisecond) // 长时间运行的任务
			return "完成"
		})

		// 应该因为超时而返回错误
		_, ok := result.(error)
		if !ok {
			t.Errorf("上下文超时后应返回错误，实际得到: %v", result)
		}
	})

	// 测试多个协程等待同一个任务结果，但上下文被取消
	t.Run("多协程上下文取消", func(t *testing.T) {
		executor := NewUniqueTaskExecutor()

		// 主执行协程使用长超时
		mainCtx, mainCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer mainCancel()

		// 其他协程使用短超时
		shortCtx, shortCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer shortCancel()

		// 启动长任务
		var wg sync.WaitGroup
		wg.Add(2)

		// 主协程
		go func() {
			defer wg.Done()
			result := executor.ExecuteOnceWithContext(mainCtx, "shared-task", func() any {
				time.Sleep(100 * time.Millisecond)
				return "成功完成"
			})

			// 主协程应该成功获得结果
			if result != "成功完成" {
				t.Errorf("主协程应收到成功结果，实际得到: %v", result)
			}
		}()

		// 短超时协程
		time.Sleep(5 * time.Millisecond) // 确保主协程先启动
		go func() {
			defer wg.Done()
			result := executor.ExecuteOnceWithContext(shortCtx, "shared-task", func() any {
				// 这个函数不会被调用，因为任务已经在执行
				return "不会执行"
			})

			// 短超时协程应该因为上下文取消而收到错误
			_, ok := result.(error)
			if !ok {
				t.Errorf("短超时协程应收到错误，实际得到: %v", result)
			}
		}()

		wg.Wait()
	})
}

// 测试TaskRunner
func TestTaskRunner(t *testing.T) {
	t.Run("基本功能", func(t *testing.T) {
		runner := NewTaskRunner()

		go runner.Execute(func() any {
			return "成功"
		})

		result := runner.Wait(context.Background())
		if result != "成功" {
			t.Errorf("预期结果为'成功'，实际获得: %v", result)
		}
	})

	t.Run("错误结果", func(t *testing.T) {
		runner := NewTaskRunner()

		go runner.Execute(func() any {
			return errors.New("任务错误")
		})

		result := runner.Wait(context.Background())
		err, ok := result.(error)
		if !ok {
			t.Errorf("预期结果为error，实际获得: %T", result)
		} else if err.Error() != "任务错误" {
			t.Errorf("错误消息不匹配")
		}
	})

	t.Run("上下文取消", func(t *testing.T) {
		runner := NewTaskRunner()
		ctx, cancel := context.WithCancel(context.Background())

		// 立即取消上下文
		cancel()

		// 这里不执行任务，直接等待应该收到取消错误
		result := runner.Wait(ctx)
		_, ok := result.(error)
		if !ok {
			t.Errorf("上下文取消后应返回错误，实际得到: %v", result)
		}
	})
}
