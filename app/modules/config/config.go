package config

var (
	manager *ConfigManager
)

// 同一命名空间下，所有Data Id 的配置的集合，用于管理所有配置
type Config struct {
	GameMain *MainConfigGroup `toml:"game_main"`
}

func NewConfig() *Config {
	return &Config{
		GameMain: NewMainConfigGroup(),
	}
}

func (c *Config) GetGameMainConfig() *MainConfigGroup {
	return c.GameMain
}

func (c *Config) GetServerName() string {
	return c.GameMain.Server.Name
}

func (c *Config) GetServerStage() string {
	return c.GameMain.Server.Stage
}

func (c *Config) GetRedisConfig() *GameMainRedis {
	return c.GameMain.GetRedisConfig()
}
