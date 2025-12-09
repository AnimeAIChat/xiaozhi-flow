package stepfun

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ASRConfig struct {
	APIKey string
	Model  string
	Voice  string
	Prompt string
}

type ASRProvider struct {
	config     *ASRConfig
	conn       *websocket.Conn
	connMutex  sync.Mutex
	outputChan chan<- map[string]interface{}
}

func NewASRProvider(config *ASRConfig, outputChan chan<- map[string]interface{}) *ASRProvider {
	return &ASRProvider{
		config:     config,
		outputChan: outputChan,
	}
}

func (p *ASRProvider) Start(ctx context.Context, audioStream <-chan []byte) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	model := p.config.Model
	if model == "" {
		model = "step-asr"
	}
	url := fmt.Sprintf("wss://api.stepfun.com/v1/realtime?model=%s", model)

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))

	conn, resp, err := dialer.DialContext(ctx, url, headers)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		return fmt.Errorf("WebSocket connection failed (status code:%d): %v", status, err)
	}
	p.conn = conn

	// Send session.update
	prompt := p.config.Prompt
	if prompt == "" {
		prompt = "你是由阶跃星辰提供的AI聊天助手，你擅长中文、英文及多语种对话。"
	}
	voice := p.config.Voice
	if voice == "" {
		voice = "cixing"
	}

	sessionPayload := map[string]interface{}{
		"event_id": fmt.Sprintf("event_%d", time.Now().UnixNano()),
		"type":     "session.update",
		"session": map[string]interface{}{
			"modalities":          []string{"text", "audio"},
			"instructions":        prompt,
			"voice":               voice,
			"input_audio_format":  "pcm16",
			"output_audio_format": "pcm16",
			"turn_detection": map[string]interface{}{
				"type":                       "server_vad",
				"energy_awakeness_threshold": 100,
			},
		},
	}

	if err := p.sendJSON(sessionPayload); err != nil {
		p.conn.Close()
		return fmt.Errorf("failed to send session.update: %v", err)
	}

	// Start reading response
	go p.readLoop(ctx)

	// Start sending audio
	go p.writeLoop(ctx, audioStream)

	return nil
}

func (p *ASRProvider) sendJSON(v interface{}) error {
	p.connMutex.Lock()
	defer p.connMutex.Unlock()

	if p.conn == nil {
		return fmt.Errorf("connection closed")
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return p.conn.WriteMessage(websocket.TextMessage, bytes)
}

func (p *ASRProvider) readLoop(ctx context.Context) {
	defer func() {
		p.connMutex.Lock()
		if p.conn != nil {
			p.conn.Close()
		}
		p.connMutex.Unlock()
	}()

	var baseEvent BaseEvent
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if p.conn == nil {
				return
			}
			
			p.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			msgType, data, err := p.conn.ReadMessage()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				p.sendError(fmt.Errorf("read error: %v", err))
				return
			}

			if msgType != websocket.TextMessage {
				continue
			}

			if err := json.Unmarshal(data, &baseEvent); err != nil {
				continue
			}

			switch baseEvent.Type {
			case "error":
				e := ErrorEvent{}
				if err := json.Unmarshal(data, &e); err == nil {
					p.sendError(fmt.Errorf("server error: %v", e.Error.Message))
				}
				return
			case "conversation.item.input_audio_transcription.completed":
				e := ConversationItemInputAudioTranscriptionCompletedEvent{}
				if err := json.Unmarshal(data, &e); err == nil {
					if e.Transcript != "" {
						p.outputChan <- map[string]interface{}{
							"text":     e.Transcript,
							"is_final": true,
						}
					}
				}
			}
		}
	}
}

func (p *ASRProvider) writeLoop(ctx context.Context, audioStream <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-audioStream:
			if !ok {
				return
			}
			
			encoded := base64.StdEncoding.EncodeToString(data)
			payload := map[string]interface{}{
				"event_id": fmt.Sprintf("event_%d", time.Now().UnixNano()),
				"type":     "input_audio_buffer.append",
				"audio":    encoded,
			}
			
			if err := p.sendJSON(payload); err != nil {
				p.sendError(fmt.Errorf("write error: %v", err))
				return
			}
		}
	}
}

func (p *ASRProvider) sendError(err error) {
	select {
	case p.outputChan <- map[string]interface{}{"error": err.Error()}:
	default:
	}
}
