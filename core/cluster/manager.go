package cluster

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"gitee.com/orbit-w/orbit/config"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"go.uber.org/zap"
)

var manager *Manager

// Manager 集群管理器
type Manager struct {
	serviceName    string             //服务名称，用于在 Nacos 中注册的服务名
	registry       *NacosRegistry     //Nacos注册发现实例
	currentNode    *Node              //当前节点
	updateCtx      context.Context    //定时更新节点信息的上下文
	updateCancel   context.CancelFunc //定时更新节点信息的取消函数
	updateWg       sync.WaitGroup     //定时更新节点信息的等待组
	updateInterval time.Duration      // 节点信息更新间隔，默认15秒
}

// NewManager 创建集群管理器
func NewManager(serviceName string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	manager = &Manager{
		serviceName:    serviceName,
		updateCtx:      ctx,
		updateCancel:   cancel,
		updateInterval: 15 * time.Second, // 默认15秒更新一次
	}
	return manager
}

func GetManager() *Manager {
	return manager
}

// Start 启动集群管理器
func (m *Manager) Start() error {
	// 启动定时更新节点信息的goroutine
	m.startNodeInfoUpdater()
	return nil
}

// Stop 停止集群管理器
func (m *Manager) Stop() error {
	// 停止定时更新
	if m.updateCancel != nil {
		m.updateCancel()
		m.updateWg.Wait()
	}

	if m.registry != nil {
		return m.registry.Deregister()
	}
	return nil
}

// StartNode 启动节点并注册到节点发现管理器
// stage： 环境标识，如 dev、test、prod
// nodeID: 节点唯一标识
// nodeAddress: 节点地址，格式为 "IP:Port" 或 "Host:Port"
// serviceName: 服务名称，用于在 Nacos 中注册的服务名
// 返回集群管理器和错误
func StartNode(cfg *config.NacosConfig, stage, nodeID, nodeAddress string) error {
	// 1. 获取配置
	// 2. 创建 Nacos 注册发现实例
	registry, err := NewNacosRegistry(cfg, nodeID, nodeAddress, manager.serviceName)
	if err != nil {
		return fmt.Errorf("create nacos registry failed: %w", err)
	}

	// 3. 准备节点元数据
	metadata := map[string]string{
		"version": "1.0.0",
		"stage":   stage,
	}

	// 4. 注册服务节点
	if err := registry.Register(metadata); err != nil {
		return fmt.Errorf("register service failed: %w", err)
	}
	manager.registry = registry
	manager.NewNode(nodeID, nodeAddress)
	return nil
}

func (m *Manager) NewNode(nodeID, nodeAddress string) *Node {
	now := time.Now()
	currentNode := &Node{
		ID:        nodeID,
		Address:   nodeAddress,
		CreatedAt: now,
	}
	currentNode.SetState(NodeStateOnline)
	currentNode.SetLastHeartbeat(now)
	currentNode.SetUpdatedAt(now)
	m.currentNode = currentNode
	return currentNode
}

func (m *Manager) GetCurrentNode() *Node {
	return m.currentNode
}

// GetNodes 获取所有节点
// 直接从 NacosRegistry 获取，避免数据冗余
func (m *Manager) GetNodes() map[string]*Node {
	if m.registry == nil {
		return make(map[string]*Node)
	}
	return m.registry.GetNodes()
}

// GetNode 获取指定节点
// 直接从 NacosRegistry 获取，避免数据冗余
func (m *Manager) GetNode(nodeID string) (*Node, bool) {
	if m.registry == nil {
		return nil, false
	}
	return m.registry.GetNode(nodeID)
}

// GetOnlineNodes 获取所有在线节点
func (m *Manager) GetOnlineNodes() []*Node {
	if m.registry == nil {
		return nil
	}

	nodes := m.registry.GetNodes()
	var onlineNodes []*Node
	for _, node := range nodes {
		if node.GetState() == NodeStateOnline {
			onlineNodes = append(onlineNodes, node)
		}
	}
	return onlineNodes
}

// UpdateNodeMetadata 更新当前节点元数据
func (m *Manager) UpdateNodeMetadata(metadata map[string]string) error {
	if m.registry == nil {
		return fmt.Errorf("registry is nil")
	}
	return m.registry.UpdateNodeMetadata(metadata)
}

// SetUpdateInterval 设置节点信息更新间隔
func (m *Manager) SetUpdateInterval(interval time.Duration) {
	m.updateInterval = interval
}

// startNodeInfoUpdater 启动定时更新节点信息的goroutine
func (m *Manager) startNodeInfoUpdater() {
	m.updateWg.Add(1)
	go func() {
		defer m.updateWg.Done()
		ticker := time.NewTicker(m.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-m.updateCtx.Done():
				return
			case <-ticker.C:
				m.updateNodeInfoToNacos()
			}
		}
	}()
}

// updateNodeInfoToNacos 将当前节点的Node信息更新到Nacos
func (m *Manager) updateNodeInfoToNacos() {
	if m.currentNode == nil {
		return
	}
	node := m.currentNode

	// 构建metadata，包含节点的所有状态信息
	metadata := map[string]string{
		"state":        strconv.FormatInt(int64(node.GetState()), 10),
		"zone_count":   strconv.FormatInt(int64(node.GetZoneCount()), 10),
		"entity_count": strconv.FormatInt(node.GetEntityCount(), 10),
		"ccu":          strconv.FormatInt(int64(node.GetCCU()), 10),
	}

	// 添加最后心跳时间
	if lastHeartbeat := node.GetLastHeartbeat(); !lastHeartbeat.IsZero() {
		metadata["last_heartbeat"] = strconv.FormatInt(lastHeartbeat.Unix(), 10)
	}

	// 添加更新时间
	if updatedAt := node.GetUpdatedAt(); !updatedAt.IsZero() {
		metadata["updated_at"] = strconv.FormatInt(updatedAt.Unix(), 10)
	}

	// 更新到Nacos
	if err := m.UpdateNodeMetadata(metadata); err != nil {
		logger.GetLogger().Warn("update node info to nacos failed",
			zap.String("node_id", node.ID),
			zap.Error(err))
	} else {
		logger.GetLogger().Debug("update node info to nacos success",
			zap.String("node_id", node.ID),
			zap.Int32("zone_count", node.GetZoneCount()),
			zap.Int64("entity_count", node.GetEntityCount()),
			zap.Int32("ccu", node.GetCCU()))
	}
}
