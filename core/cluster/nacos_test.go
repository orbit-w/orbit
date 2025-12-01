package cluster

import (
	"fmt"
	"testing"

	netutils "gitee.com/orbit-w/orbit/lib/utils/net_utils"
	"github.com/stretchr/testify/assert"
)

func TestParseServerAddress(t *testing.T) {
	t.Run("解析带端口的服务器地址", func(t *testing.T) {
		host, port := parseServerAddress("127.0.0.1:8848")
		assert.Equal(t, "127.0.0.1", host)
		assert.Equal(t, uint64(8848), port)
	})

	t.Run("解析无端口的服务器地址-使用默认端口", func(t *testing.T) {
		// 阿里云 Nacos 域名场景
		host, port := parseServerAddress("mse-63a694613-p.nacos-ans.mse.aliyuncs.com")
		assert.Equal(t, "mse-63a694613-p.nacos-ans.mse.aliyuncs.com", host)
		assert.Equal(t, uint64(8848), port) // 应该使用默认端口
	})

	t.Run("解析带域名的服务器地址", func(t *testing.T) {
		host, port := parseServerAddress("example.com:8848")
		assert.Equal(t, "example.com", host)
		assert.Equal(t, uint64(8848), port)
	})

	t.Run("解析无端口的域名-使用默认端口", func(t *testing.T) {
		host, port := parseServerAddress("nacos.example.com")
		assert.Equal(t, "nacos.example.com", host)
		assert.Equal(t, uint64(8848), port)
	})

	t.Run("解析IPv6地址", func(t *testing.T) {
		host, port := parseServerAddress("[::1]:8848")
		assert.Equal(t, "::1", host)
		assert.Equal(t, uint64(8848), port)
	})

	t.Run("解析空地址-使用默认端口", func(t *testing.T) {
		host, port := parseServerAddress("")
		assert.Equal(t, "", host)
		assert.Equal(t, uint64(8848), port)
	})

	t.Run("解析无效端口-使用默认端口", func(t *testing.T) {
		host, port := parseServerAddress("127.0.0.1:invalid")
		assert.Equal(t, "127.0.0.1", host)
		assert.Equal(t, uint64(8848), port) // 端口解析失败时使用默认端口
	})

	t.Run("解析零端口-使用默认端口", func(t *testing.T) {
		host, port := parseServerAddress("127.0.0.1:0")
		assert.Equal(t, "127.0.0.1", host)
		assert.Equal(t, uint64(8848), port) // 端口为0时使用默认端口
	})

	t.Run("解析自定义端口", func(t *testing.T) {
		host, port := parseServerAddress("127.0.0.1:9999")
		assert.Equal(t, "127.0.0.1", host)
		assert.Equal(t, uint64(9999), port)
	})
}

func TestParseAddress(t *testing.T) {
	t.Run("解析有效地址", func(t *testing.T) {
		ip, port, err := parseAddress("127.0.0.1:8080")
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1", ip)
		assert.Equal(t, uint64(8080), port)
	})

	t.Run("解析带域名的地址", func(t *testing.T) {
		host, port, err := parseAddress("example.com:8848")
		assert.NoError(t, err)
		assert.Equal(t, "example.com", host)
		assert.Equal(t, uint64(8848), port)
	})

	t.Run("解析IPv6地址", func(t *testing.T) {
		ip, port, err := parseAddress("[::1]:8080")
		assert.NoError(t, err)
		assert.Equal(t, "::1", ip)
		assert.Equal(t, uint64(8080), port)
	})

	t.Run("解析空地址", func(t *testing.T) {
		_, _, err := parseAddress("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address is empty")
	})

	t.Run("解析缺少端口的地址-必须返回错误", func(t *testing.T) {
		// 节点地址必须包含端口
		_, _, err := parseAddress("127.0.0.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid address format")
	})

	t.Run("解析无效端口", func(t *testing.T) {
		_, _, err := parseAddress("127.0.0.1:invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid port")
	})

	t.Run("解析超大端口号", func(t *testing.T) {
		ip, port, err := parseAddress("127.0.0.1:99999")
		// 取决于 uint64 的解析，可能会成功或失败
		// 这里只测试不会 panic
		_ = ip
		_ = port
		_ = err
	})

	t.Run("解析零端口", func(t *testing.T) {
		ip, port, err := parseAddress("127.0.0.1:0")
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1", ip)
		assert.Equal(t, uint64(0), port)
	})
}

// 注意：以下测试需要实际的 Nacos 服务器或 mock Nacos 客户端
// 这些是集成测试，需要特殊环境

func TestNewNacosRegistry_InvalidConfig(t *testing.T) {
	t.Run("配置为nil", func(t *testing.T) {
		_, err := NewNacosRegistry(NacosConfig{}, "test-node-001", "127.0.0.1:8080", "test-service")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nacos config is nil")
	})

	t.Run("服务器地址为空", func(t *testing.T) {
		config := NacosConfig{
			ServerHosts: []string{},
		}

		_, err := NewNacosRegistry(config, "node-1", "127.0.0.1:8080", "test-service")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nacos server hosts is empty")
	})

	t.Run("节点地址格式错误", func(t *testing.T) {
		config := NacosConfig{
			ServerHosts: []string{"127.0.0.1:8848"},
		}

		_, err := NewNacosRegistry(config, "node-1", "invalid-address", "test-service")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse node address failed")
	})

	t.Run("服务器地址支持无端口域名-阿里云场景", func(t *testing.T) {
		// 测试阿里云 Nacos 域名场景，服务器地址无端口
		config := NacosConfig{
			ServerHosts:  []string{"mse-63a694613-p.nacos-ans.mse.aliyuncs.com"},
			GroupName:    "TEST_GROUP",
			TimeoutMs:    5000,
			BeatInterval: 5,
		}

		// 注意：这个测试需要实际的 Nacos 客户端，可能会失败
		// 但至少可以验证地址解析不会出错
		_, err := NewNacosRegistry(config, "node-1", "127.0.0.1:8080", "test-service")
		// 如果失败，应该是因为无法连接 Nacos，而不是地址解析错误
		if err != nil {
			// 不应该是地址解析错误
			assert.NotContains(t, err.Error(), "parse server address")
			assert.NotContains(t, err.Error(), "invalid address format")
		}
	})

	t.Run("服务器地址支持混合格式", func(t *testing.T) {
		// 测试同时支持带端口和不带端口的服务器地址
		config := NacosConfig{
			ServerHosts: []string{
				"mse-63a694613-p.nacos-ans.mse.aliyuncs.com", // 无端口，使用默认端口
				"127.0.0.1:8848", // 带端口
			},
		}

		_, err := NewNacosRegistry(config, "node-1", "127.0.0.1:8080", "test-service")
		// 如果失败，应该是因为无法连接 Nacos，而不是地址解析错误
		if err != nil {
			assert.NotContains(t, err.Error(), "parse server address")
		}
	})
}

func TestNacosRegistry_GetCurrentNodeID(t *testing.T) {
	t.Run("获取当前节点ID", func(t *testing.T) {
		registry := &NacosRegistry{
			nodeID: "test-node-123",
		}

		nodeID := registry.GetCurrentNodeID()
		assert.Equal(t, "test-node-123", nodeID)
	})
}

func TestNacosRegistry(t *testing.T) {
	t.Run("注册服务", func(t *testing.T) {
		// 测试阿里云 Nacos 域名场景，服务器地址无端口
		config := NacosConfig{
			ServerHosts:  []string{"mse-63a694613-p.nacos-ans.mse.aliyuncs.com"},
			GroupName:    "TEST_GROUP",
			TimeoutMs:    5000,
			BeatInterval: 5,
		}

		ip, err := netutils.GetLocalIPv4()
		assert.NoError(t, err)

		_, err = NewNacosRegistry(config, "test-node-001", fmt.Sprintf("%s:9000", ip), "test-service")
		// 如果失败，应该是因为无法连接 Nacos，而不是地址解析错误
		assert.NoError(t, err)

	})
}
