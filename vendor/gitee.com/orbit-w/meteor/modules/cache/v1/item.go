package v1

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type Item[V proto.Message] struct {
	Value    V
	ExpireAt time.Time
}
