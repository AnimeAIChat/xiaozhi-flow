package transport

import (
	"xiaozhi-server-go/internal/platform/logging"
	"context"
	"sync"
	"time"
	"xiaozhi-server-go/internal/platform/config"
	internalutils "xiaozhi-server-go/internal/utils"
)

// TransportManager 传输管理器
type TransportManager struct {
	transports map[string]Transport
	logger     *logging.Logger
	config     *config.Config
	mu         sync.RWMutex
}

// NewTransportManager 创建新的传输管理器
func NewTransportManager(config *config.Config, logger *logging.Logger) *TransportManager {
	return &TransportManager{
		transports: make(map[string]Transport),
		logger:     logger,
		config:     config,
	}
}

// RegisterTransport 注册传输层
func (m *TransportManager) RegisterTransport(name string, transport Transport) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transports[name] = transport
	m.logger.Debug("注册传输层: %s (%s)", name, transport.GetType())
}

// Start 启动所有传输层（实现TransportManager接口）
func (m *TransportManager) Start(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, transport := range m.transports {
		// m.logger.Info(fmt.Sprintf("启动传输层: %s", name))

		// 为每个传输层启动独立的goroutine
		go func(name string, transport Transport) {
			if err := transport.Start(ctx); err != nil {
				m.logger.Error("传输层 %s 运行失败: %v", name, err)
			}
		}(name, transport)
	}

	m.StartTicker(ctx)

	return nil
}

func (m *TransportManager) StartTicker(ctx context.Context) {
	// 设置定时器，打印各个传输层的状态信息
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				clientCnt := 0
				sessionCnt := 0
				for _, transport := range m.transports {
					c, s := transport.GetActiveConnectionCount()
					clientCnt += c
					sessionCnt += s
				}
				//m.logger.Info("当前活跃连接数: %d, 当前活跃会话数: %d", clientCnt, sessionCnt)
				systemMemoryUse, _ := internalutils.GetSystemMemoryUsage()
				systemCPUUse, _ := internalutils.GetSystemCPUUsage()
				// Database functionality removed - server status updates disabled
				m.logger.Debug("Server status: CPU %.1f%%, Memory %.1f%%, Connections %d, Sessions %d", systemCPUUse, systemMemoryUse, clientCnt, sessionCnt)
			}
		}
	}()
}

// Stop 停止所有传输层（实现TransportManager接口）
func (m *TransportManager) Stop() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastErr error
	for name, transport := range m.transports {
		if err := transport.Stop(); err != nil {
			m.logger.Error("停止传输层 %s 失败: %v", name, err)
			lastErr = err
		}
	}
	return lastErr
}

// CloseDeviceConnection 关闭指定设备的连接
func (m *TransportManager) CloseDeviceConnection(deviceID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastErr error
	for name, transport := range m.transports {
		if err := transport.CloseDeviceConnection(deviceID); err != nil {
			m.logger.Error("传输层 %s 关闭设备连接失败: %v", name, err)
			lastErr = err
		}
	}
	return lastErr
}

// GetStats 获取传输管理器统计信息（实现TransportManager接口）
func (m *TransportManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})

	clientCount := 0
	sessionCount := 0
	transportStats := make(map[string]interface{})

	for name, transport := range m.transports {
		c, s := transport.GetActiveConnectionCount()
		clientCount += c
		sessionCount += s

		transportStats[name] = map[string]interface{}{
			"type":       transport.GetType(),
			"clients":    c,
			"sessions":   s,
			"status":     "running", // 可以根据实际情况调整
		}
	}

	stats["total_clients"] = clientCount
	stats["total_sessions"] = sessionCount
	stats["transport_count"] = len(m.transports)
	stats["transports"] = transportStats

	return stats
}

// GetTransport 获取指定名称的传输层
func (m *TransportManager) GetTransport(name string) Transport {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.transports[name]
}


