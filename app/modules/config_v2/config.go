package config

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"

	"gitee.com/orbit-w/orbit/lib/module/logger"
)

var (
	cfg          Config
	configClient config_client.IConfigClient
	mu           sync.RWMutex
)

type Config struct {
	Server  Server  `toml:"server"`
	Cluster Cluster `toml:"cluster"`
	Redis   Redis   `toml:"redis"`
}

type Cluster struct {
	Nacos NacosConfig `toml:"nacos"`
}

func (c *Config) GetNacosConfig() NacosConfig {
	return c.Cluster.Nacos
}

// NacosConfig Nacos 配置
type NacosConfig struct {
	ServerHosts  []string `toml:"server_hosts"`  // Nacos 服务器地址列表，格式: "127.0.0.1:8848"
	NamespaceID  string   `toml:"namespace_id"`  // 命名空间ID
	GroupName    string   `toml:"group_name"`    // 服务组名，默认: "DEFAULT_GROUP"
	Username     string   `toml:"username"`      // 用户名（可选）
	Password     string   `toml:"password"`      // 密码（可选）
	TimeoutMs    uint64   `toml:"timeout_ms"`    // 超时时间（毫秒），默认: 5000
	BeatInterval int64    `toml:"beat_interval"` // 心跳间隔（秒），默认: 5
	// 配置中心相关参数
	DataId string `toml:"data_id"` // 配置的 DataId，用于从 Nacos 配置中心读取配置
}

type Server struct {
	Name  string `toml:"name"`  // 服务名称
	Stage string `toml:"stage"` // 环境
	Host  string `toml:"host"`  // 主机地址
	Port  string `toml:"port"`  // 端口
}

func GetConfig() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return &cfg
}

func GenDataId(serviceName, stage, packageName string) string {
	return fmt.Sprintf("%s-%s-%s.yaml", serviceName, stage, packageName)
}

func LoadConfigFromNacos(nacosConfig NacosConfig, dataId, group string) error {
	// 设置默认值
	if nacosConfig.GroupName == "" {
		nacosConfig.GroupName = "DEFAULT_GROUP"
	}
	if group == "" {
		group = nacosConfig.GroupName
	}
	if group == "" {
		group = constant.DEFAULT_GROUP
	}
	if nacosConfig.DataId == "" {
		nacosConfig.DataId = dataId
	}
	if nacosConfig.DataId == "" {
		return fmt.Errorf("dataId is required")
	}
	if nacosConfig.TimeoutMs == 0 {
		nacosConfig.TimeoutMs = 5000
	}

	// 构建服务器配置
	if len(nacosConfig.ServerHosts) == 0 {
		return fmt.Errorf("nacos server hosts is empty")
	}

	serverConfigs := make([]constant.ServerConfig, 0, len(nacosConfig.ServerHosts))
	for _, host := range nacosConfig.ServerHosts {
		serverHost, serverPort := parseServerAddress(host)
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      serverHost,
			Port:        serverPort,
			ContextPath: "/nacos",
		})
	}

	// 构建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         nacosConfig.NamespaceID,
		TimeoutMs:           nacosConfig.TimeoutMs,
		NotLoadCacheAtStart: true,
		LogLevel:            "info",
	}

	// 如果有用户名密码，设置认证
	if nacosConfig.Username != "" && nacosConfig.Password != "" {
		clientConfig.Username = nacosConfig.Username
		clientConfig.Password = nacosConfig.Password
	}

	// 创建 Nacos 配置客户端
	var err error
	configClient, err = clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return fmt.Errorf("create nacos config client failed: %w", err)
	}

	// 从 Nacos 读取配置
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: nacosConfig.DataId,
		Group:  group,
	})
	if err != nil {
		return fmt.Errorf("get config from nacos failed: %w", err)
	}

	if content == "" {
		return fmt.Errorf("config content is empty from nacos")
	}

	// 解析 TOML 配置
	mu.Lock()
	if err := toml.Unmarshal([]byte(content), &cfg); err != nil {
		mu.Unlock()
		return fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 保存 Nacos 配置到 cfg 中
	cfg.Cluster.Nacos = nacosConfig
	mu.Unlock()

	// 监听配置变化
	if err := configClient.ListenConfig(vo.ConfigParam{
		DataId: nacosConfig.DataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			// 配置变化时的回调
			mu.Lock()
			defer mu.Unlock()
			if err := toml.Unmarshal([]byte(data), &cfg); err != nil {
				// 记录错误但不中断程序
				logger.GetLogger().Error("failed to update config from nacos",
					zap.String("namespace", namespace),
					zap.String("group", group),
					zap.String("dataId", dataId),
					zap.Error(err))
				return
			}
			// 更新 Nacos 配置（保持连接信息）
			cfg.Cluster.Nacos = nacosConfig
			logger.GetLogger().Info("config updated from nacos",
				zap.String("namespace", namespace),
				zap.String("group", group),
				zap.String("dataId", dataId))
		},
	}); err != nil {
		// 监听失败不影响启动，只记录错误
		logger.GetLogger().Warn("failed to listen config changes",
			zap.String("dataId", nacosConfig.DataId),
			zap.String("group", group),
			zap.Error(err))
	}

	return nil
}

// parseServerAddress 解析服务器地址（支持域名，无端口时使用默认端口 8848）
func parseServerAddress(address string) (string, uint64) {
	if address == "" {
		return "", 8848 // 默认端口
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		// 如果没有端口，使用默认端口 8848
		return address, 8848
	}

	if host == "" {
		// 如果 host 为空，使用整个地址作为 host
		return address, 8848
	}

	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return host, 8848
	}

	if port == 0 {
		port = 8848 // 确保端口不为 0
	}

	return host, port
}

// CloseConfigClient 关闭配置客户端
func CloseConfigClient() {
	if configClient != nil {
		configClient.CloseClient()
	}
}
