package utils

import (
	"fmt"
	"net"
)

// GetAvailablePort 获取一个可用的端口号，从指定端口开始尝试
func GetAvailablePort(startPort int) (int, error) {
	for port := startPort; port <= 65535; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found from %d", startPort)
}

// GetAvailablePortWithRange 获取指定范围内可用的端口号
func GetAvailablePortWithRange(startPort, endPort int) (int, error) {
	for port := startPort; port <= endPort; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found in range %d-%d", startPort, endPort)
}

// IsPortAvailable 检查指定端口是否可用
func IsPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}