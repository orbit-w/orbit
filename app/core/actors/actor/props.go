package actor

import "time"

type Props struct {
	InitHandler  func() error
	Meta         *Meta
	AliveTimeout time.Duration
	kvs          map[string]any
}

func NewProps() *Props {
	return &Props{
		kvs: make(map[string]any),
	}
}

func (pp *Props) GetAliveTimeout() time.Duration {
	if pp == nil {
		return DefaultAliveTimeout
	}
	return pp.AliveTimeout
}

func (pp *Props) GetInitHandler() func() error {
	if pp == nil {
		return nil
	}
	if pp.InitHandler == nil {
		return nil
	}
	return pp.InitHandler
}

func (pp *Props) GetMeta() *Meta {
	if pp == nil {
		return nil
	}
	if pp.Meta == nil {
		return nil
	}
	return pp.Meta
}

func (pp *Props) GetKvs(iter func(k string, v any)) {
	if pp == nil {
		return
	}
	if pp.kvs == nil {
		return
	}
	for k, v := range pp.kvs {
		iter(k, v)
	}
}

func (pp *Props) getOrCreateActorPID(name, pattern string) *Process {
	p, err := GetOrStartActor(name, pattern, pp)
	if err != nil {
		panic(err)
	}
	return p
}

type PropsOption func(pp *Props)

func WithInitHandler(handler func() error) PropsOption {
	return func(pp *Props) {
		pp.InitHandler = handler
	}
}

func WithMeta(meta *Meta) PropsOption {
	return func(pp *Props) {
		pp.Meta = meta
	}
}

// WithAliveTimeout 配置Actor的活跃超时时间
//
// 参数:
//   - timeout: 活跃超时时间，如果Actor在此时间内没有收到任何消息或处理任何事件，
//     将被视为不活跃。系统可能会根据此配置停止或重启不活跃的Actor。
//
// 详细说明:
//  1. 每个Actor都有一个最后活动时间戳，当收到消息或处理事件时会更新此时间戳
//  2. 系统会定期检查Actor的活跃状态，如果发现Actor超过指定的timeout时间没有活动，
//     将触发清理机制，可能会停止或重启该Actor
//  3. 这个机制有助于识别并清理"僵尸"Actor，避免资源泄漏
//  4. 如果不设置此选项，将使用系统默认的超时时间(通常为30分钟)
//
// 使用场景:
//   - 限制长时间不活跃的Actor占用系统资源
//   - 对不同类型的Actor设置不同的活跃检测策略
//   - 确保关键Actor在长时间不活跃时能被重启
//
// 示例:
//
//	props := NewProps(
//	  WithAliveTimeout(30 * time.Minute),  // 设置30分钟活跃超时
//	)
//	actorRef := NewActorRef(props, "my-actor", "my-pattern")
func WithAliveTimeout(timeout time.Duration) PropsOption {
	return func(pp *Props) {
		pp.AliveTimeout = timeout
	}
}
