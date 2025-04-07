package future

import (
	"context"
	"errors"
	"time"

	"sync/atomic"
)

/*
   @Author: orbit-w
   @File: request
   @2025 4月 周六 19:01
*/

var (
	// ErrFutureCanceled 表示 Future 已被取消
	ErrFutureCanceled = errors.New("future has been canceled")

	// ErrFutureContextCanceled 表示 Future 因上下文取消而终止
	ErrFutureContextCanceled = errors.New("future canceled due to context cancelation")
)

const (
	StateNormal = iota
	StateResponded
	StateStopped
	StateContextCanceled
)

// Future 表示一个异步操作的结果
// 使用场景：
// 单独的接收者，需要等待异步操作的结果
// 例如：
// 1. 需要异步执行一个操作，并获取其结果
// 2. 需要异步执行一个操作，并获取其结果，并设置超时时间
// 3. 需要异步执行一个操作，并获取其结果，并设置超时时间，并设置上下文取消
type Future struct {
	ch    chan Message
	state atomic.Uint32
}

func NewFuture() *Future {
	return &Future{
		ch: make(chan Message, 1), // 缓冲大小为 1，避免发送方阻塞
	}
}

func (f *Future) Stop() {
	if f.state.CompareAndSwap(StateNormal, StateStopped) {
		// 标记为已关闭并关闭通道
		close(f.ch)
	}
}

// WaitWithTimeout 等待 Future 完成，并设置超时时间
// 如果 Future 已关闭或因上下文取消而终止，则返回相应的错误
// 不可重入
func (f *Future) WaitWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return f.Wait(ctx)
}

// Wait 等待 Future 完成
// 如果 Future 已关闭或因上下文取消而终止，则返回相应的错误
// 不可重入
func (f *Future) Wait(ctx context.Context) error {
	_, err := f.wait(ctx)
	return err
}

func (f *Future) Response(msg any, err error) error {
	if !f.state.CompareAndSwap(StateNormal, StateResponded) {
		switch f.state.Load() {
		case StateStopped:
			return ErrFutureCanceled
		case StateContextCanceled:
			return ErrFutureContextCanceled
		default:
			return ErrFutureCanceled
		}
	}

	select {
	case f.ch <- Message{msg, err}:
	default:
	}
	return nil
}

// Result 获取 Future 的结果
// 如果 Future 已关闭或因上下文取消而终止，则返回相应的错误
// 不可重入
func (f *Future) Result(ctx context.Context) (any, error) {
	return f.wait(ctx)
}

func (f *Future) wait(ctx context.Context) (any, error) {
	select {
	case msg, ok := <-f.ch:
		if !ok {
			return nil, ErrFutureCanceled
		}
		return msg.Get()
	case <-ctx.Done():
		f.state.CompareAndSwap(StateNormal, StateContextCanceled)
		return nil, ctx.Err()
	}
}

type Message struct {
	msg any
	err error
}

func (m *Message) Get() (any, error) {
	return m.msg, m.err
}
