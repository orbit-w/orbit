package config

import (
	"fmt"
	"sync/atomic"

	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

const (
	DefaultTimeoutMs uint64 = 5000 //默认5秒超时

	StateNormal = iota
	StateStopping
	StateStopped
)

type ConfigManager struct {
	centerConfig CenterConfig                // 配置中心基础配置，用于获取主配置
	cfg          *Config                     // 配置
	configClient config_client.IConfigClient // Nacos 配置客户端
	state        atomic.Int32                // 状态
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		centerConfig: centerConfig,
	}
}

// 启动配置管理器, 如果启动失败，则panic
func (m *ConfigManager) Start(filename string) {
	// 加载配置中心配置
	m.LoadConfig(filename)
	// 初始化Nacos客户端
	if err := m.InitNacosClient(); err != nil {
		panic(err)
	}
	// 初始化配置
	if err := m.InitConfig(); err != nil {
		panic(err)
	}
}

func (m *ConfigManager) Stop() error {
	if m.state.CompareAndSwap(StateNormal, StateStopping) {
		return fmt.Errorf("config manager is already stopping")
	}
	if m.configClient != nil {
		m.configClient.CloseClient()
	}
	m.state.Store(StateStopped)
	return nil
}

func (m *ConfigManager) InitConfig() error {
	m.cfg = NewConfig()

	redisItem := &GameMainRedis{}
	if err := m.LoadConfigItme(redisItem); err != nil {
		return fmt.Errorf("listen redis config failed: %w", err)
	}

	serverItem := &Server{}
	if err := m.ListenConfigItem(serverItem); err != nil {
		return fmt.Errorf("listen server config failed: %w", err)
	}

	mongoItem := &GameMainMongoFormat{}
	if err := m.ListenConfigItem(mongoItem); err != nil {
		return fmt.Errorf("listen mongo config failed: %w", err)
	}

	return nil

}

func (m *ConfigManager) InitNacosClient() error {
	var (
		timeoutMs uint64 = DefaultTimeoutMs
	)

	cfg := m.centerConfig.Nacos
	if cfg.TimeoutMs == 0 {
		timeoutMs = 5000
	}
	// 构建服务器配置
	if len(cfg.ServerHosts) == 0 {
		return fmt.Errorf("nacos server hosts is empty")
	}

	serverConfigs := make([]constant.ServerConfig, 0, len(cfg.ServerHosts))
	for _, host := range cfg.ServerHosts {
		serverHost, serverPort := parseServerAddress(host)
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      serverHost,
			Port:        serverPort,
			ContextPath: "/nacos", // 设置上下文路径
		})
	}

	// 构建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:        cfg.NamespaceID,
		TimeoutMs:          timeoutMs,
		DisableUseSnapShot: true,    // 是否禁用快照功能（测试环境建议设置为 true）
		LogLevel:           "debug", // 使用 debug 级别以便查看详细错误信息
	}

	logger.GetLogger().Info("init nacos client", zap.Any("clientConfig", clientConfig), zap.Any("serverConfigs", serverConfigs))

	// 如果有用户名密码，设置认证
	if cfg.Username != "" && cfg.Password != "" {
		clientConfig.Username = cfg.Username
		clientConfig.Password = cfg.Password
	}

	// 创建 Nacos 配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return fmt.Errorf("create nacos config client failed: %w", err)
	}
	m.configClient = configClient
	return nil
}

// 加载配置
func (m *ConfigManager) LoadConfigItme(item ConfigItem) error {
	if m.configClient == nil {
		return fmt.Errorf("nacos config client is not initialized")
	}

	content, err := m.configClient.GetConfig(vo.ConfigParam{
		DataId: item.GetDataId(),
		Group:  item.GetGroupId(),
	})
	if err != nil {
		// 提供更详细的错误信息，帮助诊断问题
		logger.GetLogger().Error("get config from nacos failed",
			zap.String("dataId", item.GetDataId()),
			zap.String("group", item.GetGroupId()),
			zap.String("namespace", m.centerConfig.Nacos.NamespaceID),
			zap.Strings("servers", m.centerConfig.Nacos.ServerHosts),
			zap.Error(err))
		return err
	}

	if content == "" {
		logger.GetLogger().Error("config content is empty from nacos",
			zap.String("dataId", item.GetDataId()),
			zap.String("group", item.GetGroupId()))
		return fmt.Errorf("config content is empty from nacos")
	}

	return item.Onload(GetConfig(), content)
}

// 加载配置并监听配置变化
func (m *ConfigManager) ListenConfigItem(item ConfigItem) error {
	if m.configClient == nil {
		return fmt.Errorf("nacos config client is not initialized")
	}

	// 首先加载配置
	if err := m.LoadConfigItme(item); err != nil {
		return fmt.Errorf("load config failed: %w", err)
	}

	err := m.configClient.ListenConfig(vo.ConfigParam{
		DataId: item.GetDataId(),
		Group:  item.GetGroupId(),
		OnChange: func(namespace, group, dataId, data string) {
			logger.GetLogger().Info("config change notification received",
				zap.String("dataId", dataId),
				zap.String("group", group),
				zap.String("namespace", namespace),
				zap.Int("contentLength", len(data)))

			err := item.Onload(GetConfig(), data)
			if err != nil {
				logger.GetLogger().Error("failed to load config after change",
					zap.Error(err),
					zap.String("group", group),
					zap.String("dataId", dataId))
				return
			}

			logger.GetLogger().Info("config change processed successfully",
				zap.String("dataId", dataId),
				zap.String("group", group))
		},
	})

	if err != nil {
		return fmt.Errorf("listen config failed: %w", err)
	}

	logger.GetLogger().Info("config listener registered successfully",
		zap.String("dataId", item.GetDataId()),
		zap.String("group", item.GetGroupId()))

	return nil
}
