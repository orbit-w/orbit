package future

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Hello struct {
	msg string
}

func TestFuture(t *testing.T) {
	future := NewFuture()
	go func() {
		time.Sleep(1 * time.Second)
		future.Response(&Hello{msg: "hello"}, nil)
	}()
	hello, err := future.Result(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "hello", hello.(*Hello).msg)
	future.Stop()
}

func TestFutureWithTimeout(t *testing.T) {
	future := NewFuture()
	go func() {
		time.Sleep(10 * time.Second)
		future.Stop()
	}()
	err := future.WaitWithTimeout(100 * time.Millisecond)
	assert.Equal(t, true, errors.Is(err, context.DeadlineExceeded))
}

func TestFutureContextCanceled(t *testing.T) {
	future := NewFuture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := future.Wait(ctx)
	assert.Equal(t, true, errors.Is(err, context.Canceled))
}
