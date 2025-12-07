package netutils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetLocalIPv4 获取本机非回环的IPv4地址
func GetLocalIPv4() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// 跳过禁用的、回环的接口
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			// 只返回IPv4地址
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				return ip4.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid local IP found")
}

// GetPublicIPv4 获取外网IPv4地址
func GetPublicIPv4() (string, error) {
	// 使用多个备用API服务
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var lastErr error
	for _, service := range services {
		req, err := http.NewRequest("GET", service, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusOK {
			ip := strings.TrimSpace(string(body))
			// 验证是否为有效的IPv4地址
			if net.ParseIP(ip) != nil {
				return ip, nil
			}
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed to get public IP: %w", lastErr)
	}
	return "", fmt.Errorf("failed to get public IP from all services")
}
