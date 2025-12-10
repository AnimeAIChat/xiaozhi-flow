package coze

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/coze-dev/coze-go"
)

type LLMConfig struct {
	BaseURL     string
	BotID       string
	UserID      string
	AccessToken string
	ClientID    string
	PublicKey   string
	PrivateKey  string
}

type LLMProvider struct {
	config                 *LLMConfig
	client                 coze.CozeAPI
	sessionConversationMap sync.Map
}

type Message struct {
	Role    string
	Content string
}

func NewLLMProvider(config *LLMConfig) (*LLMProvider, error) {
	p := &LLMProvider{
		config: config,
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.coze.cn"
	}

	var authCli coze.Auth
	if config.ClientID != "" && config.PublicKey != "" && config.PrivateKey != "" {
		// JWT Auth
		client, err := coze.NewJWTOAuthClient(coze.NewJWTOAuthClientParam{
			ClientID:      config.ClientID,
			PublicKey:     config.PublicKey,
			PrivateKeyPEM: config.PrivateKey,
		}, coze.WithAuthBaseURL(baseURL))
		if err != nil {
			return nil, fmt.Errorf("Coze create JWT auth client failed: %v", err)
		}

		authCli = coze.NewJWTAuth(client, nil)
	} else {
		// Token Auth
		authCli = coze.NewTokenAuth(config.AccessToken)
	}
	p.client = coze.NewCozeAPI(authCli, coze.WithBaseURL(baseURL))

	return p, nil
}

func (p *LLMProvider) Chat(ctx context.Context, sessionID string, messages []Message) (<-chan string, error) {
	responseChan := make(chan string, 10)

	go func() {
		defer close(responseChan)

		var lastMsg string
		if len(messages) > 0 {
			lastMsg = messages[len(messages)-1].Content
		}

		conversationId, ok := p.sessionConversationMap.Load(sessionID)
		if !ok {
			conversation, err := p.client.Conversations.Create(ctx, &coze.CreateConversationsReq{
				Messages: []*coze.Message{},
			})
			if err != nil {
				responseChan <- fmt.Sprintf("【Coze create conversation failed: %v】", err)
				return
			}
			conversationId = conversation.ID
			p.sessionConversationMap.Store(sessionID, conversationId)
		}

		stream, err := p.client.Chat.Stream(ctx, &coze.CreateChatsReq{
			BotID:  p.config.BotID,
			UserID: p.config.UserID,
			Messages: []*coze.Message{
				coze.BuildUserQuestionObjects([]*coze.MessageObjectString{
					coze.NewTextMessageObject(lastMsg),
				}, nil),
			},
			ConversationID: conversationId.(string),
		})
		if err != nil {
			responseChan <- fmt.Sprintf("【Coze chat stream failed: %v】", err)
			return
		}
		defer stream.Close()

		for {
			event, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// Stream finished
				} else {
					responseChan <- fmt.Sprintf("【Coze stream error: %v】", err)
				}
				break
			}

			if event.Event == coze.ChatEventConversationMessageDelta {
				responseChan <- event.Message.Content
			}
		}
	}()

	return responseChan, nil
}
