package cluster

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleUsage 使用示例
func ExampleUsage() {
	// 1. 创建服务发现配置
	config := DiscoveryConfig{
		Endpoints:         []string{"127.0.0.1:2379"}, // etcd 端点
		KeyPrefix:         "/orbit/nodes/",            // key 前缀
		LeaseTTL:          10,                         // 租约 TTL (秒)
		HeartbeatInterval: 3 * time.Second,            // 心跳间隔
		DialTimeout:       5 * time.Second,            // 连接超时
	}

	// 2. 创建服务发现管理器
	discovery, err := NewDiscovery(config)
	if err != nil {
		log.Fatalf("Failed to create discovery: %v", err)
	}
	defer discovery.Close()

	// 3. 创建并注册节点
	node := &Node{
		ID:      "node-001",
		Address: "192.168.1.100:8080",
		State:   NodeStateOnline,
	}

	err = discovery.Register(node)
	if err != nil {
		log.Fatalf("Failed to register node: %v", err)
	}
	fmt.Println("Node registered successfully")

	// 4. 更新节点信息
	node.AddZoneCount(5)
	node.AddEntityCount(100)
	node.AddCCU(50)
	err = discovery.UpdateNode(node)
	if err != nil {
		log.Printf("Failed to update node: %v", err)
	}

	// 5. 获取所有在线节点列表
	nodes, err := discovery.GetNodes()
	if err != nil {
		log.Printf("Failed to get nodes: %v", err)
	} else {
		fmt.Printf("Total online nodes: %d\n", len(nodes))
		for _, n := range nodes {
			fmt.Printf("  - Node ID: %s, Address: %s, CCU: %d\n",
				n.ID, n.Address, n.GetCCU())
		}
	}

	// 6. 从缓存获取在线节点（更快）
	onlineNodes := discovery.GetOnlineNodes()
	fmt.Printf("Online nodes from cache: %d\n", len(onlineNodes))

	// 7. 获取指定节点信息
	targetNode, err := discovery.GetNode("node-001")
	if err != nil {
		log.Printf("Failed to get node: %v", err)
	} else {
		fmt.Printf("Node info: ID=%s, Address=%s, ZoneCount=%d, EntityCount=%d, CCU=%d\n",
			targetNode.ID, targetNode.Address,
			targetNode.GetZoneCount(), targetNode.GetEntityCount(), targetNode.GetCCU())
	}

	// 8. 获取节点数量
	count := discovery.GetNodeCount()
	fmt.Printf("Total node count: %d\n", count)

	// 9. 运行一段时间后注销
	time.Sleep(10 * time.Second)
	err = discovery.Unregister()
	if err != nil {
		log.Printf("Failed to unregister: %v", err)
	}
}

// ExampleWithContext 使用 context 的示例
func ExampleWithContext(ctx context.Context) {
	config := DiscoveryConfig{
		Endpoints: []string{"127.0.0.1:2379"},
	}

	discovery, err := NewDiscovery(config)
	if err != nil {
		log.Fatalf("Failed to create discovery: %v", err)
	}
	defer discovery.Close()

	node := &Node{
		ID:      "node-002",
		Address: "192.168.1.101:8080",
		State:   NodeStateOnline,
	}

	err = discovery.Register(node)
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	// 定期获取节点列表
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nodes := discovery.GetOnlineNodes()
			fmt.Printf("Current online nodes: %d\n", len(nodes))
		}
	}
}
