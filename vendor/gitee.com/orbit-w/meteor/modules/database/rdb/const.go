package rdb

import "time"

const (
	DefaultDialTimeout  = 5 * time.Second
	DefaultReadTimeout  = 30 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultPoolTimeout  = 30 * time.Second
)
