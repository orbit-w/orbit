package cluster

import (
	"fmt"
	"sync"
)

// Manager 集群管理器
type Manager struct {
	registry *NacosRegistry
	mu       sync.RWMutex
	nodes    map[string]*Node
}

// NewManager 创建集群管理器
func NewManager(registry *NacosRegistry) *Manager {
	return &Manager{
		registry: registry,
		nodes:    make(map[string]*Node),
	}
}

// Start 启动集群管理器
func (m *Manager) Start() error {
	if m.registry == nil {
		return fmt.Errorf("registry is nil")
	}

	// 同步节点列表
	m.syncNodes()

	return nil
}

// Stop 停止集群管理器
func (m *Manager) Stop() error {
	if m.registry != nil {
		return m.registry.Deregister()
	}
	return nil
}

// syncNodes 同步节点列表
func (m *Manager) syncNodes() {
	if m.registry == nil {
		return
	}

	registryNodes := m.registry.GetNodes()
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nodes = registryNodes
}

// GetNodes 获取所有节点
func (m *Manager) GetNodes() map[string]*Node {
	m.syncNodes()
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodes := make(map[string]*Node)
	for k, v := range m.nodes {
		nodes[k] = v
	}
	return nodes
}

// GetNode 获取指定节点
func (m *Manager) GetNode(nodeID string) (*Node, bool) {
	m.syncNodes()
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, ok := m.nodes[nodeID]
	return node, ok
}

// GetOnlineNodes 获取所有在线节点
func (m *Manager) GetOnlineNodes() []*Node {
	m.syncNodes()
	m.mu.RLock()
	defer m.mu.RUnlock()

	var onlineNodes []*Node
	for _, node := range m.nodes {
		if node.State == NodeStateOnline {
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

