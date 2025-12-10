package gosherpa

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type ASRConfig struct {
	Cluster string
}

type ASRProvider struct {
	config     *ASRConfig
	conn       *websocket.Conn
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
	
	addr := p.config.Cluster
	if addr == "" {
		return fmt.Errorf("cluster address is required")
	}

	conn, _, err := dialer.DialContext(ctx, addr, nil)
	if err != nil {
		return err
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
		if p.conn != nil {
			p.conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if p.conn == nil {
				return
			}
			
			messageType, message, err := p.conn.ReadMessage()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				// Log error or send to output?
				// For now, just return as connection might be closed
				return
			}

			if messageType == websocket.TextMessage {
				text := string(message)
				if text != "" {
					p.outputChan <- map[string]interface{}{
						"text":     text,
						"is_final": true, // Gosherpa seems to send final results? Or partial? The old code treated it as final.
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
			
			if p.conn != nil {
				if err := p.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
					return
				}
			}
		}
	}
}
