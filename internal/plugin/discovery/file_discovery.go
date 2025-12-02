package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"gopkg.in/yaml.v3"

	pluginv1 "xiaozhi-server-go/api/v1"
	"xiaozhi-server-go/internal/plugin/config"
)

// Discovery 插件发现接口
type Discovery interface {
	// 开始发现
	Start(ctx context.Context) error
	// 停止发现
	Stop(ctx context.Context) error
	// 发现插件
	Discover(ctx context.Context) ([]*pluginv1.PluginInfo, error)
}

// FileDiscovery 文件系统发现服务
type FileDiscovery struct {
	logger  hclog.Logger
	config  *config.DiscoveryConfig
	plugins map[string]*pluginv1.PluginInfo
	mu      sync.RWMutex
}

// NewFileDiscovery 创建文件系统发现服务
func NewFileDiscovery(cfg *config.DiscoveryConfig, logger hclog.Logger) (*FileDiscovery, error) {
	if cfg == nil {
		cfg = &config.DiscoveryConfig{
			Enabled:      true,
			ScanInterval: 30 * time.Second,
			Paths:        []string{"./plugins"},
		}
	}

	return &FileDiscovery{
		logger:  logger.Named("file-discovery"),
		config:  cfg,
		plugins: make(map[string]*pluginv1.PluginInfo),
	}, nil
}

// Start 开始发现
func (d *FileDiscovery) Start(ctx context.Context) error {
	if !d.config.Enabled {
		d.logger.Info("File discovery disabled")
		return nil
	}

	d.logger.Info("Starting file discovery", "paths", d.config.Paths)

	// 初始扫描
	if err := d.scanPaths(ctx); err != nil {
		d.logger.Warn("Initial scan failed", "error", err)
	}

	d.logger.Info("File discovery started")
	return nil
}

// Stop 停止发现
func (d *FileDiscovery) Stop(ctx context.Context) error {
	d.logger.Info("Stopping file discovery")
	return nil
}

// Discover 发现插件
func (d *FileDiscovery) Discover(ctx context.Context) ([]*pluginv1.PluginInfo, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	plugins := make([]*pluginv1.PluginInfo, 0, len(d.plugins))
	for _, plugin := range d.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// scanPaths 扫描路径
func (d *FileDiscovery) scanPaths(ctx context.Context) error {
	for _, path := range d.config.Paths {
		if err := d.scanPath(ctx, path); err != nil {
			d.logger.Warn("Failed to scan path", "path", path, "error", err)
		}
	}
	return nil
}

// scanPath 扫描单个路径
func (d *FileDiscovery) scanPath(ctx context.Context, rootPath string) error {
	d.logger.Debug("Scanning path", "path", rootPath)

	return filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 跳过目录
		if entry.IsDir() {
			return nil
		}

		// 检查是否为可执行文件
		if !isExecutableFile(path) {
			return nil
		}

		// 发现插件
		pluginInfo, err := d.discoverPlugin(path)
		if err != nil {
			d.logger.Debug("Failed to discover plugin", "path", path, "error", err)
			return nil
		}

		if pluginInfo != nil {
			d.mu.Lock()
			d.plugins[pluginInfo.ID] = pluginInfo
			d.mu.Unlock()
			d.logger.Debug("Plugin discovered", "id", pluginInfo.ID, "name", pluginInfo.Name)
		}

		return nil
	})
}

// discoverPlugin 发现插件信息
func (d *FileDiscovery) discoverPlugin(filePath string) (*pluginv1.PluginInfo, error) {
	// 生成插件ID
	pluginID := generatePluginID(filePath)

	// 检查是否有配置文件
	configPath := filePath + ".yaml"
	if _, err := os.Stat(configPath); err == nil {
		return d.loadPluginFromConfig(configPath, pluginID)
	}

	// 检查是否有 JSON 配置文件
	jsonConfigPath := filePath + ".json"
	if _, err := os.Stat(jsonConfigPath); err == nil {
		return d.loadPluginFromJSONConfig(jsonConfigPath, pluginID)
	}

	// 创建默认插件信息
	return &pluginv1.PluginInfo{
		ID:          pluginID,
		Name:        filepath.Base(filePath),
		Version:     "1.0.0",
		Description: fmt.Sprintf("Discovered plugin at %s", filePath),
		Author:      "Unknown",
		Type:        pluginv1.PluginTypeUtility,
		Capabilities: []string{"execute"},
		Metadata: map[string]interface{}{
			"path":         filePath,
			"discovered_at": time.Now().Unix(),
			"source":       "file_discovery",
		},
	}, nil
}

// loadPluginFromConfig 从配置文件加载插件信息
func (d *FileDiscovery) loadPluginFromConfig(configPath, pluginID string) (*pluginv1.PluginInfo, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var pluginConfig struct {
		Name        string   `yaml:"name"`
		Version     string   `yaml:"version"`
		Description string   `yaml:"description"`
		Author      string   `yaml:"author"`
		Type        string   `yaml:"type"`
		Tags        []string `yaml:"tags"`
		Capabilities []string `yaml:"capabilities"`
		Metadata    map[string]interface{} `yaml:"metadata"`
	}

	if err := yaml.Unmarshal(data, &pluginConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 解析插件类型
	var pluginType pluginv1.PluginType
	switch strings.ToLower(pluginConfig.Type) {
	case "audio":
		pluginType = pluginv1.PluginTypeAudio
	case "llm":
		pluginType = pluginv1.PluginTypeLLM
	case "device":
		pluginType = pluginv1.PluginTypeDevice
	case "utility":
		pluginType = pluginv1.PluginTypeUtility
	default:
		pluginType = pluginv1.PluginTypeCustom
	}

	if pluginConfig.Metadata == nil {
		pluginConfig.Metadata = make(map[string]interface{})
	}
	pluginConfig.Metadata["config_file"] = configPath
	pluginConfig.Metadata["discovered_at"] = time.Now().Unix()
	pluginConfig.Metadata["source"] = "file_discovery"

	return &pluginv1.PluginInfo{
		ID:          pluginID,
		Name:        pluginConfig.Name,
		Version:     pluginConfig.Version,
		Description: pluginConfig.Description,
		Author:      pluginConfig.Author,
		Tags:        pluginConfig.Tags,
		Type:        pluginType,
		Capabilities: pluginConfig.Capabilities,
		Metadata:    pluginConfig.Metadata,
	}, nil
}

// loadPluginFromJSONConfig 从JSON配置文件加载插件信息
func (d *FileDiscovery) loadPluginFromJSONConfig(configPath, pluginID string) (*pluginv1.PluginInfo, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON config file: %w", err)
	}

	var pluginConfig struct {
		Name        string                 `json:"name"`
		Version     string                 `json:"version"`
		Description string                 `json:"description"`
		Author      string                 `json:"author"`
		Type        string                 `json:"type"`
		Tags        []string               `json:"tags"`
		Capabilities []string              `json:"capabilities"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal(data, &pluginConfig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config file: %w", err)
	}

	// 解析插件类型
	var pluginType pluginv1.PluginType
	switch strings.ToLower(pluginConfig.Type) {
	case "audio":
		pluginType = pluginv1.PluginTypeAudio
	case "llm":
		pluginType = pluginv1.PluginTypeLLM
	case "device":
		pluginType = pluginv1.PluginTypeDevice
	case "utility":
		pluginType = pluginv1.PluginTypeUtility
	default:
		pluginType = pluginv1.PluginTypeCustom
	}

	if pluginConfig.Metadata == nil {
		pluginConfig.Metadata = make(map[string]interface{})
	}
	pluginConfig.Metadata["config_file"] = configPath
	pluginConfig.Metadata["discovered_at"] = time.Now().Unix()
	pluginConfig.Metadata["source"] = "file_discovery"

	return &pluginv1.PluginInfo{
		ID:          pluginID,
		Name:        pluginConfig.Name,
		Version:     pluginConfig.Version,
		Description: pluginConfig.Description,
		Author:      pluginConfig.Author,
		Tags:        pluginConfig.Tags,
		Type:        pluginType,
		Capabilities: pluginConfig.Capabilities,
		Metadata:    pluginConfig.Metadata,
	}, nil
}

// isExecutableFile 检查文件是否为可执行文件
func isExecutableFile(filePath string) bool {
	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	executableExts := []string{".exe", ".dll", ".so", ".dylib"}

	for _, executableExt := range executableExts {
		if ext == executableExt {
			return true
		}
	}

	// 检查是否为常见的可执行文件名
	baseName := strings.ToLower(filepath.Base(filePath))
	executableNames := []string{
		"plugin", "main", "run", "start", "server", "daemon",
		"asr", "tts", "llm", "device", "utility",
	}

	for _, executableName := range executableNames {
		if baseName == executableName {
			return true
		}
	}

	return false
}

// generatePluginID 生成插件ID
func generatePluginID(filePath string) string {
	// 使用文件路径的哈希作为ID，或者使用文件名
	baseName := filepath.Base(filePath)
	ext := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, ext)

	// 清理名称，只保留字母数字和连字符
	cleanName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)

	// 确保ID不为空
	if cleanName == "" {
		cleanName = "plugin"
	}

	return cleanName
}

// NewDiscovery 创建发现服务（工厂函数）
func NewDiscovery(cfg *config.DiscoveryConfig, logger hclog.Logger) (Discovery, error) {
	if cfg == nil {
		cfg = &config.DiscoveryConfig{
			Enabled:      true,
			ScanInterval: 30 * time.Second,
			Paths:        []string{"./plugins"},
		}
	}

	return NewFileDiscovery(cfg, logger)
}