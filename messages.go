package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
)

const (
	roleUser   = "user"
	modelUser  = "userModel"
	finishUser = "userEnd"
)

type (
	Message struct {
		Role         string       `json:"role"` // TODO : create a type role for consistency ?
		Content      string       `json:"content"`
		FinishReason finishReason `json:"finish_reason"`
		Model        string       `json:"name"` // WARN : for now it will mix the models and company
	}
	Conversation struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		HasChange bool      `json:"has_change"`
		LastModel string    `json:"last_model"`
		Messages  []Message `json:"messages"`
	}

	userOpenaiMessage openai.ChatCompletionMessage
	gptMessage        openai.ChatCompletionChoice
	finishReason      string
)

func (m Message) render() string {
	senderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))

	var style lipgloss.Style
	var sender string
	switch m.Role {
	case roleUser:
		sender = "You :"
	case openai.ChatMessageRoleAssistant:
		sender = "AI :"
	case openai.ChatMessageRoleSystem:
		sender = "System :"
	}
	switch m.FinishReason {
	case finishUser:
		style = senderStyle
	case finishReason(openai.FinishReasonNull):
		style = redStyle
	case finishReason(openai.FinishReasonLength):
		style = blueStyle
	case finishReason(openai.FinishReasonStop):
		style = greenStyle
	}
	return style.Render(sender) + " " + m.Content
}

func (conv *Conversation) openaiMessages() []openai.ChatCompletionMessage {
	list := make([]openai.ChatCompletionMessage, len(conv.Messages))
	for i, message := range conv.Messages {
		list[i] = openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		}
	}
	return list
}

func (message userOpenaiMessage) toMessage() Message {
	return Message{
		Role:         message.Role,
		Content:      message.Content,
		FinishReason: "userEnd",
		Model:        "",
	}
}

func (message gptMessage) toMessage(model string) Message {
	return Message{
		Role:         message.Message.Role,
		Content:      message.Message.Content,
		FinishReason: finishReason(message.FinishReason),
		Model:        model,
	}
}
