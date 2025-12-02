package config

import (
	"fmt"

	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

type Server struct {
	Name  string `yaml:"name"`  // 服务名称
	Stage string `yaml:"stage"` // 环境
	Host  string `yaml:"host"`  // 主机地址
	Port  string `yaml:"port"`  // 端口
}

func (s *Server) GetGroupId() string {
	return "server"
}

func (s *Server) GetDataId() string {
	return "game-main.yaml"
}

func (s *Server) Onload(cfg *Config, content string) error {
	server := new(Server)
	if err := toml.Unmarshal([]byte(content), server); err != nil {
		logger.GetLogger().Error("failed to unmarshal server config", zap.Error(err), zap.String("group", s.GetGroupId()), zap.String("dataId", s.GetDataId()))
		return fmt.Errorf("failed to unmarshal server config: %w", err)
	}

	cfg.GameMain.Server = *server
	return nil
}
