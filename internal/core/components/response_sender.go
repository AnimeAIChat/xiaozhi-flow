package components

import (
	"encoding/json"
	"fmt"
	"xiaozhi-server-go/internal/platform/logging"
	"xiaozhi-server-go/internal/utils"
)

// MessageWriter defines the interface for sending raw messages
type MessageWriter interface {
	WriteMessage(messageType int, data []byte) error
}

// ResponseSender handles formatting and sending messages to the client
type ResponseSender struct {
	conn      MessageWriter
	logger    *logging.Logger
	sessionID string
}

// NewResponseSender creates a new ResponseSender
func NewResponseSender(conn MessageWriter, logger *logging.Logger, sessionID string) *ResponseSender {
	return &ResponseSender{
		conn:      conn,
		logger:    logger,
		sessionID: sessionID,
	}
}

// SendHello sends the initial hello message
func (s *ResponseSender) SendHello(version int, transport string, audioParams map[string]interface{}) error {
	hello := make(map[string]interface{})
	hello["type"] = "hello"
	hello["version"] = version
	hello["transport"] = transport
	hello["session_id"] = s.sessionID
	hello["audio_params"] = audioParams

	data, err := json.Marshal(hello)
	if err != nil {
		return fmt.Errorf("failed to marshal hello message: %v", err)
	}

	return s.conn.WriteMessage(1, data)
}

// SendTTSState sends TTS state updates (start, stop, etc.)
func (s *ResponseSender) SendTTSState(state string, text string, textIndex int) error {
	stateMsg := map[string]interface{}{
		"type":        "tts",
		"state":       state,
		"session_id":  s.sessionID,
		"text":        text,
		"index":       textIndex,
		"audio_codec": "opus",
	}

	data, err := json.Marshal(stateMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal %s state: %v", state, err)
	}

	if err := s.conn.WriteMessage(1, data); err != nil {
		return fmt.Errorf("failed to send %s state: %v", state, err)
	}
	return nil
}

// SendSTT sends Speech-to-Text results
func (s *ResponseSender) SendSTT(text string) error {
	sttMsg := map[string]interface{}{
		"type":       "stt",
		"text":       text,
		"session_id": s.sessionID,
	}

	jsonData, err := json.Marshal(sttMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal STT message: %v", err)
	}

	if err := s.conn.WriteMessage(1, jsonData); err != nil {
		return fmt.Errorf("failed to send STT message: %v", err)
	}
	return nil
}

// SendEmotion sends emotion updates
func (s *ResponseSender) SendEmotion(emotion string) error {
	data := map[string]interface{}{
		"type":       "llm",
		"text":       utils.GetEmotionEmoji(emotion),
		"emotion":    emotion,
		"session_id": s.sessionID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal emotion message: %v", err)
	}

	return s.conn.WriteMessage(1, jsonData)
}

// SendAudioFrame sends a single audio frame
func (s *ResponseSender) SendAudioFrame(data []byte) error {
	return s.conn.WriteMessage(2, data)
}

// SendRawText sends raw text message
func (s *ResponseSender) SendRawText(text string) error {
	return s.conn.WriteMessage(1, []byte(text))
}

// SendAudio sends audio data
func (s *ResponseSender) SendAudio(data []byte) error {
	return s.conn.WriteMessage(2, data)
}
