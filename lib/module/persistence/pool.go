package persistence

import (
	"strings"
	"sync"
)

var (
	collectionKeyPool sync.Pool
)

func init() {
	collectionKeyPool = sync.Pool{
		New: func() any {
			return &strings.Builder{}
		},
	}
}
