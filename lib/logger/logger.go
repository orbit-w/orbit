package logger

import (
	mlog "github.com/orbit-w/meteor/modules/mlog"
)

var logger = mlog.NewFileLogger(mlog.WithLevel("info"),
	mlog.WithFormat("console"),
	mlog.WithRotation(500, 7, 3, false),
	mlog.WithInitialFields(map[string]any{"app": "content-moderation"}),
	mlog.WithOutputPaths("logs/content-moderation.log"))

func SetLogger(log *mlog.Logger) {
	logger = log
}

func GetLogger() *mlog.Logger {
	return logger
}

func StopLogger() {
	logger.Sync()
}
