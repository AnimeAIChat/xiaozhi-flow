package workflow

import (
	"time"
)

func CreateDefaultConversationWorkflow() *Workflow {
	return &Workflow{
		ID:          "conversation-v1",
		Name:        "Default Conversation",
		Description: "Standard ASR -> LLM -> TTS flow",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config: WorkflowConfig{
			Timeout:       60 * time.Second,
			MaxRetries:    0,
			ParallelLimit: 1,
		},
		Nodes: []Node{
			{
				ID:          "asr_node",
				Name:        "ASR",
				Type:        NodeTypeTask,
				Plugin:      "core.asr",
				Description: "Speech to Text",
				Inputs: []InputSchema{
					{Name: "audio_data", Type: "string", Required: true},
				},
				Outputs: []OutputSchema{
					{Name: "text", Type: "string"},
				},
				Position: Position{X: 100, Y: 100},
			},
			{
				ID:          "llm_node",
				Name:        "LLM",
				Type:        NodeTypeTask,
				Plugin:      "core.llm",
				Description: "Language Model",
				Inputs: []InputSchema{
					{Name: "text", Type: "string", Required: true},
				},
				Outputs: []OutputSchema{
					{Name: "text", Type: "string"},
				},
				Position: Position{X: 300, Y: 100},
			},
			{
				ID:          "tts_node",
				Name:        "TTS",
				Type:        NodeTypeTask,
				Plugin:      "core.tts",
				Description: "Text to Speech",
				Inputs: []InputSchema{
					{Name: "text", Type: "string", Required: true},
				},
				Outputs: []OutputSchema{
					{Name: "audio_data", Type: "string"},
				},
				Position: Position{X: 500, Y: 100},
			},
		},
		Edges: []Edge{
			{
				ID:   "edge_1",
				From: "asr_node",
				To:   "llm_node",
			},
			{
				ID:   "edge_2",
				From: "llm_node",
				To:   "tts_node",
			},
		},
	}
}
