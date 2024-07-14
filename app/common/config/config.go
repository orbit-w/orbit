package config

import (
	"fmt"
	"github.com/spf13/viper"
	"net"
)

func ParseConfig(path string) {
	viper.SetConfigName("config") // 配置文件的名字（没有扩展名）
	viper.SetConfigType("toml")   // 或者 viper.SetConfigType("TOML")
	viper.AddConfigPath(path)     // 查找配置文件的路径

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file, path: %s, err: %s \n", path, err.Error()))
	}

	if tp := viper.GetString(TagPort); tp == "" {
		viper.Set(TagPort, "8950")
	}
}

func StreamHost() string {
	ip := viper.GetString(TagIp)
	port := viper.GetString(TagPort)
	ipAddr := net.ParseIP(ip)
	return net.JoinHostPort(ipAddr.String(), port)
}
