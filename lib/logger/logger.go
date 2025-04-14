package logger

import (
	mlog "gitee.com/orbit-w/meteor/modules/mlog"
)

var logger = mlog.NewFileLogger(mlog.WithLevel("info"),
	mlog.WithFormat("console"),
	mlog.WithRotation(500, 7, 3, false),
	mlog.WithInitialFields(map[string]any{"app": "orbit-server"}),
	mlog.WithOutputPaths("logs/orbit.log"))

func SetLogger(log *mlog.Logger) {
	logger = log
}

func GetLogger() *mlog.Logger {
	return logger
}

func StopLogger() {
	logger.Sync()
}
