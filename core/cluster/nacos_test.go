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
		cfg := Setup("./config_center.yaml")

		ip, err := netutils.GetLocalIPv4()
		assert.NoError(t, err)

		_, err = NewNacosRegistry(cfg.Nacos, "test-node-001", fmt.Sprintf("%s:9000", ip), "test-service")
		// 如果失败，应该是因为无法连接 Nacos，而不是地址解析错误
		assert.NoError(t, err)

	})
}
