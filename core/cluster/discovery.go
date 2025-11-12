package cluster

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// DefaultKeyPrefix etcd key 前缀
	DefaultKeyPrefix = "/orbit/nodes/"
	// DefaultLeaseTTL 默认租约 TTL (秒)
	DefaultLeaseTTL = 30
	// DefaultHeartbeatInterval 默认心跳间隔 (秒)
	// 心跳间隔不能大于租约 TTL，否则会导致租约过期, 建议心跳间隔 = TTL / 3
	DefaultHeartbeatInterval = DefaultLeaseTTL / 3

	// DefaultDialTimeout 默认连接超时
	DefaultDialTimeout = 5 * time.Second
	// DefaultRequestTimeout 默认请求超时
	DefaultRequestTimeout = 5 * time.Second
)

// Discovery 服务发现管理器
type Discovery struct {
	client     *clientv3.Client
	leaseID    clientv3.LeaseID
	keyPrefix  string
	leaseTTL   int64
	heartbeat  time.Duration
	localNode  *Node
	nodes      cmap.ConcurrentMap[string, *Node] // 节点缓存
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	serializer Serializer
}

// DiscoveryConfig 服务发现配置
type DiscoveryConfig struct {
	Endpoints         []string      // etcd 端点列表
	KeyPrefix         string        // key 前缀，默认 /orbit/nodes/
	LeaseTTL          int64         // 租约 TTL (秒)，默认 10
	HeartbeatInterval time.Duration // 心跳间隔，默认 3 秒
	DialTimeout       time.Duration // 连接超时，默认 5 秒
	Username          string        // etcd 用户名（可选）
	Password          string        // etcd 密码（可选）
}

// NewDiscovery 创建服务发现管理器
func NewDiscovery(config DiscoveryConfig) (*Discovery, error) {
	if len(config.Endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints cannot be empty")
	}

	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultKeyPrefix
	}
	if config.LeaseTTL == 0 {
		config.LeaseTTL = DefaultLeaseTTL
	}
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = DefaultHeartbeatInterval * time.Second
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = DefaultDialTimeout
	}

	etcdConfig := clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: config.DialTimeout,
	}

	if config.Username != "" {
		etcdConfig.Username = config.Username
		etcdConfig.Password = config.Password
	}

	client, err := clientv3.New(etcdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Discovery{
		client:     client,
		keyPrefix:  config.KeyPrefix,
		leaseTTL:   config.LeaseTTL,
		heartbeat:  config.HeartbeatInterval,
		nodes:      cmap.New[*Node](),
		ctx:        ctx,
		cancel:     cancel,
		serializer: NewJSONSerializer(),
	}, nil
}

// Register 注册节点到 etcd
func (d *Discovery) Register(node *Node) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.localNode != nil {
		return fmt.Errorf("node already registered")
	}

	d.localNode = node
	d.localNode.State = NodeStateOnline
	d.localNode.CreatedAt = time.Now()
	d.localNode.UpdatedAt = time.Now()

	// 创建租约
	lease, err := d.client.Grant(d.ctx, d.leaseTTL)
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}
	d.leaseID = lease.ID

	// 注册节点
	key := d.genNodeKey(node.ID)
	value, err := d.serializeNode(node)
	if err != nil {
		return fmt.Errorf("failed to serialize node: %w", err)
	}

	_, err = d.client.Put(d.ctx, key, value, clientv3.WithLease(d.leaseID))
	if err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	// 启动心跳
	d.wg.Add(1)
	go d.keepAlive()

	// 启动 watch
	d.wg.Add(1)
	go d.watchNodes()

	return nil
}

// Unregister 注销节点
func (d *Discovery) Unregister() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.localNode == nil {
		return nil
	}

	d.cancel()

	key := d.genNodeKey(d.localNode.ID)
	_, err := d.client.Delete(d.ctx, key)
	if err != nil {
		return fmt.Errorf("failed to unregister node: %w", err)
	}

	// 撤销租约
	if d.leaseID != 0 {
		_, _ = d.client.Revoke(d.ctx, d.leaseID)
	}

	d.localNode = nil
	d.wg.Wait()

	return nil
}

// UpdateNode 更新节点信息
func (d *Discovery) UpdateNode(node *Node) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.localNode == nil {
		return fmt.Errorf("node not registered")
	}

	// 更新本地节点信息
	node.ID = d.localNode.ID
	node.Address = d.localNode.Address
	node.UpdatedAt = time.Now()
	d.localNode = node

	// 更新到 etcd
	key := d.genNodeKey(node.ID)
	value, err := d.serializeNode(node)
	if err != nil {
		return fmt.Errorf("failed to serialize node: %w", err)
	}

	_, err = d.client.Put(d.ctx, key, value, clientv3.WithLease(d.leaseID))
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	return nil
}

// GetNode 获取指定节点信息
func (d *Discovery) GetNode(nodeID string) (*Node, error) {
	if node, exists := d.nodes.Get(nodeID); exists {
		return node, nil
	}

	// 从 etcd 获取
	key := d.genNodeKey(nodeID)
	resp, err := d.client.Get(d.ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get node from etcd: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	node, err := d.deserializeNode(resp.Kvs[0].Value)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize node: %w", err)
	}

	d.nodes.Set(nodeID, node)
	return node, nil
}

// GetNodes 获取所有在线节点列表
func (d *Discovery) GetNodes() ([]*Node, error) {
	key := d.keyPrefix
	resp, err := d.client.Get(d.ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes from etcd: %w", err)
	}

	nodes := make([]*Node, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		node, err := d.deserializeNode(kv.Value)
		if err != nil {
			continue
		}
		// 只返回在线节点
		if node.State == NodeStateOnline {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// GetOnlineNodes 获取所有在线节点（从缓存）
func (d *Discovery) GetOnlineNodes() []*Node {
	nodes := make([]*Node, 0)
	d.nodes.IterCb(func(nodeID string, node *Node) {
		if node.State == NodeStateOnline {
			nodes = append(nodes, node)
		}
	})
	return nodes
}

// GetNodeCount 获取节点数量
func (d *Discovery) GetNodeCount() int {
	return d.nodes.Count()
}

// Close 关闭服务发现管理器
func (d *Discovery) Close() error {
	if err := d.Unregister(); err != nil {
		return err
	}
	return d.client.Close()
}

// keepAlive 保持租约存活（心跳）
func (d *Discovery) keepAlive() {
	defer d.wg.Done()

	ticker := time.NewTicker(d.heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.mu.RLock()
			if d.localNode == nil || d.leaseID == 0 {
				d.mu.RUnlock()
				return
			}

			// 更新心跳时间
			d.localNode.LastHeartbeat = time.Now()
			d.localNode.UpdatedAt = time.Now()

			// 更新节点信息到 etcd
			key := d.genNodeKey(d.localNode.ID)
			value, err := d.serializeNode(d.localNode)
			if err == nil {
				_, _ = d.client.Put(d.ctx, key, value, clientv3.WithLease(d.leaseID))
			}

			// 续约
			_, err = d.client.KeepAliveOnce(d.ctx, d.leaseID)
			if err != nil {
				// 续约失败，尝试重新创建租约
				if lease, err := d.client.Grant(d.ctx, d.leaseTTL); err == nil {
					d.leaseID = lease.ID
					_, _ = d.client.Put(d.ctx, key, value, clientv3.WithLease(d.leaseID))
				}
			}
			d.mu.RUnlock()
		}
	}
}

// watchNodes 监听节点变化
func (d *Discovery) watchNodes() {
	defer d.wg.Done()

	watchChan := d.client.Watch(d.ctx, d.keyPrefix, clientv3.WithPrefix())

	for {
		select {
		case <-d.ctx.Done():
			return
		case watchResp := <-watchChan:
			if watchResp.Err() != nil {
				continue
			}

			for _, event := range watchResp.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					// 节点注册或更新
					node, err := d.deserializeNode(event.Kv.Value)
					if err == nil {
						d.nodes.Set(node.ID, node)
					}
				case clientv3.EventTypeDelete:
					// 节点删除
					nodeID := d.extractNodeID(string(event.Kv.Key))
					if nodeID != "" {
						d.nodes.Remove(nodeID)
					}
				}
			}
		}
	}
}

// genNodeKey 获取节点的 etcd key
func (d *Discovery) genNodeKey(nodeID string) string {
	builder := strings.Builder{}
	builder.WriteString(d.keyPrefix)
	builder.WriteString("/")
	builder.WriteString(nodeID)
	return builder.String()
}

// extractNodeID 从 key 中提取节点 ID
func (d *Discovery) extractNodeID(key string) string {
	parts := strings.Split(key, "/")
	return parts[len(parts)-1]
}

// serializeNode 序列化节点信息
func (d *Discovery) serializeNode(node *Node) (string, error) {
	data := map[string]any{
		"id":             node.ID,
		"address":        node.Address,
		"state":          int32(node.State),
		"zone_count":     node.ZoneCount.Load(),
		"entity_count":   node.EntityCount.Load(),
		"ccu":            node.CCU.Load(),
		"last_heartbeat": node.LastHeartbeat.Unix(),
		"created_at":     node.CreatedAt.Unix(),
		"updated_at":     node.UpdatedAt.Unix(),
	}

	bytes, err := d.serializer.Serialize(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// deserializeNode 反序列化节点信息
func (d *Discovery) deserializeNode(data []byte) (*Node, error) {
	var m map[string]any
	if err := d.serializer.Deserialize(data, &m); err != nil {
		return nil, err
	}

	node := &Node{
		ID:            getString(m, "id"),
		Address:       getString(m, "address"),
		LastHeartbeat: time.Unix(getInt64(m, "last_heartbeat"), 0),
		CreatedAt:     time.Unix(getInt64(m, "created_at"), 0),
		UpdatedAt:     time.Unix(getInt64(m, "updated_at"), 0),
	}

	node.State = NodeState(getInt32(m, "state"))
	node.ZoneCount.Store(getInt32(m, "zone_count"))
	node.EntityCount.Store(getInt64(m, "entity_count"))
	node.CCU.Store(getInt32(m, "ccu"))

	return node, nil
}

// 辅助函数
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt32(m map[string]any, key string) int32 {
	switch v := m[key].(type) {
	case int32:
		return v
	case int:
		return int32(v)
	case int64:
		return int32(v)
	case float64:
		return int32(v)
	}
	return 0
}

func getInt64(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	}
	return 0
}
