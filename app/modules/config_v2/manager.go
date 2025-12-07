package config_v2

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/go-viper/mapstructure/v2"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	DefaultTimeoutMs uint64 = 5000 // 默认5秒超时
	DefaultPort      uint64 = 8848 // 默认Nacos端口

	StateNormal = iota
	StateStopping
	StateStopped
)

// ConfigManager 配置管理器，使用 Nacos SDK 获取配置，注入到 Viper 中
type ConfigManager struct {
	centerConfig *CenterConfig               // 配置中心基础配置
	configClient config_client.IConfigClient // Nacos 配置客户端
	vipers       map[string]*viper.Viper     // key: dataId.group, value: viper实例
	mu           sync.RWMutex                // 保护 vipers map
	state        atomic.Int32                // 状态
	callbacks    map[string][]func()         // 配置变更回调, key: dataId.group
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		centerConfig: &CenterConfig{},
		vipers:       make(map[string]*viper.Viper),
		callbacks:    make(map[string][]func()),
	}
}

// Start 启动配置管理器
func (m *ConfigManager) Start(filename string) error {
	// 加载配置中心配置
	if err := m.loadCenterConfig(filename); err != nil {
		return fmt.Errorf("load center config failed: %w", err)
	}

	// 初始化 Nacos 客户端
	if err := m.initNacosClient(); err != nil {
		return fmt.Errorf("init nacos client failed: %w", err)
	}

	// 初始化所有配置源
	if err := m.initAllSources(); err != nil {
		return fmt.Errorf("init sources failed: %w", err)
	}

	logger.GetLogger().Info("config manager started successfully")
	return nil
}

// Stop 停止配置管理器
func (m *ConfigManager) Stop() error {
	if !m.state.CompareAndSwap(StateNormal, StateStopping) {
		return fmt.Errorf("config manager is already stopping")
	}

	if m.configClient != nil {
		m.configClient.CloseClient()
	}

	m.state.Store(StateStopped)
	logger.GetLogger().Info("config manager stopped")
	return nil
}

// loadCenterConfig 加载配置中心配置文件
func (m *ConfigManager) loadCenterConfig(filename string) error {
	v := viper.New()
	v.SetConfigFile(filename)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file failed: %w", err)
	}

	// 配置 mapstructure 使用 yaml tag 而不是默认的 mapstructure tag
	if err := v.Unmarshal(m.centerConfig, func(config *mapstructure.DecoderConfig) {
		config.TagName = "yaml"
	}); err != nil {
		return fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 设置默认值
	if m.centerConfig.Nacos.TimeoutMs == 0 {
		m.centerConfig.Nacos.TimeoutMs = DefaultTimeoutMs
	}

	return nil
}

// initNacosClient 初始化 Nacos 配置客户端
func (m *ConfigManager) initNacosClient() error {
	cfg := m.centerConfig.Nacos
	timeoutMs := cfg.TimeoutMs
	if timeoutMs == 0 {
		timeoutMs = DefaultTimeoutMs
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
			ContextPath: "/nacos",
		})
	}

	// 构建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:        cfg.NamespaceID,
		TimeoutMs:          timeoutMs,
		DisableUseSnapShot: true,
		LogLevel:           "info",
	}

	// 如果有用户名密码，设置认证
	if cfg.Username != "" && cfg.Password != "" {
		clientConfig.Username = cfg.Username
		clientConfig.Password = cfg.Password
	}

	logger.GetLogger().Info("initializing nacos client",
		zap.Any("clientConfig", clientConfig),
		zap.Any("serverConfigs", serverConfigs))

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

// parseServerAddress 解析服务器地址（支持域名，无端口时使用默认端口 8848）
func parseServerAddress(address string) (string, uint64) {
	if address == "" {
		return "", DefaultPort
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		// 如果没有端口，使用默认端口
		return address, DefaultPort
	}

	if host == "" {
		return address, DefaultPort
	}

	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return host, DefaultPort
	}

	if port == 0 {
		port = DefaultPort
	}

	return host, port
}

// initAllSources 初始化所有配置源
func (m *ConfigManager) initAllSources() error {
	for _, source := range m.centerConfig.Sources {
		if err := m.loadConfigSource(source); err != nil {
			return fmt.Errorf("load source %s/%s failed: %w", source.Group, source.DataID, err)
		}

		// 监听配置变化
		if source.NeedListen {
			if err := m.listenConfigSource(source); err != nil {
				return fmt.Errorf("listen source %s/%s failed: %w", source.Group, source.DataID, err)
			}
		}
	}
	return nil
}

// loadConfigSource 从 Nacos 加载配置并注入到 Viper
func (m *ConfigManager) loadConfigSource(source *ConfigSource) error {
	if m.configClient == nil {
		return fmt.Errorf("nacos config client is not initialized")
	}

	// 从 Nacos 获取配置内容
	content, err := m.configClient.GetConfig(vo.ConfigParam{
		DataId: source.DataID,
		Group:  source.Group,
	})
	if err != nil {
		logger.GetLogger().Error("get config from nacos failed",
			zap.String("dataId", source.DataID),
			zap.String("group", source.Group),
			zap.String("namespace", m.centerConfig.Nacos.NamespaceID),
			zap.Error(err))
		return fmt.Errorf("get config from nacos failed: %w", err)
	}

	if content == "" {
		logger.GetLogger().Warn("config content is empty from nacos",
			zap.String("dataId", source.DataID),
			zap.String("group", source.Group))
		return fmt.Errorf("config content is empty")
	}

	// 创建 Viper 实例并注入配置内容
	v := viper.New()
	v.SetConfigType(source.Format)

	// 将配置内容读入 Viper
	if err := v.ReadConfig(bytes.NewBufferString(content)); err != nil {
		logger.GetLogger().Error("read config into viper failed",
			zap.String("dataId", source.DataID),
			zap.Error(err))
		return fmt.Errorf("read config into viper failed: %w", err)
	}

	// 保存 Viper 实例
	m.SetViper(source.DataID, source.Group, v)
	logger.GetLogger().Info("config source loaded successfully",
		zap.String("dataId", source.DataID),
		zap.String("group", source.Group),
		zap.String("format", source.Format),
		zap.Int("contentLength", len(content)))

	return nil
}

// listenConfigSource 监听配置变化
func (m *ConfigManager) listenConfigSource(source *ConfigSource) error {
	if m.configClient == nil {
		return fmt.Errorf("nacos config client is not initialized")
	}

	err := m.configClient.ListenConfig(vo.ConfigParam{
		DataId: source.DataID,
		Group:  source.Group,
		OnChange: func(namespace, group, dataId, data string) {
			logger.GetLogger().Info("config change notification received",
				zap.String("dataId", dataId),
				zap.String("group", group),
				zap.String("namespace", namespace),
				zap.Int("contentLength", len(data)))

			// 重新加载配置到 Viper
			v := viper.New()
			v.SetConfigType(source.Format)

			if err := v.ReadConfig(bytes.NewBufferString(data)); err != nil {
				logger.GetLogger().Error("reload config into viper failed",
					zap.String("dataId", dataId),
					zap.Error(err))
				return
			}

			// 更新 Viper 实例
			m.SetViper(dataId, group, v)

			logger.GetLogger().Info("config reloaded successfully",
				zap.String("dataId", dataId),
				zap.String("group", group))

			// 执行回调
			m.executeCallbacks(dataId, group)
		},
	})

	if err != nil {
		return fmt.Errorf("listen config failed: %w", err)
	}

	logger.GetLogger().Info("config listener registered successfully",
		zap.String("dataId", source.DataID),
		zap.String("group", source.Group))

	return nil
}

// executeCallbacks 执行配置变更回调
func (m *ConfigManager) executeCallbacks(dataId, group string) {
	key := m.genConfigNameSpaceId(dataId, group)
	m.mu.RLock()
	callbacks := m.callbacks[key]
	m.mu.RUnlock()

	for _, callback := range callbacks {
		callback()
	}
}

// OnConfigChange 注册配置变更回调
func (m *ConfigManager) OnConfigChange(dataId, group string, callback func()) {
	key := m.genConfigNameSpaceId(dataId, group)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks[key] = append(m.callbacks[key], callback)
}

func (m *ConfigManager) SetViper(dataId, group string, viper *viper.Viper) {
	key := m.genConfigNameSpaceId(dataId, group)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.vipers[key] = viper
}

// GetViper 获取指定 DataId 和 Group 的 viper 实例
func (m *ConfigManager) GetViper(dataId, group string) *viper.Viper {
	key := m.genConfigNameSpaceId(dataId, group)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.vipers[key]
}

// Get 获取配置值（从指定的 DataId 和 Group）
func (m *ConfigManager) Get(dataId, group, key string) any {
	v := m.GetViper(dataId, group)
	if v == nil {
		logger.GetLogger().Warn("viper not found",
			zap.String("dataId", dataId),
			zap.String("group", group))
		return nil
	}
	return v.Get(key)
}

// GetString 获取字符串配置
func (m *ConfigManager) GetString(dataId, group, key string) string {
	v := m.GetViper(dataId, group)
	if v == nil {
		return ""
	}
	return v.GetString(key)
}

// GetInt 获取整数配置
func (m *ConfigManager) GetInt(dataId, group, key string) int {
	v := m.GetViper(dataId, group)
	if v == nil {
		return 0
	}
	return v.GetInt(key)
}

// GetBool 获取布尔配置
func (m *ConfigManager) GetBool(dataId, group, key string) bool {
	v := m.GetViper(dataId, group)
	if v == nil {
		return false
	}
	return v.GetBool(key)
}

// GetStringSlice 获取字符串数组配置
func (m *ConfigManager) GetStringSlice(dataId, group, key string) []string {
	v := m.GetViper(dataId, group)
	if v == nil {
		return nil
	}
	return v.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置
func (m *ConfigManager) GetStringMap(dataId, group, key string) map[string]any {
	v := m.GetViper(dataId, group)
	if v == nil {
		return nil
	}
	return v.GetStringMap(key)
}

// Unmarshal 将配置反序列化到结构体（从指定的 DataId 和 Group）
func (m *ConfigManager) Unmarshal(dataId, group string, rawVal any) error {
	v := m.GetViper(dataId, group)
	if v == nil {
		return fmt.Errorf("viper not found for dataId: %s, group: %s", dataId, group)
	}
	return v.Unmarshal(rawVal)
}

// UnmarshalKey 将配置的某个 key 反序列化到结构体
func (m *ConfigManager) UnmarshalKey(dataId, group, key string, rawVal any) error {
	v := m.GetViper(dataId, group)
	if v == nil {
		return fmt.Errorf("viper not found for dataId: %s, group: %s", dataId, group)
	}
	return v.UnmarshalKey(key, rawVal)
}

func (m *ConfigManager) genConfigNameSpaceId(dataId string, group string) string {
	return fmt.Sprintf("%s.%s", dataId, group)
}
