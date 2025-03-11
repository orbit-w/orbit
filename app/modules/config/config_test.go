package config

import (
	"fmt"
	"testing"
)

/*
   @Author: orbit-w
   @File: config_test
   @2024 7月 周六 20:29
*/

func TestLoadConfig(t *testing.T) {
	LoadConfig("./config.toml")
	c := GetConfig()
	fmt.Println(c)
}
