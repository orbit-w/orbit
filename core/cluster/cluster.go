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
	State         NodeState    // 节点状态
	ZoneCount     atomic.Int32 // 当前管理的服务区数量
	EntityCount   atomic.Int64 // 当前管理的实体数量
	CCU           atomic.Int32 // 当前连接的客户端数量
	LastHeartbeat time.Time    // 最后心跳时间
	CreatedAt     time.Time    // 创建时间
	UpdatedAt     time.Time    // 更新时间
}

func (n *Node) AddCCU(ccu int32) {
	n.CCU.Add(int32(ccu))
}

func (n *Node) SubCCU(ccu int32) {
	n.CCU.Add(-int32(ccu))
}

func (n *Node) GetCCU() int32 {
	return n.CCU.Load()
}

func (n *Node) AddEntityCount(count int64) {
	n.EntityCount.Add(count)
}

func (n *Node) SubEntityCount(count int64) {
	n.EntityCount.Add(-count)
}

func (n *Node) IncrEntityCount() {
	n.EntityCount.Add(1)
}

func (n *Node) DecrEntityCount() {
	n.EntityCount.Add(-1)
}

func (n *Node) GetEntityCount() int64 {
	return n.EntityCount.Load()
}

func (n *Node) AddZoneCount(count int32) {
	n.ZoneCount.Add(count)
}

func (n *Node) SubZoneCount(count int32) {
	n.ZoneCount.Add(-count)
}

func (n *Node) GetZoneCount() int32 {
	return n.ZoneCount.Load()
}
