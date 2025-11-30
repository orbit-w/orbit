package subpub_redis

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync/atomic"

	"gitee.com/orbit-w/meteor/modules/mlog"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	proto "google.golang.org/protobuf/proto"
)

type IPubSub interface {
	Publish(pid int64, v any) error
	Subscribe()
	Stop()
}

type PubSub struct {
	Uuid        string
	state       atomic.Uint32
	encoderEnum int    //JSON｜Proto3: 默认编码协议是JSON
	topic       string //主题名称
	cli         redis.UniversalClient
	sub         *redis.PubSub
	log         *mlog.Logger
	invoker     func(pid int64, body []byte, err error)
}

var (
	ctx = context.Background()
)

type IEncoder interface {
	Marshal(v any) ([]byte, error)
}

func NewPubSub(_cli redis.UniversalClient, ee int, topic string, _invoker func(pid int64, body []byte, err error)) IPubSub {
	return &PubSub{
		Uuid:        uuid.New().String(),
		topic:       topic,
		invoker:     _invoker,
		cli:         _cli,
		encoderEnum: ee,
		log:         mlog.WithPrefix("subpub_redis"),
	}
}

func (ps *PubSub) Publish(pid int64, v any) error {
	body, err := encode(ps.encoderEnum, ps.Uuid, pid, v)
	if err != nil {
		return ErrPublish(err)
	}
	err = ps.cli.Publish(ctx, ps.topic, body).Err()
	if err != nil {
		return ErrPublish(err)
	}
	return nil
}

func (ps *PubSub) Subscribe() {
	ps.subscribe(ps.decodeAndInvoke)
}

func (ps *PubSub) Stop() {
	if ps.state.CompareAndSwap(stateReady, stateStopped) {
		if ps.sub != nil {
			_ = ps.sub.Close()
		}
	}
}

func (ps *PubSub) subscribe(handle func(msg *redis.Message)) {
	pubSub := ps.cli.Subscribe(ctx, ps.topic)
	ps.sub = pubSub
	ch := pubSub.Channel()

	go func() {
		defer func() {
			if ps.sub != nil {
				_ = ps.sub.Close()
			}
		}()
		for msg := range ch {
			handle(msg)
		}
	}()
}

// Invoke failed, the business side is aware of the error.
func (ps *PubSub) decodeAndInvoke(msg *redis.Message) {
	var (
		err error
		pb  = new(PubMessage)
	)

	if err = proto.Unmarshal([]byte(msg.Payload), pb); err != nil {
		err = fmt.Errorf("[PubSub] decode proto failed : %w", err)
	}

	if pb.Sender == ps.Uuid {
		return
	}

	ps.invoke(pb.Pid, pb.Data, err)
}

func (ps *PubSub) invoke(pid int64, data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			log.Println("Stack: ", string(debug.Stack()))
		}
	}()
	ps.invoker(pid, data, err)
}
