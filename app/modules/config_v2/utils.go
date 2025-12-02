package config

import (
	"net"
	"strconv"
)

const (
	DefaultPort uint64 = 9848
)

// parseServerAddress 解析服务器地址（支持域名，无端口时使用默认端口 9848）
func parseServerAddress(address string) (string, uint64) {
	if address == "" {
		return "", DefaultPort // 默认端口
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		// 如果没有端口，使用默认端口 DefaultPort
		return address, DefaultPort
	}

	if host == "" {
		// 如果 host 为空，使用整个地址作为 host
		return address, DefaultPort
	}

	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return host, DefaultPort
	}

	if port == 0 {
		port = DefaultPort // 确保端口不为 0
	}

	return host, port
}
