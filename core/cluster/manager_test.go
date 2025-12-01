package cluster

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gitee.com/orbit-w/orbit/app/modules/config"
	netutils "gitee.com/orbit-w/orbit/lib/utils/net_utils"
	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Nacos config.NacosConfig `toml:"nacos"`
}

func Setup(filename string) config.NacosConfig {
	// 读取测试配置
	viper.SetConfigFile(filename)
	viper.SetConfigType("toml")

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic("viper read config failed")
	}

	// 读取配置文件
	content, err := os.ReadFile(filename)
	if err != nil {
		panic("read config failed")
	}

	cfg := new(TestConfig)
	if err := toml.Unmarshal(content, &cfg); err != nil {
		panic("unmarshal config failed")
	}
	return cfg.Nacos
}

func TestManager_Start(t *testing.T) {
	t.Run("启动管理器", func(t *testing.T) {
		manager := NewManager("test-service")
		err := manager.Start()

		assert.NoError(t, err)
		// 清理
		_ = manager.Stop()
	})

	t.Run("多次启动", func(t *testing.T) {
		manager := NewManager("test-service")
		err1 := manager.Start()
		err2 := manager.Start()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		// 清理
		_ = manager.Stop()
	})
}

func TestManager_Stop(t *testing.T) {
	t.Run("停止未启动的管理器", func(t *testing.T) {
		manager := NewManager("test-service")
		err := manager.Stop()

		assert.NoError(t, err)
	})

	t.Run("停止已启动的管理器", func(t *testing.T) {
		manager := NewManager("test-service")
		_ = manager.Start()

		// 等待一小段时间确保 goroutine 启动
		time.Sleep(50 * time.Millisecond)

		err := manager.Stop()
		assert.NoError(t, err)
	})

	t.Run("多次停止", func(t *testing.T) {
		manager := NewManager("test-service")
		_ = manager.Start()
		time.Sleep(50 * time.Millisecond)

		err1 := manager.Stop()
		err2 := manager.Stop()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})
}

func TestManager_NewNode(t *testing.T) {
	t.Run("创建新节点", func(t *testing.T) {
		manager := NewManager("test-service")

		manager.Start()
		defer manager.Stop()

		nodeID := "game-node-001"
		ip, err := netutils.GetLocalIPv4()
		assert.NoError(t, err)
		nodeAddress := fmt.Sprintf("%s:9000", ip)
		cfg := Setup("./nacos.toml")
		err = StartNode(cfg, "dev", nodeID, nodeAddress)
		assert.NoError(t, err)
		node := manager.GetCurrentNode()

		assert.NotNil(t, node)
		assert.Equal(t, nodeID, node.ID)
		assert.Equal(t, nodeAddress, node.Address)
		assert.Equal(t, NodeStateOnline, node.GetState())
		assert.False(t, node.GetLastHeartbeat().IsZero())
		assert.False(t, node.GetUpdatedAt().IsZero())
		assert.Equal(t, node, manager.currentNode)
		time.Sleep(time.Second)
		fmt.Println("manager.nodes", manager.GetNodes())
		time.Sleep(30 * time.Second)
	})
}

func TestManager_GetCurrentNode(t *testing.T) {
	t.Run("获取当前节点-未设置", func(t *testing.T) {
		manager := NewManager("test-service")
		node := manager.GetCurrentNode()

		assert.Nil(t, node)
	})

	t.Run("获取当前节点-已设置", func(t *testing.T) {
		manager := NewManager("test-service")
		createdNode := manager.NewNode("node-1", "127.0.0.1:8080")
		retrievedNode := manager.GetCurrentNode()

		assert.NotNil(t, retrievedNode)
		assert.Equal(t, createdNode, retrievedNode)
	})
}

func TestManager_GetNodes(t *testing.T) {
	t.Run("获取节点列表-无注册表", func(t *testing.T) {
		manager := NewManager("test-service")
		nodes := manager.GetNodes()

		assert.NotNil(t, nodes)
		assert.Equal(t, 0, len(nodes))
	})

	t.Run("获取节点列表-有注册表", func(t *testing.T) {
		// 注意：这个测试需要实际的 NacosRegistry，或者需要 mock
		// 这里只测试基本逻辑
		manager := NewManager("test-service")
		nodes := manager.GetNodes()

		// 即使没有注册表，也应该返回空 map 而不是 nil
		assert.NotNil(t, nodes)
	})
}

func TestManager_GetNode(t *testing.T) {
	t.Run("获取指定节点-无注册表", func(t *testing.T) {
		manager := NewManager("test-service")
		node, ok := manager.GetNode("node-1")

		assert.Nil(t, node)
		assert.False(t, ok)
	})
}

func TestManager_GetOnlineNodes(t *testing.T) {
	t.Run("获取在线节点-无注册表", func(t *testing.T) {
		manager := NewManager("test-service")
		nodes := manager.GetOnlineNodes()

		assert.Nil(t, nodes)
	})
}

func TestManager_UpdateNodeMetadata(t *testing.T) {
	t.Run("更新节点元数据-无注册表", func(t *testing.T) {
		manager := NewManager("test-service")
		metadata := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}

		err := manager.UpdateNodeMetadata(metadata)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "registry is nil")
	})
}

func TestManager_SetUpdateInterval(t *testing.T) {
	t.Run("设置更新间隔", func(t *testing.T) {
		manager := NewManager("test-service")
		interval := 30 * time.Second

		manager.SetUpdateInterval(interval)
		assert.Equal(t, interval, manager.updateInterval)
	})

	t.Run("设置不同的更新间隔", func(t *testing.T) {
		manager := NewManager("test-service")

		manager.SetUpdateInterval(10 * time.Second)
		assert.Equal(t, 10*time.Second, manager.updateInterval)

		manager.SetUpdateInterval(60 * time.Second)
		assert.Equal(t, 60*time.Second, manager.updateInterval)
	})
}

func TestManager_StartStopCycle(t *testing.T) {
	t.Run("启动停止循环", func(t *testing.T) {
		manager := NewManager("test-service")

		// 启动
		err := manager.Start()
		assert.NoError(t, err)

		// 等待一小段时间
		time.Sleep(100 * time.Millisecond)

		// 停止
		err = manager.Stop()
		assert.NoError(t, err)

		// 再次启动和停止
		err = manager.Start()
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		err = manager.Stop()
		assert.NoError(t, err)
	})
}

func TestManager_UpdateInterval(t *testing.T) {
	t.Run("更新间隔影响定时器", func(t *testing.T) {
		manager := NewManager("test-service")
		manager.SetUpdateInterval(100 * time.Millisecond)

		// 创建节点
		node := manager.NewNode("node-1", "127.0.0.1:8080")
		assert.NotNil(t, node)

		// 启动管理器
		err := manager.Start()
		assert.NoError(t, err)

		// 等待一段时间，确保定时器触发
		time.Sleep(150 * time.Millisecond)

		// 停止
		err = manager.Stop()
		assert.NoError(t, err)
	})
}
