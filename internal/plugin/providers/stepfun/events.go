package stepfun

// BaseEvent 公共事件字段
type BaseEvent struct {
	EventID string `json:"event_id,omitempty"`
	Type    string `json:"type"`
}

// Error 事件
type ErrorDetail struct {
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	EventID string `json:"event_id,omitempty"`
}

type ErrorEvent struct {
	BaseEvent
	Error ErrorDetail `json:"error"`
}

// Session 会话对象
type Session struct {
	ID                      string   `json:"id,omitempty"`
	Object                  string   `json:"object,omitempty"`
	Model                   string   `json:"model,omitempty"`
	Modalities              []string `json:"modalities,omitempty"`
	Instructions            string   `json:"instructions,omitempty"`
	Voice                   string   `json:"voice,omitempty"`
	InputAudioFormat        string   `json:"input_audio_format,omitempty"`
	OutputAudioFormat       string   `json:"output_audio_format,omitempty"`
	MaxResponseOutputTokens string   `json:"max_response_output_tokens,omitempty"`
}

// Session 相关事件
type SessionCreatedEvent struct {
	BaseEvent
	Session Session `json:"session"`
}

type ConversationItemInputAudioTranscriptionCompletedEvent struct {
	BaseEvent
	ItemID      string `json:"item_id"`
	ContentIndex int    `json:"content_index"`
	Transcript  string `json:"transcript"`
}
