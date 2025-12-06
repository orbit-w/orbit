package mongodbdriver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type ConfigFormat string

// 配置文件格式
const (
	FormatJSON ConfigFormat = "json"
	FormatYAML ConfigFormat = "yaml"
	FormatTOML ConfigFormat = "toml"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{}
}

// LoadConfigFromBytes 从字节数组加载配置，自动检测格式（支持 JSON、YAML、TOML）
func (l *ConfigLoader) LoadConfigFromBytes(content []byte, format ConfigFormat) (*MongoDBConfig, error) {
	if len(content) == 0 {
		return &MongoDBConfig{}, nil
	}

	var config MongoDBConfig
	var err error

	// 去除首尾空白字符以便检测格式
	trimmed := strings.TrimSpace(string(content))
	if len(trimmed) == 0 {
		return &MongoDBConfig{}, nil
	}

	switch format {
	case FormatJSON:
		err = json.Unmarshal(content, &config)
	case FormatYAML:
		err = yaml.Unmarshal(content, &config)
	case FormatTOML:
		err = toml.Unmarshal(content, &config)
	default:
		return nil, fmt.Errorf("不支持的配置文件格式: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// LoadConfig 加载配置文件
func (l *ConfigLoader) LoadConfig(filePath string) (*MongoDBConfig, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("配置文件未找到: %s", filePath)
	}

	// 读取配置文件
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config MongoDBConfig

	// 根据文件扩展名选择解析方式
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(content, &config)
	case ".toml":
		err = toml.Unmarshal(content, &config)
	default:
		return nil, fmt.Errorf("不支持的配置文件格式: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}
