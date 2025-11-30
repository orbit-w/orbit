package cluster

import (
	"sync/atomic"
	"time"
)

// NodeState 节点状态
type NodeState int32

const (
	NodeStateUnknown     NodeState = iota // 未知状态
	NodeStateOnline                       // 在线
	NodeStateOffline                      // 离线
	NodeStateMaintenance                  // 维护中
)

// Node 节点信息
type Node struct {
	ID            string       // 节点ID
	Address       string       // 节点地址 (IP:Port)
	state         atomic.Int32 // 节点状态（使用原子操作保证并发安全）
	ZoneCount     atomic.Int32 // 当前管理的服务区数量
	EntityCount   atomic.Int64 // 当前管理的实体数量
	CCU           atomic.Int32 // 当前连接的客户端数量
	LastHeartbeat atomic.Value // 最后心跳时间（使用 atomic.Value 保证并发安全）
	CreatedAt     time.Time    // 创建时间（只读，创建后不再修改）
	UpdatedAt     atomic.Value // 更新时间（使用 atomic.Value 保证并发安全）
}

// SetState 设置节点状态
func (n *Node) SetState(state NodeState) {
	n.state.Store(int32(state))
}

// GetState 获取节点状态
func (n *Node) GetState() NodeState {
	return NodeState(n.state.Load())
}

// SetLastHeartbeat 设置最后心跳时间
func (n *Node) SetLastHeartbeat(t time.Time) {
	n.LastHeartbeat.Store(t)
}

// GetLastHeartbeat 获取最后心跳时间
func (n *Node) GetLastHeartbeat() time.Time {
	if v := n.LastHeartbeat.Load(); v != nil {
		return v.(time.Time)
	}
	return time.Time{}
}

// SetUpdatedAt 设置更新时间
func (n *Node) SetUpdatedAt(t time.Time) {
	n.UpdatedAt.Store(t)
}

// GetUpdatedAt 获取更新时间
func (n *Node) GetUpdatedAt() time.Time {
	if v := n.UpdatedAt.Load(); v != nil {
		return v.(time.Time)
	}
	return time.Time{}
}

// AddCCU 增加CCU
func (n *Node) AddCCU(ccu int32) {
	n.CCU.Add(ccu)
}

// SubCCU 减少CCU
func (n *Node) SubCCU(ccu int32) {
	n.CCU.Add(-ccu)
}

// GetCCU 获取CCU
func (n *Node) GetCCU() int32 {
	return n.CCU.Load()
}

// AddEntityCount 增加实体数量
func (n *Node) AddEntityCount(count int64) {
	n.EntityCount.Add(count)
}

// SubEntityCount 减少实体数量
func (n *Node) SubEntityCount(count int64) {
	n.EntityCount.Add(-count)
}

// IncrEntityCount 实体数量加1
func (n *Node) IncrEntityCount() {
	n.EntityCount.Add(1)
}

// DecrEntityCount 实体数量减1
func (n *Node) DecrEntityCount() {
	n.EntityCount.Add(-1)
}

// GetEntityCount 获取实体数量
func (n *Node) GetEntityCount() int64 {
	return n.EntityCount.Load()
}

// AddZoneCount 增加服务区数量
func (n *Node) AddZoneCount(count int32) {
	n.ZoneCount.Add(count)
}

// SubZoneCount 减少服务区数量
func (n *Node) SubZoneCount(count int32) {
	n.ZoneCount.Add(-count)
}

// GetZoneCount 获取服务区数量
func (n *Node) GetZoneCount() int32 {
	return n.ZoneCount.Load()
}
