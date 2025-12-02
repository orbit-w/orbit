package config

type ConfigItem interface {
	GetGroupId() string
	GetDataId() string
	Onload(cfg *Config, content string) error
}
