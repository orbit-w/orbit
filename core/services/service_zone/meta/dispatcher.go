package zone_meta

import (
	"errors"
	"fmt"
	"math/rand"

	"gitee.com/orbit-w/orbit/core/cluster"
)

var (
	ErrNoAvailableNode   = errors.New("no available nodes")
	ErrNodeNotFound      = errors.New("designated node not found")
	ErrInvalidDispatcher = errors.New("invalid dispatcher configuration")
)

// ZoneDispatcherService 调度器服务，负责根据调度策略选择节点
type ZoneDispatcherService struct {
	clusterMgr *cluster.Manager
}

// NewZoneDispatcherService 创建调度器服务
func NewZoneDispatcherService() *ZoneDispatcherService {
	return &ZoneDispatcherService{
		clusterMgr: cluster.GetManager(),
	}
}

// SelectNode 根据调度策略选择节点
// 返回选中的节点和错误信息
func (s *ZoneDispatcherService) SelectNode(dispatcher *ZoneDispatcher) (*cluster.Node, error) {
	if dispatcher == nil {
		return nil, ErrInvalidDispatcher
	}

	switch dispatcher.Type {
	case Zone_DispatcherType_ForDesignated:
		return s.selectDesignatedNode(dispatcher)
	case Zone_DispatcherType_ForLeastLoaded:
		return s.selectLeastLoadedNode()
	case Zone_DispatcherType_ForRandom:
		return s.selectRandomNode(dispatcher)
	default:
		return nil, fmt.Errorf("unknown dispatcher type: %v", dispatcher.Type)
	}
}

// selectDesignatedNode 选择指定节点
// ForDesignated 策略：严格匹配 NodeId，如果节点不存在或不在线则返回错误
func (s *ZoneDispatcherService) selectDesignatedNode(dispatcher *ZoneDispatcher) (*cluster.Node, error) {
	if dispatcher.NodeId == "" {
		return nil, fmt.Errorf("%w: NodeId is required for ForDesignated type", ErrInvalidDispatcher)
	}

	node, exists := s.clusterMgr.GetNode(dispatcher.NodeId)
	if !exists {
		return nil, fmt.Errorf("%w: node %s", ErrNodeNotFound, dispatcher.NodeId)
	}

	if node.GetState() != cluster.NodeStateOnline {
		return nil, fmt.Errorf("node %s is not online (state: %v)", dispatcher.NodeId, node.GetState())
	}

	return node, nil
}

// selectRandomNode 随机选择节点
// ForRandom 策略：支持两种模式
// 1. 纯随机：从所有在线节点中随机选择
// 2. 负载均衡：根据节点负载（CCU、EntityCount、ZoneCount）选择负载最低的节点
func (s *ZoneDispatcherService) selectRandomNode(_ *ZoneDispatcher) (*cluster.Node, error) {
	onlineNodes := s.clusterMgr.GetOnlineNodes()
	if len(onlineNodes) == 0 {
		return nil, ErrNoAvailableNode
	}

	// 如果只有一个节点，直接返回
	if len(onlineNodes) == 1 {
		return onlineNodes[0], nil
	}

	// 默认使用负载均衡策略
	return s.selectRandomNodeFrom(onlineNodes), nil
}

// selectLeastLoadedNode 从所有在线节点中选择负载最低的节点
func (s *ZoneDispatcherService) selectLeastLoadedNode() (*cluster.Node, error) {
	onlineNodes := s.clusterMgr.GetOnlineNodes()
	if len(onlineNodes) == 0 {
		return nil, ErrNoAvailableNode
	}

	return s.selectLeastLoadedNodeFrom(onlineNodes), nil
}

// selectLeastLoadedNodeFrom 从给定节点列表中选择负载最低的节点
// 负载计算方式：考虑 ZoneCount、EntityCount 和 CCU 三个维度
func (s *ZoneDispatcherService) selectLeastLoadedNodeFrom(nodes []*cluster.Node) *cluster.Node {
	if len(nodes) == 0 {
		return nil
	}

	if len(nodes) == 1 {
		return nodes[0]
	}

	// 选择负载最低的节点
	// 负载计算：ZoneCount * 1000 + EntityCount / 100 + CCU
	// 这个公式优先考虑 ZoneCount，其次是 EntityCount，最后是 CCU
	minLoad := int64(-1)
	var selectedNode *cluster.Node

	for _, node := range nodes {
		load := calculateNodeLoad(node)
		if minLoad == -1 || load < minLoad {
			minLoad = load
			selectedNode = node
		}
	}

	return selectedNode
}

// selectRandomNodeFrom 从给定节点列表中随机选择一个节点（纯随机，不考虑负载）
func (s *ZoneDispatcherService) selectRandomNodeFrom(nodes []*cluster.Node) *cluster.Node {
	if len(nodes) == 0 {
		return nil
	}

	if len(nodes) == 1 {
		return nodes[0]
	}

	return nodes[rand.Intn(len(nodes))]
}

// calculateNodeLoad 计算节点负载
// 负载 = ZoneCount * 1000 + EntityCount / 100 + CCU
func calculateNodeLoad(node *cluster.Node) int64 {
	zoneCount := int64(node.ZoneCount.Load())
	entityCount := node.EntityCount.Load()
	ccu := int64(node.CCU.Load())

	return zoneCount*1000 + entityCount/100 + ccu
}

// DispatchZone 为 Zone 分配节点（便捷方法）
// 该方法会根据 dispatcher 的配置选择节点，并更新 dispatcher 的 NodeId
func (s *ZoneDispatcherService) DispatchZone(zoneId string, dispatcher *ZoneDispatcher) (*cluster.Node, error) {
	node, err := s.SelectNode(dispatcher)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch zone %s: %w", zoneId, err)
	}

	// 更新 dispatcher 的 NodeId
	dispatcher.NodeId = node.ID

	return node, nil
}
