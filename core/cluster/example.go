package cluster

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"gitee.com/orbit-w/orbit/app/modules/config"
	"gitee.com/orbit-w/orbit/lib/module/logger"
)

// ExampleUsage 使用示例
func ExampleUsage() {
	// 1. 加载配置
	config.LoadConfig("configs/config.toml")
	cfg := config.GetConfig()

	// 2. 获取节点信息
	nodeID := "node-001"
	serverCfg := cfg.Server
	nodeAddress := net.JoinHostPort(serverCfg.Host, serverCfg.Port)
	serviceName := "orbit-service"

	// 3. 创建 Nacos 注册发现实例
	nacosConfig := &cfg.Cluster.Nacos
	registry, err := NewNacosRegistry(nacosConfig, nodeID, nodeAddress, serviceName)
	if err != nil {
		logger.GetLogger().Error("create nacos registry failed", zap.Error(err))
		return
	}

	// 4. 注册服务节点
	metadata := map[string]string{
		"version": "1.0.0",
		"stage":   cfg.Server.Stage,
	}
	if err := registry.Register(metadata); err != nil {
		logger.GetLogger().Error("register service failed", zap.Error(err))
		return
	}

	// 5. 创建集群管理器
	manager := NewManager(registry)
	if err := manager.Start(); err != nil {
		logger.GetLogger().Error("start cluster manager failed", zap.Error(err))
		return
	}

	// 6. 获取节点列表
	nodes := manager.GetNodes()
	fmt.Printf("Total nodes: %d\n", len(nodes))
	for nodeID, node := range nodes {
		fmt.Printf("Node: %s, Address: %s, State: %d\n", nodeID, node.Address, node.State)
	}

	// 7. 更新节点元数据（例如：更新 CCU、EntityCount 等）
	updatedMetadata := map[string]string{
		"version":     "1.0.0",
		"stage":       cfg.Server.Stage,
		"ccu":         "100",
		"entity_count": "1000",
		"zone_count":   "10",
	}
	if err := manager.UpdateNodeMetadata(updatedMetadata); err != nil {
		logger.GetLogger().Error("update node metadata failed", zap.Error(err))
	}

	// 8. 程序退出时注销服务
	defer func() {
		if err := manager.Stop(); err != nil {
			logger.GetLogger().Error("stop cluster manager failed", zap.Error(err))
		}
	}()
}

