package doubao

import (
	"bytes"
	"time"
)

// AsrEventListener defines the interface for ASR events
type AsrEventListener interface {
	OnAsrResult(result string, isFinalResult bool) bool
}

// ASRConfig ASR配置结构
type ASRConfig struct {
	Name string `yaml:"name"` // ASR提供者名称
	Type string
	Data map[string]interface{}
}

// BaseASR ASR基础实现
type BaseASR struct {
	config     *ASRConfig
	deleteFile bool

	// 音频处理相关
	lastChunkTime time.Time
	audioBuffer   *bytes.Buffer

	// 静音检测配置
	silenceThreshold float64 // 能量阈值
	silenceDuration  int     // 静音持续时间(ms)

	BEnableSilenceDetection bool      // 是否启用静音检测
	StartListenTime         time.Time // 最后一次ASR处理时间
	SilenceCount            int       // 连续静音计数

	UserPreferences map[string]interface{}

	listener   AsrEventListener
	sessionID  string // 会话ID，用于事件发布
}

func (p *BaseASR) ResetStartListenTime() {
	p.StartListenTime = time.Now()
}

func (p *BaseASR) SilenceTime() time.Duration {
	if !p.BEnableSilenceDetection {
		return 0
	}
	if p.StartListenTime.IsZero() {
		return 0
	}
	return time.Since(p.StartListenTime)
}

func (p *BaseASR) EnableSilenceDetection(bEnable bool) {
	p.BEnableSilenceDetection = bEnable
}

func (p *BaseASR) GetSilenceCount() int {
	return p.SilenceCount
}

func (p *BaseASR) ResetSilenceCount() {
	p.SilenceCount = 0
}

// SetListener 设置事件监听器
func (p *BaseASR) SetListener(listener AsrEventListener) {
	p.listener = listener
}

// GetListener 获取事件监听器
func (p *BaseASR) GetListener() AsrEventListener {
	return p.listener
}

// SetSessionID 设置会话ID
func (p *BaseASR) SetSessionID(sessionID string) {
	p.sessionID = sessionID
}

// GetSessionID 获取会话ID
func (p *BaseASR) GetSessionID() string {
	return p.sessionID
}

func (p *BaseASR) SetUserPreferences(preferences map[string]interface{}) error {
	p.UserPreferences = preferences
	return nil
}

// Config 获取配置
func (p *BaseASR) Config() *ASRConfig {
	return p.config
}


// GetAudioBuffer 获取音频缓冲区
func (p *BaseASR) GetAudioBuffer() *bytes.Buffer {
	return p.audioBuffer
}

// GetLastChunkTime 获取最后一个音频块的时间
func (p *BaseASR) GetLastChunkTime() time.Time {
	return p.lastChunkTime
}

// SetLastChunkTime 设置最后一个音频块的时间
func (p *BaseASR) SetLastChunkTime(t time.Time) {
	p.lastChunkTime = t
}

// DeleteFile 获取是否删除文件标志
func (p *BaseASR) DeleteFile() bool {
	return p.deleteFile
}

// NewBaseASR 创建ASR基础提供者
func NewBaseASR(config *ASRConfig, deleteFile bool) *BaseASR {
	return &BaseASR{
		config:     config,
		deleteFile: deleteFile,
	}
}

// Initialize 初始化提供者
func (p *BaseASR) Initialize() error {
	return nil
}

// Cleanup 清理资源
func (p *BaseASR) Cleanup() error {
	return nil
}

// 初始化音频处理
func (p *BaseASR) InitAudioProcessing() {
	p.audioBuffer = new(bytes.Buffer)
	p.silenceThreshold = 0.01 // 默认能量阈值
	p.silenceDuration = 800   // 默认静音判断时长(ms)
}
