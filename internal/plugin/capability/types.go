package capability

import (
	"context"
)

// Type defines the category of the capability
type Type string

const (
	TypeLLM  Type = "llm"
	TypeASR  Type = "asr"
	TypeTTS  Type = "tts"
	TypeVAD  Type = "vad"
	TypeTool Type = "tool"
)

// Schema describes the data structure for config, inputs, or outputs
// This is a simplified JSON Schema representation
type Schema struct {
	Type       string              `json:"type"` // object, string, number, array, boolean
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string        `json:"type"`
	Description string        `json:"description,omitempty"`
	Default     interface{}   `json:"default,omitempty"`
	Enum        []interface{} `json:"enum,omitempty"`
	Items       *Schema       `json:"items,omitempty"`   // For arrays
	Secret      bool          `json:"secret,omitempty"`  // For sensitive config like API keys
}

// Definition describes what a capability does and what it needs
type Definition struct {
	ID          string `json:"id"`          // Unique ID, e.g., "openai_chat"
	Type        Type   `json:"type"`        // llm, asr, etc.
	Name        string `json:"name"`        // Human readable name
	Description string `json:"description"` 

	ConfigSchema Schema `json:"config_schema"` // Static config (API keys, model selection)
	InputSchema  Schema `json:"input_schema"`  // Runtime inputs (messages, audio bytes)
	OutputSchema Schema `json:"output_schema"` // Runtime outputs (text, audio bytes)
}

// Executor is the interface that must be implemented to run the capability
type Executor interface {
	// Execute runs the capability
	// config: The static configuration map
	// inputs: The runtime input map
	Execute(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (map[string]interface{}, error)
}

// StreamExecutor is an extension of Executor that supports streaming output
type StreamExecutor interface {
	Executor
	// ExecuteStream runs the capability and returns a channel of updates
	ExecuteStream(ctx context.Context, config map[string]interface{}, inputs map[string]interface{}) (<-chan map[string]interface{}, error)
}

// Provider is an interface for a plugin that provides multiple capabilities
type Provider interface {
	// GetCapabilities returns all capabilities provided by this plugin
	GetCapabilities() []Definition
	
	// CreateExecutor creates an executor for a specific capability
	CreateExecutor(capabilityID string) (Executor, error)
}
