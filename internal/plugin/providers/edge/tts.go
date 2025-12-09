package edge

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wujunwei928/edge-tts-go/edge_tts"
)

type TTSConfig struct {
	Voice     string
	OutputDir string
}

func synthesizeSpeech(config *TTSConfig, text string) (string, error) {
	voice := config.Voice
	if voice == "" {
		voice = "zh-CN-XiaoxiaoNeural"
	}

	outputDir := config.OutputDir
	if outputDir == "" {
		outputDir = "data/tmp"
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败 '%s': %v", outputDir, err)
	}
	tempFile := filepath.Join(outputDir, fmt.Sprintf("edge_tts_go_%d.mp3", time.Now().UnixNano()))

	connOptions := []edge_tts.CommunicateOption{
		edge_tts.SetVoice(voice),
	}

	conn, err := edge_tts.NewCommunicate(text, connOptions...)
	if err != nil {
		return "", fmt.Errorf("创建 edge-tts-go Communicate 失败: %v", err)
	}

	audioData, err := conn.Stream()
	if err != nil {
		return "", fmt.Errorf("edge-tts-go 获取音频流失败: %v", err)
	}

	if err := os.WriteFile(tempFile, audioData, 0644); err != nil {
		return "", fmt.Errorf("写入音频文件 '%s' 失败: %v", tempFile, err)
	}

	return tempFile, nil
}
