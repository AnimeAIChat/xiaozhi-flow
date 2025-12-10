package deepgram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ASRConfig struct {
	APIKey   string
	Language string
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

	// Add query parameters
	lang := p.config.Language
	if lang == "" {
		lang = "en"
	}
	queryParams := fmt.Sprintf("?language=%s&sample_rate=%v&encoding=%v",
		lang, 16000, "linear16")

	headers := http.Header{
		"Authorization": []string{"token " + p.config.APIKey},
	}

	conn, resp, err := dialer.DialContext(ctx, "wss://api.deepgram.com/v1/listen"+queryParams, headers)
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		return fmt.Errorf("WebSocket connection failed (status code:%d): %v", statusCode, err)
	}

	p.conn = conn

	// Start reading response
	go p.readLoop(ctx)

	// Start sending audio
	go p.writeLoop(ctx, audioStream)

	return nil
}

func (p *ASRProvider) readLoop(ctx context.Context) {
	defer func() {
		p.connMutex.Lock()
		if p.conn != nil {
			p.conn.Close()
		}
		p.connMutex.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if p.conn == nil {
				return
			}
			
			p.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			_, message, err := p.conn.ReadMessage()
			if err != nil {
				// If context is done, this is expected
				if ctx.Err() != nil {
					return
				}
				p.sendError(fmt.Errorf("read error: %v", err))
				return
			}

			var response map[string]interface{}
			if err := json.Unmarshal(message, &response); err != nil {
				continue
			}

			// Handle error response
			if resultType, ok := response["type"].(string); ok && resultType == "Error" {
				description := "unknown error"
				if desc, ok := response["description"].(string); ok {
					description = desc
				}
				p.sendError(fmt.Errorf("Deepgram API error: %s", description))
				return
			}

			// Handle successful transcription
			if resultType, ok := response["type"].(string); ok && resultType == "Results" {
				isFinal, _ := response["is_final"].(bool)

				if channel, ok := response["channel"].(map[string]interface{}); ok {
					if alternatives, ok := channel["alternatives"].([]interface{}); ok && len(alternatives) > 0 {
						if firstAlt, ok := alternatives[0].(map[string]interface{}); ok {
							if transcript, ok := firstAlt["transcript"].(string); ok {
								transcript = strings.TrimSpace(transcript)
								if transcript != "" {
									p.outputChan <- map[string]interface{}{
										"text":     transcript,
										"is_final": isFinal,
									}
								}
							}
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
				// Stream closed
				return
			}
			
			p.connMutex.Lock()
			if p.conn != nil {
				if err := p.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
					p.connMutex.Unlock()
					p.sendError(fmt.Errorf("write error: %v", err))
					return
				}
			}
			p.connMutex.Unlock()
		}
	}
}

func (p *ASRProvider) sendError(err error) {
	select {
	case p.outputChan <- map[string]interface{}{"error": err.Error()}:
	default:
	}
}
