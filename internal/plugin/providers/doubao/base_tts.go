package doubao

import (
	"fmt"
	"os"
	"path/filepath"
	"xiaozhi-server-go/internal/platform/config"
	internalutils "xiaozhi-server-go/internal/utils"
)

// TTSConfig TTS配置结构
type TTSConfig struct {
	Name            string              `yaml:"name"` // TTS提供者名称
	Type            string              `yaml:"type"`
	OutputDir       string              `yaml:"output_dir"`
	Voice           string              `yaml:"voice,omitempty"`
	Format          string              `yaml:"format,omitempty"`
	SampleRate      int                 `yaml:"sample_rate,omitempty"`
	AppID           string              `yaml:"appid"`
	Token           string              `yaml:"token"`
	Cluster         string              `yaml:"cluster"`
	SupportedVoices []config.VoiceInfo `yaml:"supported_voices"` // 支持的语音列表
}

// BaseTTS TTS基础实现
type BaseTTS struct {
	config     *TTSConfig
	deleteFile bool
	sessionID  string // 会话ID，用于事件发布
}

// Config 获取配置
func (p *BaseTTS) Config() *TTSConfig {
	return p.config
}

// GetSessionID 获取会话ID
func (p *BaseTTS) GetSessionID() string {
	return p.sessionID
}

// SetSessionID 设置会话ID
func (p *BaseTTS) SetSessionID(sessionID string) {
	p.sessionID = sessionID
}

// NewBaseTTS 创建TTS基础提供者
func NewBaseTTS(config *TTSConfig, deleteFile bool) *BaseTTS {
	return &BaseTTS{
		config:     config,
		deleteFile: deleteFile,
	}
}

// Initialize 初始化提供者
func (p *BaseTTS) Initialize() error {
	if err := os.MkdirAll(p.config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}
	return nil
}

func IsSupportedVoice(voice string, supportedVoices []config.VoiceInfo) (bool, string, error) {
	if voice == "" {
		return false, "", fmt.Errorf("声音不能为空")
	}
	cnNames := map[string]string{}
	enNames := map[string]string{}
	voiceNames := []string{}
	for _, v := range supportedVoices {
		cnNames[v.DisplayName] = v.Name // 中文名
		enNames[v.Name] = v.Name        // 英文名（实际是音色名）
		voiceNames = append(voiceNames, v.Name)
	}

	// 如果是中文名，则转换为音色名称
	if enVoice, ok := cnNames[voice]; ok {
		voice = enVoice
	}

	// 如果是英文名，则转换为音色名称
	if enVoice, ok := enNames[voice]; ok {
		voice = enVoice
	}

	// 检查声音是否在支持的列表中
	if !internalutils.IsInArray(voice, voiceNames) {
		return false, "", fmt.Errorf("不支持的声音: %s, 可用声音: %v", voice, voiceNames)
	}

	return true, voice, nil
}

func (p *BaseTTS) SetVoice(voice string) (error, string) {
	p.Config().Voice = voice
	return nil, voice
}

// Cleanup 清理资源
func (p *BaseTTS) Cleanup() error {
	if p.deleteFile {
		// 清理输出目录中的临时文件
		pattern := filepath.Join(p.config.OutputDir, "*.{wav,mp3,opus}")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("查找临时文件失败: %v", err)
		}
		for _, file := range matches {
			if err := os.Remove(file); err != nil {
				return fmt.Errorf("删除临时文件失败: %v", err)
			}
		}
	}
	return nil
}
