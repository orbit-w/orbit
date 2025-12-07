package config_v2

import (
	"testing"
)

// TestInitConfig 测试配置初始化
func TestInitConfig(t *testing.T) {
	// 注意：此测试需要实际的 Nacos 服务器和配置文件
	// 在实际环境中运行前，请确保：
	// 1. config_center.yaml 文件存在且配置正确
	// 2. Nacos 服务器可访问
	// 3. 相应的配置项已在 Nacos 中配置

	configFile := "./config_center.yaml"

	// 测试配置初始化
	t.Run("InitConfig", func(t *testing.T) {
		err := InitConfig(configFile)
		if err != nil {
			t.Fatalf("InitConfig failed: %v", err)
		}
		defer StopConfig()

		// 测试获取配置
		// 参数：dataId, group, key
		serverName := GetString(DtaIDGameMain, GroupServer, "name")
		t.Logf("Server name: %s", serverName)

		// 测试获取 Redis 配置
		redisAddrs := GetStringSlice(DtaIDGameMain, GroupRedis, "addr")
		t.Logf("Redis addrs: %v", redisAddrs)
	})
}

// TestGetConfigValues 测试获取配置值
func TestGetConfigValues(t *testing.T) {
	t.Skip("跳过集成测试，需要实际的 Nacos 环境")

	err := InitConfig("config_center.yaml")
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	defer StopConfig()

	tests := []struct {
		name   string
		dataId string
		group  string
		key    string
	}{
		{"GetString", "game.main", "DEFAULT_GROUP", "name"},
		{"GetInt", "game.main", "DEFAULT_GROUP", "port"},
		{"GetBool", "game.main", "DEFAULT_GROUP", "enabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := Get(tt.dataId, tt.group, tt.key)
			t.Logf("%s: %v", tt.name, val)
		})
	}
}

// TestUnmarshal 测试反序列化配置
func TestUnmarshal(t *testing.T) {
	t.Skip("跳过集成测试，需要实际的 Nacos 环境")

	type ServerConfig struct {
		Name  string `yaml:"name"`
		Stage string `yaml:"stage"`
		Host  string `yaml:"host"`
		Port  int    `yaml:"port"`
	}

	err := InitConfig("config_center.yaml")
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	defer StopConfig()

	var config ServerConfig
	err = Unmarshal("game.main", "DEFAULT_GROUP", &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	t.Logf("Server config: %+v", config)
}

// TestOnConfigChange 测试配置变更回调
func TestOnConfigChange(t *testing.T) {
	t.Skip("跳过集成测试，需要实际的 Nacos 环境")

	err := InitConfig("config_center.yaml")
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	defer StopConfig()

	// 注册配置变更回调
	OnConfigChange("game.main", "DEFAULT_GROUP", func() {
		t.Log("Config changed!")
		serverName := GetString("game.main", "DEFAULT_GROUP", "name")
		t.Logf("New server name: %s", serverName)
	})

	// 等待配置变更...
	// 注意：这需要手动在 Nacos 控制台修改配置来测试
}
