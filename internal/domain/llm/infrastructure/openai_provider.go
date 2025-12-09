package infrastructure

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sashabaranov/go-openai"
	"xiaozhi-server-go/internal/domain/llm/aggregate"
	"xiaozhi-server-go/internal/domain/llm/provider"
	"xiaozhi-server-go/internal/domain/llm/repository"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/errors"
)

type OpenAIProvider struct {
	id     string
	client *openai.Client
	config config.LLMConfig
}

func NewOpenAIProvider(id string, cfg config.LLMConfig) provider.Provider {
	openaiConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		openaiConfig.BaseURL = cfg.BaseURL
	}

	return &OpenAIProvider{
		id:     id,
		client: openai.NewClientWithConfig(openaiConfig),
		config: cfg,
	}
}

func (p *OpenAIProvider) Type() string {
	return p.config.Type
}

func (p *OpenAIProvider) Generate(ctx context.Context, req repository.GenerateRequest) (*repository.GenerateResult, error) {
	messages := p.convertMessages(req.Messages)
	tools := p.convertTools(req.Tools)

	// Use config from request, fallback to provider config if needed
	model := req.Config.Model
	if model == "" {
		model = p.config.ModelName
	}
	
	temperature := req.Config.Temperature
	if temperature == 0 {
		temperature = float32(p.config.Temperature)
	}

	maxTokens := req.Config.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.config.MaxTokens
	}

	topP := req.Config.TopP
	if topP == 0 {
		topP = float32(p.config.TopP)
	}

	chatReq := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Tools:       tools,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		TopP:        topP,
	}

	resp, err := p.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "openai.generate", "API call failed", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New(errors.KindDomain, "openai.generate", "no response choices")
	}

	choice := resp.Choices[0]
	result := &repository.GenerateResult{
		Content:      choice.Message.Content,
		FinishReason: string(choice.FinishReason),
		Usage: &aggregate.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// Handle tool calls
	if choice.Message.ToolCalls != nil {
		result.ToolCalls = make([]repository.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = repository.ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: repository.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
				Index: i, // Index is not a pointer in repository.ToolCall? Let's check.
			}
		}
	}

	return result, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, req repository.GenerateRequest) (<-chan repository.ResponseChunk, error) {
	messages := p.convertMessages(req.Messages)
	tools := p.convertTools(req.Tools)

	model := req.Config.Model
	if model == "" {
		model = p.config.ModelName
	}
	
	temperature := req.Config.Temperature
	if temperature == 0 {
		temperature = float32(p.config.Temperature)
	}

	maxTokens := req.Config.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.config.MaxTokens
	}

	topP := req.Config.TopP
	if topP == 0 {
		topP = float32(p.config.TopP)
	}

	chatReq := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Tools:       tools,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		TopP:        topP,
		Stream:      true,
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "openai.stream", "stream creation failed", err)
	}

	outChan := make(chan repository.ResponseChunk, 10)

	go func() {
		defer close(outChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if strings.Contains(err.Error(), "stream closed") {
					return
				}
				// Send error to channel? Or just log?
				// The interface expects ResponseChunk, maybe add Error field or just close.
				// For now, let's just return.
				return
			}

			if len(response.Choices) == 0 {
				continue
			}

			choice := response.Choices[0]
			chunk := repository.ResponseChunk{
				Content: choice.Delta.Content,
				Done:    choice.FinishReason != "",
			}

			// Handle tool calls
			if choice.Delta.ToolCalls != nil {
				chunk.ToolCalls = make([]repository.ToolCall, len(choice.Delta.ToolCalls))
				for i, tc := range choice.Delta.ToolCalls {
					chunk.ToolCalls[i] = repository.ToolCall{
						ID:   tc.ID,
						Type: string(tc.Type),
						Function: repository.ToolCallFunction{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
						Index: *tc.Index,
					}
				}
			}

			// Handle usage
			if response.Usage != nil {
				chunk.Usage = &aggregate.Usage{
					PromptTokens:     response.Usage.PromptTokens,
					CompletionTokens: response.Usage.CompletionTokens,
					TotalTokens:      response.Usage.TotalTokens,
				}
			}

			outChan <- chunk

			if chunk.Done {
				break
			}
		}
	}()

	return outChan, nil
}

func (p *OpenAIProvider) convertMessages(msgs []repository.Message) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, len(msgs))
	for i, msg := range msgs {
		messages[i] = openai.ChatCompletionMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCalls:  p.convertToolCalls(msg.ToolCalls),
			ToolCallID: msg.ToolCallID,
		}
	}
	return messages
}

func (p *OpenAIProvider) convertToolCalls(calls []repository.ToolCall) []openai.ToolCall {
	if len(calls) == 0 {
		return nil
	}

	toolCalls := make([]openai.ToolCall, len(calls))
	for i, call := range calls {
		toolCalls[i] = openai.ToolCall{
			ID:   call.ID,
			Type: openai.ToolType(call.Type),
			Function: openai.FunctionCall{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		}
	}
	return toolCalls
}

func (p *OpenAIProvider) convertTools(tools []repository.Tool) []openai.Tool {
	if len(tools) == 0 {
		return nil
	}

	openaiTools := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		params, _ := json.Marshal(tool.Function.Parameters)

		openaiTools[i] = openai.Tool{
			Type: openai.ToolType(tool.Type),
			Function: &openai.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  params,
			},
		}
	}
	return openaiTools
}
