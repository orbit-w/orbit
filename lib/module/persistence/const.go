package persistence

import "time"

const (
	DefaultTimeout    = time.Second * 15
	MongoWriteTimeout = time.Second * 10
	MongoReadTimeout  = time.Second * 10
)
