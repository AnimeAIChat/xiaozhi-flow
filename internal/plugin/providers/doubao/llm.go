package doubao

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

type LLMConfig struct {
	APIKey       string
	BaseURL      string
	Model        string
	MaxTokens    int
	ThinkingType string
}

type LLMProvider struct {
	config *LLMConfig
	client *http.Client
}

func NewLLMProvider(config *LLMConfig) *LLMProvider {
	if config.MaxTokens <= 0 {
		config.MaxTokens = 2048
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	return &LLMProvider{
		config: config,
		client: &http.Client{},
	}
}

type Message struct {
	Role       string
	Content    string
	ToolCallID string
	ToolCalls  []ToolCall
}

type ToolCall struct {
	ID       string
	Type     string
	Function ToolCallFunction
}

type ToolCallFunction struct {
	Name      string
	Arguments string
}

type Tool struct {
	Type     string
	Function ToolFunction
}

type ToolFunction struct {
	Name        string
	Description string
	Parameters  interface{}
}

type Response struct {
	Content   string
	ToolCalls []ToolCall
	Error     error
}

// doubaoRequest 自定义请求结构体,支持thinking参数
type doubaoRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	Stream      bool                     `json:"stream"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
	TopP        float64                  `json:"top_p,omitempty"`
	Tools       []openai.Tool            `json:"tools,omitempty"`
	Thinking    map[string]string        `json:"thinking,omitempty"` // 支持thinking参数
}

// doubaoStreamResponse SSE响应结构
type doubaoStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string            `json:"role,omitempty"`
			Content   string            `json:"content,omitempty"`
			ToolCalls []openai.ToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

func (p *LLMProvider) Chat(ctx context.Context, messages []Message, tools []Tool) (<-chan Response, error) {
	responseChan := make(chan Response, 10)

	go func() {
		defer close(responseChan)

		// 转换消息格式
		reqMessages := make([]map[string]interface{}, len(messages))
		for i, msg := range messages {
			msgMap := map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			}

			if msg.ToolCallID != "" {
				msgMap["tool_call_id"] = msg.ToolCallID
			}

			if len(msg.ToolCalls) > 0 {
				toolCalls := make([]map[string]interface{}, len(msg.ToolCalls))
				for j, tc := range msg.ToolCalls {
					toolCalls[j] = map[string]interface{}{
						"id":   tc.ID,
						"type": tc.Type,
						"function": map[string]interface{}{
							"name":      tc.Function.Name,
							"arguments": tc.Function.Arguments,
						},
					}
				}
				msgMap["tool_calls"] = toolCalls
			}

			reqMessages[i] = msgMap
		}

		// 转换工具格式
		var openaiTools []openai.Tool
		if len(tools) > 0 {
			openaiTools = make([]openai.Tool, len(tools))
			for i, tool := range tools {
				openaiTools[i] = openai.Tool{
					Type: openai.ToolType(tool.Type),
					Function: &openai.FunctionDefinition{
						Name:        tool.Function.Name,
						Description: tool.Function.Description,
						Parameters:  tool.Function.Parameters,
					},
				}
			}
		}

		// 构建自定义请求
		reqBody := doubaoRequest{
			Model:     p.config.Model,
			Messages:  reqMessages,
			Stream:    true,
			MaxTokens: p.config.MaxTokens,
		}
		
		if len(openaiTools) > 0 {
			reqBody.Tools = openaiTools
		}

		// 添加thinking参数
		if p.config.ThinkingType != "" && p.config.ThinkingType != "disabled" {
			reqBody.Thinking = map[string]string{
				"type": p.config.ThinkingType,
			}
		}

		// 序列化请求
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			responseChan <- Response{
				Content: fmt.Sprintf("【请求序列化失败: %v】", err),
				Error:   err,
			}
			return
		}

		// 创建HTTP请求
		url := fmt.Sprintf("%s/chat/completions", p.config.BaseURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			responseChan <- Response{
				Content: fmt.Sprintf("【创建请求失败: %v】", err),
				Error:   err,
			}
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))

		// 发送请求
		resp, err := p.client.Do(req)
		if err != nil {
			responseChan <- Response{
				Content: fmt.Sprintf("【Doubao服务响应异常: %v】", err),
				Error:   err,
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			responseChan <- Response{
				Content: fmt.Sprintf("【Doubao服务错误 %d: %s】", resp.StatusCode, string(body)),
				Error:   fmt.Errorf("HTTP %d", resp.StatusCode),
			}
			return
		}

		// 读取SSE流
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					responseChan <- Response{
						Content: fmt.Sprintf("【读取响应失败: %v】", err),
						Error:   err,
					}
				}
				break
			}

			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			// SSE格式: "data: {...}"
			if bytes.HasPrefix(line, []byte("data: ")) {
				data := bytes.TrimPrefix(line, []byte("data: "))
				
				// 检查是否是结束标记
				if string(data) == "[DONE]" {
					break
				}

				// 解析JSON
				var streamResp doubaoStreamResponse
				if err := json.Unmarshal(data, &streamResp); err != nil {
					continue
				}

				// 提取内容和工具调用
				if len(streamResp.Choices) > 0 {
					delta := streamResp.Choices[0].Delta

					// 处理工具调用
					if len(delta.ToolCalls) > 0 {
						toolCalls := make([]ToolCall, len(delta.ToolCalls))
						for i, tc := range delta.ToolCalls {
							toolCalls[i] = ToolCall{
								ID:   tc.ID,
								Type: string(tc.Type),
								Function: ToolCallFunction{
									Name:      tc.Function.Name,
									Arguments: tc.Function.Arguments,
								},
							}
						}
						responseChan <- Response{
							ToolCalls: toolCalls,
						}
						continue
					}

					// 处理文本内容
					if delta.Content != "" {
						responseChan <- Response{
							Content: delta.Content,
						}
					}
				}
			}
		}
	}()

	return responseChan, nil
}
