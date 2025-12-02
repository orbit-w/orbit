package config

import (
	"fmt"
	"testing"
)

func Test_InitConfig_Success(t *testing.T) {
	InitConfig("./config_center.toml")
	conf := GetConfig()
	fmt.Println(conf)
}
