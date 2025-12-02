package cluster

import (
	"context"
	"fmt"
	"maps"
	"net"
	"strconv"
	"sync"
	"time"

	"gitee.com/orbit-w/orbit/app/modules/config"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

// NacosRegistry Nacos 服务注册发现
type NacosRegistry struct {
	config       *config.NacosConfig
	namingClient naming_client.INamingClient
	nodeID       string
	nodeIP       string
	nodePort     uint64
	serviceName  string
	metadata     map[string]string
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.RWMutex
	nodes        map[string]*Node // 节点ID -> Node
}

// NewNacosRegistry 创建 Nacos 注册发现实例
func NewNacosRegistry(config *config.NacosConfig, nodeID, nodeAddress, serviceName string) (*NacosRegistry, error) {
	// 设置默认值
	if config.TimeoutMs == 0 {
		config.TimeoutMs = 5000
	}
	if config.BeatInterval == 0 {
		config.BeatInterval = 5
	}
	// 构建服务器配置
	if len(config.ServerHosts) == 0 {
		return nil, fmt.Errorf("nacos server hosts is empty")
	}

	serverConfigs := make([]constant.ServerConfig, 0, len(config.ServerHosts))
	for _, host := range config.ServerHosts {
		// 解析服务器地址（支持域名，无端口时使用默认端口 8848）
		serverHost, serverPort := parseServerAddress(host)
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      serverHost,
			Port:        serverPort,
			ContextPath: "/nacos",
		})
	}

	// 构建客户端配置
	beatInterval := time.Duration(config.BeatInterval) * time.Second
	clientConfig := constant.ClientConfig{
		NamespaceId:         config.NamespaceID,
		TimeoutMs:           config.TimeoutMs,
		BeatInterval:        int64(beatInterval / time.Millisecond), // 转换为毫秒
		UpdateThreadNum:     10,
		NotLoadCacheAtStart: true,
		LogLevel:            "info",
	}

	// 如果有用户名密码，设置认证
	if config.Username != "" && config.Password != "" {
		clientConfig.Username = config.Username
		clientConfig.Password = config.Password
	}

	// 创建 Nacos 客户端
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create nacos client failed: %w", err)
	}

	// 解析节点地址
	nodeIP, nodePort, err := parseAddress(nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("parse node address failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	registry := &NacosRegistry{
		config:       config,
		namingClient: namingClient,
		nodeID:       nodeID,
		nodeIP:       nodeIP,
		nodePort:     nodePort,
		serviceName:  serviceName,
		metadata:     make(map[string]string),
		ctx:          ctx,
		cancel:       cancel,
		nodes:        make(map[string]*Node),
	}

	return registry, nil
}

// parseServerAddress 解析服务器地址（支持域名，无端口时使用默认端口 8848）
// 用于解析 Nacos 服务器地址，支持阿里云等使用域名的场景
func parseServerAddress(address string) (string, uint64) {
	if address == "" {
		return "", 8848 // 默认端口
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		// 如果没有端口，使用默认端口 8848
		return address, 8848
	}

	if host == "" {
		// 如果 host 为空，使用整个地址作为 host
		return address, 8848
	}

	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return host, 8848
	}

	if port == 0 {
		port = 8848 // 确保端口不为 0
	}

	return host, port
}

// parseAddress 解析地址字符串为 IP 和 Port（必须包含端口）
// 用于解析节点地址，节点地址必须包含端口
func parseAddress(address string) (string, uint64, error) {
	if address == "" {
		return "", 0, fmt.Errorf("address is empty")
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		// 如果没有端口，返回错误而不是使用默认值
		// 这样可以明确告知调用者地址格式不正确
		return "", 0, fmt.Errorf("invalid address format %q: %w", address, err)
	}

	if host == "" {
		return "", 0, fmt.Errorf("host is empty in address %q", address)
	}

	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port %q in address %q: %w", portStr, address, err)
	}

	return host, port, nil
}

// prepareMetadata 准备元数据，确保 node_id 始终存在
// 如果 metadata 为 nil，则使用 r.metadata 中存储的元数据
func (r *NacosRegistry) prepareMetadata(metadata map[string]string) map[string]string {
	var finalMetadata map[string]string

	if metadata != nil {
		// 使用传入的元数据
		finalMetadata = make(map[string]string, len(metadata)+1)
		maps.Copy(finalMetadata, metadata)
	} else {
		// 使用存储的元数据（心跳时使用）
		r.mu.RLock()
		finalMetadata = make(map[string]string, len(r.metadata))
		maps.Copy(finalMetadata, r.metadata)
		r.mu.RUnlock()
	}

	// 确保节点ID始终在元数据中
	finalMetadata["node_id"] = r.nodeID

	return finalMetadata
}

// setMetadata 设置并保存元数据到 r.metadata
func (r *NacosRegistry) setMetadata(metadata map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if metadata != nil {
		// 如果传入新的元数据，创建副本并保存
		r.metadata = make(map[string]string, len(metadata)+1)
		maps.Copy(r.metadata, metadata)
	} else if r.metadata == nil {
		// 如果 metadata 为 nil 且 r.metadata 也为 nil，初始化为空 map
		r.metadata = make(map[string]string, 1)
	}
	// 确保节点ID始终在元数据中
	r.metadata["node_id"] = r.nodeID
}

// getMetadataCopy 获取元数据的副本
func (r *NacosRegistry) getMetadataCopy() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadataCopy := make(map[string]string, len(r.metadata))
	maps.Copy(metadataCopy, r.metadata)
	return metadataCopy
}

// Register 注册服务节点
func (r *NacosRegistry) Register(metadata map[string]string) error {
	// 设置并保存元数据
	r.setMetadata(metadata)

	// 获取元数据副本用于注册
	metadataCopy := r.getMetadataCopy()

	// 注册服务实例
	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          r.nodeIP,
		Port:        r.nodePort,
		ServiceName: r.serviceName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metadataCopy,
	})
	if err != nil {
		return fmt.Errorf("register instance failed: %w", err)
	}
	if !success {
		return fmt.Errorf("register instance failed: success=false")
	}

	logger.GetLogger().Info("nacos register success",
		zap.String("node_id", r.nodeID),
		zap.String("ip", r.nodeIP),
		zap.Uint64("port", r.nodePort),
		zap.String("service", r.serviceName),
		zap.String("namespace", r.config.NamespaceID))

	// 启动心跳
	r.startHeartbeat()

	// 启动服务发现
	r.startDiscovery()

	return nil
}

// Deregister 注销服务节点
func (r *NacosRegistry) Deregister() error {
	success, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          r.nodeIP,
		Port:        r.nodePort,
		ServiceName: r.serviceName,
		Ephemeral:   true,
	})
	if err != nil {
		return fmt.Errorf("deregister instance failed: %w", err)
	}
	if !success {
		return fmt.Errorf("deregister instance failed: success=false")
	}

	logger.GetLogger().Info("nacos deregister success",
		zap.String("node_id", r.nodeID),
		zap.String("ip", r.nodeIP),
		zap.Uint64("port", r.nodePort),
		zap.String("service", r.serviceName))

	// 停止心跳和发现
	r.cancel()
	r.wg.Wait()

	return nil
}

// startHeartbeat 启动心跳
func (r *NacosRegistry) startHeartbeat() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(time.Duration(r.config.BeatInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				// 心跳通过 Nacos 客户端的 BeatInterval 自动处理
				// 这里可以添加额外的健康检查逻辑
				// 心跳时使用存储的元数据（传入 nil）
				if err := r.updateNodeMetadata(nil); err != nil {
					logger.GetLogger().Warn("heartbeat update metadata failed",
						zap.String("node_id", r.nodeID),
						zap.Error(err))
					// 不返回错误，继续心跳循环
				}
			}
		}
	}()
}

// updateNodeMetadata 更新节点元数据到 Nacos
// metadata: 可选的元数据，如果为 nil 则使用 r.metadata 中存储的元数据
func (r *NacosRegistry) updateNodeMetadata(metadata map[string]string) error {
	// 准备元数据（确保 node_id 存在）
	finalMetadata := r.prepareMetadata(metadata)

	success, err := r.namingClient.UpdateInstance(vo.UpdateInstanceParam{
		Ip:          r.nodeIP,
		Port:        r.nodePort,
		ServiceName: r.serviceName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    finalMetadata,
	})
	if err != nil {
		logger.GetLogger().Error("update instance metadata failed",
			zap.String("node_id", r.nodeID),
			zap.String("service", r.serviceName),
			zap.Error(err))
		return fmt.Errorf("update instance metadata failed: %w", err)
	}
	if !success {
		logger.GetLogger().Warn("update instance metadata failed: success=false",
			zap.String("node_id", r.nodeID),
			zap.String("service", r.serviceName))
		return fmt.Errorf("update instance metadata failed: success=false")
	}
	logger.GetLogger().Info("update instance metadata success",
		zap.String("node_id", r.nodeID),
		zap.String("service", r.serviceName))
	return nil
}

// startDiscovery 启动服务发现
func (r *NacosRegistry) startDiscovery() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		// 订阅服务变化（订阅会自动获取一次服务列表并触发回调）
		err := r.namingClient.Subscribe(&vo.SubscribeParam{
			ServiceName:       r.serviceName,
			SubscribeCallback: r.onServiceChange,
		})
		if err != nil {
			logger.GetLogger().Error("subscribe service failed",
				zap.String("service", r.serviceName),
				zap.String("namespace", r.config.NamespaceID),
				zap.Error(err))
			return
		}

		logger.GetLogger().Info("service discovery subscribed",
			zap.String("service", r.serviceName),
			zap.String("namespace", r.config.NamespaceID))

		// 等待退出
		<-r.ctx.Done()

		// 取消订阅（通过传入 nil callback 来取消）
		_ = r.namingClient.Subscribe(&vo.SubscribeParam{
			ServiceName:       r.serviceName,
			SubscribeCallback: nil,
		})
	}()
}

// buildNodeFromInstance 从 Nacos 实例构建 Node 对象
func (r *NacosRegistry) buildNodeFromInstance(instance model.Instance) *Node {
	nodeID := r.extractNodeID(instance)
	if nodeID == "" {
		return nil
	}

	state := NodeStateOnline
	if !instance.Healthy {
		state = NodeStateOffline
	}

	now := time.Now()
	node := &Node{
		ID:        nodeID,
		Address:   fmt.Sprintf("%s:%d", instance.Ip, instance.Port),
		CreatedAt: now,
	}
	node.SetState(state)
	node.SetLastHeartbeat(now)
	node.SetUpdatedAt(now)

	// 可以从元数据中恢复其他信息（如 CCU、EntityCount 等）
	// 如果需要，可以在这里解析元数据并设置到 Node 中

	return node
}

// discoverServices 发现服务列表
func (r *NacosRegistry) discoverServices() {
	instances, err := r.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: r.serviceName,
		HealthyOnly: true,
	})
	if err != nil {
		logger.GetLogger().Error("discover services failed",
			zap.String("service", r.serviceName),
			zap.String("namespace", r.config.NamespaceID),
			zap.Error(err))
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新节点列表
	newNodes := make(map[string]*Node, len(instances))
	for _, instance := range instances {
		node := r.buildNodeFromInstance(instance)
		if node != nil {
			newNodes[node.ID] = node
		}
	}

	r.nodes = newNodes
	logger.GetLogger().Info("discover services success",
		zap.String("service", r.serviceName),
		zap.Int("count", len(r.nodes)))
}

// onServiceChange 服务变化回调
func (r *NacosRegistry) onServiceChange(services []model.Instance, err error) {
	if err != nil {
		logger.GetLogger().Error("service change callback error",
			zap.String("service", r.serviceName),
			zap.Error(err))
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新节点列表
	oldCount := len(r.nodes)
	newNodes := make(map[string]*Node, len(services))
	for _, instance := range services {
		node := r.buildNodeFromInstance(instance)
		if node != nil {
			newNodes[node.ID] = node
		}
	}

	r.nodes = newNodes
	logger.GetLogger().Info("service change detected",
		zap.String("service", r.serviceName),
		zap.Int("old_count", oldCount),
		zap.Int("new_count", len(r.nodes)))
}

// extractNodeID 从实例中提取节点ID
func (r *NacosRegistry) extractNodeID(instance model.Instance) string {
	if len(instance.Metadata) > 0 {
		if nodeID, ok := instance.Metadata["node_id"]; ok {
			return nodeID
		}
	}
	// 如果没有元数据，使用 IP:Port 作为节点ID
	return fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
}

// GetNodes 获取所有节点
// 返回节点的副本，避免外部修改影响内部状态
func (r *NacosRegistry) GetNodes() map[string]*Node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make(map[string]*Node, len(r.nodes))
	maps.Copy(nodes, r.nodes)
	return nodes
}

// GetNode 获取指定节点
func (r *NacosRegistry) GetNode(nodeID string) (*Node, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, ok := r.nodes[nodeID]
	return node, ok
}

// UpdateNodeMetadata 更新节点元数据
// 更新元数据到 Nacos，并保存到 r.metadata 中供后续心跳使用
func (r *NacosRegistry) UpdateNodeMetadata(metadata map[string]string) error {
	if metadata == nil {
		// 如果没有传入元数据，使用当前存储的元数据更新
		return r.updateNodeMetadata(nil)
	}

	// 设置并保存元数据
	r.setMetadata(metadata)

	// 更新到 Nacos
	return r.updateNodeMetadata(metadata)
}

// GetCurrentNodeID 获取当前节点ID
func (r *NacosRegistry) GetCurrentNodeID() string {
	return r.nodeID
}
