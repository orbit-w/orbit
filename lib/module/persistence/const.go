package persistence

import "time"

const (
	DefaultTimeout    = time.Second * 8
	MongoWriteTimeout = time.Second * 10
)
