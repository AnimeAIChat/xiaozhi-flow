package gosherpa

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

type TTSConfig struct {
	Cluster   string
	OutputDir string
}

func synthesizeSpeech(config *TTSConfig, text string) (string, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.DialContext(context.Background(), config.Cluster, nil)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// 获取配置的声音，如果未配置则使用默认值
	startTime := time.Now()

	// 创建临时文件路径用于保存 SherpaTTS 生成的 MP3
	outputDir := config.OutputDir
	if outputDir == "" {
		outputDir = "data/tmp"
	}
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败 '%s': %v", outputDir, err)
	}
	// Use a unique filename
	tempFile := filepath.Join(outputDir, fmt.Sprintf("go_sherpa_tts_%d.wav", time.Now().UnixNano()))

	if err := conn.WriteMessage(websocket.TextMessage, []byte(text)); err != nil {
		return "", fmt.Errorf("发送文本失败: %v", err)
	}
	
	_, bytes, err := conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("go-sherpa-tts 获取音频流失败: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("go-sherpa-tts 语音合成完成，耗时: %s\n", duration)

	// 将音频数据写入临时文件
	err = os.WriteFile(tempFile, bytes, 0644)
	if err != nil {
		return "", fmt.Errorf("写入音频文件 '%s' 失败: %v", tempFile, err)
	}

	return tempFile, nil
}
