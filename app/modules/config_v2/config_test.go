package config

import (
	"fmt"
	"testing"
	"time"
)

func Test_InitConfig_Success(t *testing.T) {
	InitConfig("./config_center.toml")
	conf := GetConfig()
	fmt.Println(conf)
}

func Test_SubscribeConfig(t *testing.T) {
	InitConfig("./config_center.toml")
	conf := GetConfig()
	fmt.Println(conf)
	time.Sleep(time.Minute * 2)
}
