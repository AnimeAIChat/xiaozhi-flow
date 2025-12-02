package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// HTTPPluginManager HTTP插件管理器实现
type HTTPPluginManager struct {
	mu      sync.RWMutex
	plugins map[string]*PluginProcess
	client  *http.Client
	logger  Logger
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewHTTPPluginManager 创建HTTP插件管理器
func NewHTTPPluginManager(logger Logger) *HTTPPluginManager {
	return &HTTPPluginManager{
		plugins: make(map[string]*PluginProcess),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// StartPlugin 启动插件进程
func (pm *HTTPPluginManager) StartPlugin(ctx context.Context, pluginID string) (*PluginProcess, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 检查插件是否已存在
	if plugin, exists := pm.plugins[pluginID]; exists {
		if plugin.Status == PluginStatusRunning {
			return plugin, nil
		}
	}

	// 模拟启动HTTP插件进程
	plugin := &PluginProcess{
		ID:        pluginID,
		Name:      fmt.Sprintf("HTTP Plugin %s", pluginID),
		Version:   "1.0.0",
		Status:    PluginStatusStarting,
		Config: PluginConfig{
			HTTP: &HTTPPluginConfig{
				URL:     fmt.Sprintf("http://localhost:%d/%s", 9000+len(pm.plugins), pluginID),
				Method:  "POST",
				Headers: map[string]string{
					"Content-Type": "application/json",
					"User-Agent":   "xiaozhi-flow/1.0",
				},
				Timeout: 30 * time.Second,
			},
		},
		StartTime: time.Now(),
		Stats:     &PluginStats{},
		Metadata: map[string]string{
			"type":    "http",
			"runtime": "simulated",
		},
	}

	// 启动HTTP服务器模拟插件
	go pm.startHTTPPluginServer(plugin)

	// 设置插件状态为运行中
	plugin.Status = PluginStatusRunning
	pm.plugins[pluginID] = plugin

	pm.logger.Info("Plugin started successfully", "plugin_id", pluginID, "url", plugin.Config.HTTP.URL)

	return plugin, nil
}

// StopPlugin 停止插件进程
func (pm *HTTPPluginManager) StopPlugin(ctx context.Context, pluginID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	if plugin.Status != PluginStatusRunning {
		return fmt.Errorf("plugin %s is not running", pluginID)
	}

	plugin.Status = PluginStatusStopping
	now := time.Now()
	plugin.EndTime = &now

	// 模拟停止插件
	plugin.Status = PluginStatusStopped

	pm.logger.Info("Plugin stopped successfully", "plugin_id", pluginID)

	return nil
}

// GetPlugin 获取插件进程
func (pm *HTTPPluginManager) GetPlugin(pluginID string) (*PluginProcess, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[pluginID]
	return plugin, exists
}

// HealthCheck 健康检查
func (pm *HTTPPluginManager) HealthCheck(ctx context.Context, pluginID string) (*PluginHealth, error) {
	plugin, exists := pm.GetPlugin(pluginID)
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	if plugin.Status != PluginStatusRunning {
		return &PluginHealth{
			Status:    "unhealthy",
			LastCheck: time.Now(),
			Message:   "Plugin is not running",
		}, nil
	}

	start := time.Now()

	// 模拟健康检查请求
	healthURL := fmt.Sprintf("%s/health", plugin.Config.HTTP.URL)
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pm.client.Do(req)
	if err != nil {
		return &PluginHealth{
			Status:       "unhealthy",
			LastCheck:    time.Now(),
			ResponseTime: time.Since(start),
			Message:      err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	responseTime := time.Since(start)

	if resp.StatusCode == http.StatusOK {
		return &PluginHealth{
			Status:       "healthy",
			LastCheck:    time.Now(),
			ResponseTime: responseTime,
			Message:      "Plugin is responding normally",
		}, nil
	}

	return &PluginHealth{
		Status:       "unhealthy",
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
		Message:      fmt.Sprintf("HTTP status: %d", resp.StatusCode),
	}, nil
}

// CallPlugin 调用插件方法
func (pm *HTTPPluginManager) CallPlugin(ctx context.Context, pluginID, method string, payload map[string]interface{}) (map[string]interface{}, error) {
	plugin, exists := pm.GetPlugin(pluginID)
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	if plugin.Status != PluginStatusRunning {
		return nil, fmt.Errorf("plugin %s is not running", pluginID)
	}

	// 更新统计信息
	plugin.Stats.CallCount++
	plugin.Stats.LastCalled = time.Now()

	start := time.Now()

	// 构造请求URL
	callURL := fmt.Sprintf("%s/call", plugin.Config.HTTP.URL)

	// 构造请求体
	requestBody := map[string]interface{}{
		"method":  method,
		"payload": payload,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", callURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for key, value := range plugin.Config.HTTP.Headers {
		req.Header.Set(key, value)
	}

	resp, err := pm.client.Do(req)
	if err != nil {
		plugin.Stats.ErrorCount++
		return nil, fmt.Errorf("failed to call plugin: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		plugin.Stats.ErrorCount++
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		plugin.Stats.ErrorCount++
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 更新统计信息
	plugin.Stats.SuccessCount++
	elapsed := time.Since(start)
	if plugin.Stats.AvgLatency == 0 {
		plugin.Stats.AvgLatency = elapsed
	} else {
		plugin.Stats.AvgLatency = (plugin.Stats.AvgLatency + elapsed) / 2
	}

	pm.logger.Debug("Plugin method called successfully",
		"plugin_id", pluginID,
		"method", method,
		"elapsed_ms", elapsed.Milliseconds(),
	)

	return result, nil
}

// ListPlugins 获取所有插件状态
func (pm *HTTPPluginManager) ListPlugins() map[string]*PluginProcess {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]*PluginProcess)
	for id, plugin := range pm.plugins {
		// 创建副本以避免并发问题
		pluginCopy := *plugin
		result[id] = &pluginCopy
	}

	return result
}

// startHTTPPluginServer 启动HTTP插件服务器模拟
func (pm *HTTPPluginManager) startHTTPPluginServer(plugin *PluginProcess) {
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"plugin_id": plugin.ID,
			"timestamp": time.Now().Unix(),
		})
	})

	// 调用端点
	mux.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		method, _ := request["method"].(string)
		payload, _ := request["payload"].(map[string]interface{})

		// 模拟插件方法调用
		result := pm.simulatePluginCall(plugin.ID, method, payload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	})

	// 信息端点
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      plugin.ID,
			"name":    plugin.Name,
			"version": plugin.Version,
			"type":    "http",
			"status":  "running",
		})
	})

	// 启动HTTP服务器
	port := 9000
	if u, err := url.Parse(plugin.Config.HTTP.URL); err == nil {
		if p, err := getPortFromURL(u); err == nil {
			port = p
		}
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	pm.logger.Info("HTTP plugin server started", "plugin_id", plugin.ID, "port", port)

	// 启动服务器（这里应该是阻塞的）
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		pm.logger.Error("HTTP plugin server error", "plugin_id", plugin.ID, "error", err)
	}
}

// simulatePluginCall 模拟插件方法调用
func (pm *HTTPPluginManager) simulatePluginCall(pluginID, method string, payload map[string]interface{}) map[string]interface{} {
	switch method {
	case "echo":
		return map[string]interface{}{
			"success": true,
			"data":    payload,
			"message": "Echo successful",
		}
	case "delay":
		// 模拟延迟
		delay := 1 * time.Second
		if d, ok := payload["delay"].(string); ok {
			if parsed, err := time.ParseDuration(d); err == nil {
				delay = parsed
			}
		}
		time.Sleep(delay)
		return map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"delay": delay.String(),
			},
			"message": "Delay completed",
		}
	case "error":
		return map[string]interface{}{
			"success": false,
			"error":   "Simulated plugin error",
			"code":    500,
		}
	case "calculate":
		// 模拟计算功能
		a, _ := payload["a"].(float64)
		b, _ := payload["b"].(float64)
		op, _ := payload["op"].(string)

		var result float64
		switch op {
		case "+":
			result = a + b
		case "-":
			result = a - b
		case "*":
			result = a * b
		case "/":
			if b != 0 {
				result = a / b
			} else {
				return map[string]interface{}{
					"success": false,
					"error":   "Division by zero",
				}
			}
		default:
			return map[string]interface{}{
				"success": false,
				"error":   "Unknown operation: " + op,
			}
		}

		return map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"a":      a,
				"b":      b,
				"op":     op,
				"result": result,
			},
			"message": "Calculation completed",
		}
	default:
		return map[string]interface{}{
			"success": false,
			"error":   "Unknown method: " + method,
		}
	}
}

// getPortFromURL 从URL中提取端口号
func getPortFromURL(u *url.URL) (int, error) {
	if u.Port() != "" {
		return strconv.Atoi(u.Port())
	}

	switch u.Scheme {
	case "http":
		return 80, nil
	case "https":
		return 443, nil
	default:
		return 0, fmt.Errorf("unknown scheme: %s", u.Scheme)
	}
}