package multiplexers_lib

/*
   @Author: orbit-w
   @File: mux_lib.go
   @2025 3月 周一 23:50
   @Update: 2025 3月 周二 22:21
*/

import (
	"context"
	"errors"
	"sync"

	multiplexers "gitee.com/orbit-w/mux-go/multiplexers"
)

var (
	// muxMap stores multiplexers for different hosts
	muxMap sync.Map
	// ErrHostNotSet is returned when attempting to Dial without a valid host
	ErrHostNotSet = errors.New("host not set")
)

// Dial creates a new connection to the specified targetHost
// If a multiplexer for this host doesn't exist, it will be created
// This implements passive instantiation of multiplexers for multiple hosts
func Dial(targetHost string, ctx context.Context) (multiplexers.IConn, error) {
	if targetHost == "" {
		return nil, ErrHostNotSet
	}

	// Get or create multiplexer for this host
	value, _ := muxMap.LoadOrStore(targetHost, multiplexers.NewWithDefaultConf(targetHost))
	mux, ok := value.(*multiplexers.Multiplexers)
	if !ok || mux == nil {
		return nil, errors.New("failed to create multiplexer")
	}

	return mux.Dial(ctx)
}

// Close closes the multiplexer for the specified host
func Close(host string) {
	if host == "" {
		return
	}

	// Close specific multiplexer
	value, ok := muxMap.Load(host)
	if ok {
		if mux, ok := value.(*multiplexers.Multiplexers); ok && mux != nil {
			mux.Close()
		}
		muxMap.Delete(host)
	}
}

// CloseAll closes all multiplexers
func CloseAll() {
	// Close all multiplexers
	muxMap.Range(func(key, value any) bool {
		if mux, ok := value.(*multiplexers.Multiplexers); ok && mux != nil {
			mux.Close()
		}
		muxMap.Delete(key)
		return true
	})
}
