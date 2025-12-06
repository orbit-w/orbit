package config

import (
	"fmt"

	mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"go.uber.org/zap"
)

type GameMainMongoFormat struct{}

func (m *GameMainMongoFormat) GetGroupId() string {
	return "mongodb"
}

func (m *GameMainMongoFormat) GetDataId() string {
	return "game.main"
}

func (m *GameMainMongoFormat) Onload(cfg *Config, content string) error {
	mongoConfig, err := mongodbdriver.NewConfigLoader().LoadConfigFromBytes([]byte(content), mongodbdriver.FormatYAML)
	if err != nil {
		logger.GetLogger().Error("failed to load game main mongo config", zap.Error(err), zap.String("group", m.GetGroupId()), zap.String("dataId", m.GetDataId()))
		return fmt.Errorf("failed to load game main mongo config: %w", err)
	}
	cfg.GetGameMainConfig().SetMongoConfig(mongoConfig)
	return nil
}
