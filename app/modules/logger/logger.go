package logger

import (
	"github.com/orbit-w/meteor/bases/logger"
	"go.uber.org/zap"
	"strings"
)

/*
   @Author: orbit-w
   @File: logger
   @2024 4月 周日 14:34
*/

var gLogger *zap.Logger

func InitLogger(stage string) {
	switch {
	case strings.Contains(stage, "dev") ||
		strings.Contains(stage, "local") ||
		strings.Contains(stage, "test"):
		gLogger = logger.NewDevelopmentLogger()
	default:
		gLogger = logger.New("logs/app.log", zap.InfoLevel)
	}
}

func StopLogger() {
	logger.Stop(gLogger)
}

func ZLogger() *zap.Logger {
	return gLogger
}
