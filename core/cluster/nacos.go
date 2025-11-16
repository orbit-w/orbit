package cluster

import (
	"context"
	"fmt"
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

// NacosConfig Nacos 配置（定义在 config 包中，这里使用别名引用）
type NacosConfig = config.NacosConfig

// NacosRegistry Nacos 服务注册发现
type NacosRegistry struct {
	config       *NacosConfig
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
func NewNacosRegistry(config *NacosConfig, nodeID, nodeAddress, serviceName string) (*NacosRegistry, error) {
	if config == nil {
		return nil, fmt.Errorf("nacos config is nil")
	}

	// 设置默认值
	if config.GroupName == "" {
		config.GroupName = "DEFAULT_GROUP"
	}
	if config.TimeoutMs == 0 {
		config.TimeoutMs = 5000
	}
	if config.BeatInterval == 0 {
		config.BeatInterval = 5
	}
	if config.LogDir == "" {
		config.LogDir = "/tmp/nacos/log"
	}
	if config.CacheDir == "" {
		config.CacheDir = "/tmp/nacos/cache"
	}

	// 构建服务器配置
	serverConfigs := make([]constant.ServerConfig, 0, len(config.ServerHosts))
	for _, host := range config.ServerHosts {
		// 解析服务器地址（可能包含端口）
		serverIP, serverPort, err := parseAddress(host)
		if err != nil {
			// 如果解析失败，使用默认端口
			serverIP = host
			serverPort = 8848
		}
		if serverPort == 0 {
			serverPort = 8848 // 默认 Nacos 端口
		}
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      serverIP,
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
		LogDir:              config.LogDir,
		CacheDir:            config.CacheDir,
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

// parseAddress 解析地址字符串为 IP 和 Port
func parseAddress(address string) (string, uint64, error) {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		// 如果没有端口，使用默认端口 0
		host = address
		portStr = "0"
	}

	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %s", portStr)
	}

	return host, port, nil
}

// Register 注册服务节点
func (r *NacosRegistry) Register(metadata map[string]string) error {
	if metadata != nil {
		r.metadata = metadata
	}

	// 添加节点ID到元数据
	if r.metadata == nil {
		r.metadata = make(map[string]string)
	}
	r.metadata["node_id"] = r.nodeID

	// 注册服务实例
	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          r.nodeIP,
		Port:        r.nodePort,
		ServiceName: r.serviceName,
		GroupName:   r.config.GroupName,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    r.metadata,
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
		zap.String("group", r.config.GroupName))

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
		GroupName:   r.config.GroupName,
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
				r.updateNodeMetadata()
			}
		}
	}()
}

// updateNodeMetadata 更新节点元数据（用于心跳时同步最新状态）
func (r *NacosRegistry) updateNodeMetadata() error {
	// 可以在这里更新节点的实时状态信息
	// 例如：CCU、EntityCount、ZoneCount 等
	r.mu.RLock()
	metadata := make(map[string]string)
	for k, v := range r.metadata {
		metadata[k] = v
	}
	r.mu.RUnlock()

	success, err := r.namingClient.UpdateInstance(vo.UpdateInstanceParam{
		Ip:          r.nodeIP,
		Port:        r.nodePort,
		ServiceName: r.serviceName,
		GroupName:   r.config.GroupName,
		Metadata:    metadata,
	})
	if err != nil {
		logger.GetLogger().Error("update instance metadata failed", zap.Error(err))
		return err
	}
	if !success {
		logger.GetLogger().Warn("update instance metadata failed: success=false")
	}
	return nil
}

// startDiscovery 启动服务发现
func (r *NacosRegistry) startDiscovery() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		// 立即获取一次服务列表
		r.discoverServices()

		// 订阅服务变化
		err := r.namingClient.Subscribe(&vo.SubscribeParam{
			ServiceName:       r.serviceName,
			GroupName:         r.config.GroupName,
			SubscribeCallback: r.onServiceChange,
		})
		if err != nil {
			logger.GetLogger().Error("subscribe service failed", zap.Error(err))
			return
		}

		// 等待退出
		<-r.ctx.Done()

		// 取消订阅（通过传入 nil callback 来取消）
		_ = r.namingClient.Subscribe(&vo.SubscribeParam{
			ServiceName:       r.serviceName,
			GroupName:         r.config.GroupName,
			SubscribeCallback: nil,
		})
	}()
}

// discoverServices 发现服务列表
func (r *NacosRegistry) discoverServices() {
	instances, err := r.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: r.serviceName,
		GroupName:   r.config.GroupName,
		HealthyOnly: true,
	})
	if err != nil {
		logger.GetLogger().Error("discover services failed", zap.Error(err))
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新节点列表
	newNodes := make(map[string]*Node)
	for _, instance := range instances {
		nodeID := r.extractNodeID(instance)
		if nodeID == "" {
			continue
		}

		node := &Node{
			ID:            nodeID,
			Address:       fmt.Sprintf("%s:%d", instance.Ip, instance.Port),
			State:         NodeStateOnline,
			LastHeartbeat: time.Now(),
			UpdatedAt:     time.Now(),
		}

		// 元数据已经是 map[string]string 类型
		if len(instance.Metadata) > 0 {
			// 可以从元数据中恢复其他信息
		}

		newNodes[nodeID] = node
	}

	r.nodes = newNodes
	logger.GetLogger().Info("discover services success", zap.Int("count", len(r.nodes)))
}

// onServiceChange 服务变化回调
func (r *NacosRegistry) onServiceChange(services []model.Instance, err error) {
	if err != nil {
		logger.GetLogger().Error("service change callback error", zap.Error(err))
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新节点列表
	newNodes := make(map[string]*Node)
	for _, instance := range services {
		nodeID := r.extractNodeID(instance)
		if nodeID == "" {
			continue
		}

		state := NodeStateOnline
		if !instance.Healthy {
			state = NodeStateOffline
		}

		node := &Node{
			ID:            nodeID,
			Address:       fmt.Sprintf("%s:%d", instance.Ip, instance.Port),
			State:         state,
			LastHeartbeat: time.Now(),
			UpdatedAt:     time.Now(),
		}

		// 元数据已经是 map[string]string 类型
		if len(instance.Metadata) > 0 {
			// 可以从元数据中恢复其他信息
		}

		newNodes[nodeID] = node
	}

	// 记录变化
	oldCount := len(r.nodes)
	r.nodes = newNodes
	logger.GetLogger().Info("service change detected",
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
func (r *NacosRegistry) GetNodes() map[string]*Node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make(map[string]*Node)
	for k, v := range r.nodes {
		nodes[k] = v
	}
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
func (r *NacosRegistry) UpdateNodeMetadata(metadata map[string]string) error {
	r.mu.Lock()
	if metadata != nil {
		r.metadata = metadata
	}
	if r.metadata == nil {
		r.metadata = make(map[string]string)
	}
	r.metadata["node_id"] = r.nodeID
	r.mu.Unlock()

	return r.updateNodeMetadata()
}
