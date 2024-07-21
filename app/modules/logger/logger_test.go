package logger

import (
	"fmt"
	"go.uber.org/zap"
	"testing"
)

/*
   @Author: orbit-w
   @File: logger_test
   @2024 7月 周六 22:26
*/

func TestInitLogger(t *testing.T) {
	InitLogger("dev")
	l := ZLogger()
	l.Info("Is Test", zap.String("Name", "Test"))
	l.Error("Is Test", zap.String("Name", "Test"))
	wrap()
	//l.DPanic("Is Test", zap.String("Name", "Test"))
	StopLogger()
	fmt.Println("TestInitLogger")
}

func wrap() {
	l := ZLogger()
	l.Error("Is Test", zap.String("Name", "Test"))
	l.Warn("Is Test", zap.String("Name", "Test"))
}
