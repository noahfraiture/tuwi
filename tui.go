package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	KEY     = "key"
	AI      = "ai"
	CONV    = "conv"
	SYSTEM  = "system"
	CHAT    = "chat"
	SAVE    = "save"
	START   = "start"
	NEWCONV = "new-conv"
)

type (
	tickMsg struct{}

	// MAIN MODEL
	// TODO : could use factory to remove duplicate code ?
	model struct {
		key    keyModel
		conv   convModel
		ai     aiModel
		system systemModel
		chat   chatModel
		save   saveModel

		conversations Conversations

		renderStyle lipgloss.Style
		senderStyle lipgloss.Style
		redStyle    lipgloss.Style
		greenStyle  lipgloss.Style
		yellowStyle lipgloss.Style
		blueStyle   lipgloss.Style
		err         []error
		state       string
		quitting    bool
	}

	keyModel struct {
		texting textinput.Model
		content string
	}

	convModel struct {
		style  lipgloss.Style
		list   list.Model
		choice *Conversation
	}

	aiModel struct {
		style  lipgloss.Style
		list   list.Model
		choice *aiVersion
	}

	systemModel struct {
		texting textinput.Model
		content string
	}

	chatModel struct {
		viewport     viewport.Model
		textarea     textarea.Model
		messages     []string
		conversation *Conversation
	}

	saveModel struct {
		texting textinput.Model
		content string
	}

	aiVersion struct {
		title, desc string
	}
	itemConv Conversation

	// TODO : factory for initial model and interface for subModel
	// the problem is that methods like updateConv() have effect on other fields than conv, like chat
	// Thus I can not properly use interface and methods for convModel
)

func (conv itemConv) Title() string {
	return conv.Name
}
func (conv itemConv) Description() string {
	return conv.Model
}
func (conv itemConv) FilterValue() string {
	return conv.Name
}

func (i aiVersion) Title() string       { return i.title }
func (i aiVersion) Description() string { return i.desc }
func (i aiVersion) FilterValue() string { return i.title }

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initialModel() model {
	return model{
		key:    initialKey(),
		conv:   initialConv(),
		ai:     initialAI(),
		system: initialSystem(),
		chat:   initialChat(),
		save:   initialSave(),

		conversations: Conversations{},

		renderStyle: lipgloss.NewStyle().Margin(1, 2),
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		redStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		greenStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		yellowStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		blueStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		state:       START,
		quitting:    false,
	}

}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO : if user has no key, start as key
	if m.state == START {
		m = m.switchToConv()
	}

	if m.err != nil {
		for _, err := range m.err {
			fmt.Println(err)
		}
		m.err = nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		}
		break
	case tea.WindowSizeMsg:
		// Resizing can't be done in other function or should return a value
		// since I can't access through a pointer to the list
		h, v := m.renderStyle.GetFrameSize()
		m.conv.list.SetSize(msg.Width-h, msg.Height-v)
		m.ai.list.SetSize(msg.Width-h, msg.Height-v)
		break
	case error:
		m.err = append(m.err, msg)
	}

	switch m.state {
	case KEY:
		return m.updateKey(msg)
	case CONV:
		return m.updateConv(msg)
	case AI:
		return m.updateAI(msg)
	case SYSTEM:
		return m.updateSystem(msg)
	case CHAT:
		return m.updateChat(msg)
	case SAVE:
		return m.updateSave(msg)
	default:
		m.err = append(m.err, errors.New("State doesn't exist\n"))
		return m, tea.Quit
	}
}

func (m model) View() string {
	if m.quitting {
		return "\n See ya  !\n\n"
	}
	switch m.state {
	case CONV:
		return m.viewConv()
	case AI:
		return m.viewAI()
	case SYSTEM:
		return m.viewSystem()
	case CHAT:
		return m.viewChat()
	case SAVE:
		return m.viewSave()
	default:
		return "State doesn't exist\n"
	}
}

// KEY - View to ask the key to the user if he never entered one

func initialKey() keyModel {
	it := textinput.New()
	it.Placeholder = "sk-xxxxxxxx\n"
	it.CharLimit = 156
	it.Width = 20
	it.Focus() // TODO : has the order any importance ?
	return keyModel{
		texting: it,
		content: "",
	}
}

func (m model) viewKey() string {
	return fmt.Sprintf(
		"Enter your key \n\n%s\n\n%s",
		m.key.texting.View(),
		"(esc to quit)",
	) + "\n"
}

func (m model) updateKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.key.content = m.key.texting.Value()
			if m.key.content == "" {
				// TODO : warning the user has no input. I could parse the input and display a cross icon until it's valid
			}
			// TODO : store key
			m = m.switchToConv()
		}
	}
	m.key.texting, cmd = m.key.texting.Update(msg)
	return m, cmd
}

func (m model) switchToKey() model {
	m.state = KEY
	m.key.content = ""
	m.key.texting.Reset() // TODO : is it necessary ?
	return m
}

// CONVERSATION - View to choose the conversation. List conversations from db (-> CHAT) + "New conversation" (-> AI)

func initialConv() convModel {
	conv := convModel{
		style:  lipgloss.NewStyle().Margin(1, 2),
		list:   list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		choice: nil,
	}
	return conv
}

func (m model) viewConv() string {
	return m.conv.style.Render(m.conv.list.View()) + "\n"
}

func (m model) updateConv(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If the user choose the NEWCONV conversation, we create a new one
	// Else we replace the existing one by the chosen one and go directly to the chat

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
			if i, ok := m.conv.list.SelectedItem().(itemConv); ok {
				if i.ID == NEWCONV {
					randomBytes := make([]rune, 8)
					var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
					for i := range randomBytes {
						randomBytes[i] = letterRunes[rand.Intn(len(letterRunes))]
					}
					m.chat.conversation = &Conversation{
						ID:        string(randomBytes),
						Model:     "",
						Name:      "",
						Messages:  nil,
						HasChange: false,
					}
					// TODO choice : if we go back here, will it reset the conversation ? If yes it must be here
					m = m.switchToAI()
				} else {
					m.chat.conversation = (*Conversation)(&i)
					m = m.switchToChat()
				}
				m.conv.choice = (*Conversation)(&i)
			}
		}
	}

	var cmd tea.Cmd
	m.conv.list, cmd = m.conv.list.Update(msg)
	return m, cmd
}

func (m model) switchToConv() model {
	m.state = CONV
	m.conv.choice = nil

	err := m.conversations.UpdateConversations()
	if err != nil {
		m.err = append(m.err, err)
	}
	listItemConv := make([]list.Item, len(m.conversations)+1)
	listItemConv[0] = itemConv(Conversation{
		ID:        NEWCONV,
		Model:     "Choose your model",
		Name:      "New conversation",
		Messages:  nil,
		HasChange: false,
	})
	var i = 1
	for _, conv := range m.conversations {
		listItemConv[i] = itemConv(conv)
		i++
	}
	m.conv.list = list.New(listItemConv, list.NewDefaultDelegate(), 0, 0)
	return m
}

// AI - View to choose the AI. List AI from openAI. -> System. CTRL+Z -> Conversation

func initialAI() aiModel {
	return aiModel{
		list: list.New([]list.Item{
			aiVersion{title: openai.GPT4, // TODO reformat this like said in jimmy's message
				desc: "PRICE"}, // TODO PRICE
			aiVersion{title: openai.GPT3Dot5Turbo,
				desc: "PRICE"},
		}, list.NewDefaultDelegate(), 0, 0),
		style:  lipgloss.NewStyle().Margin(1, 2),
		choice: nil,
	}
}

func (m model) viewAI() string {
	return m.ai.style.Render(m.ai.list.View()) + "\n"
}

func (m model) updateAI(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
			if i, ok := m.ai.list.SelectedItem().(aiVersion); ok {
				m.ai.choice = &i
				m.chat.conversation.Model = i.title
				m = m.switchToSystem()
			}
		case tea.KeyCtrlZ:
			m = m.switchToConv()
		}
	}

	var cmd tea.Cmd
	m.ai.list, cmd = m.ai.list.Update(msg)
	return m, cmd
}

func (m model) switchToAI() model {
	m.state = AI
	m.ai.choice = nil
	return m
}

// SYSTEM - View to choose the system message. -> Chat. CTRL+Z -> AI

func initialSystem() systemModel {
	it := textinput.New()
	it.Placeholder = "You are a helpful assistant\n"
	it.CharLimit = 156
	it.Width = 20
	it.Focus() // TODO : has the order any importance ?
	return systemModel{
		texting: it,
		content: "",
	}
}

func (m model) viewSystem() string {
	return fmt.Sprintf(
		"Enter system message \n\n%s\n\n%s",
		m.system.texting.View(),
		"(esc to quit)",
	) + "\n"
}

func (m model) updateSystem(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.system.content = m.system.texting.Value()
			if m.system.content == "" {
				m.system.content = "You are a helpful assistant\n"
			}
			m.chat.conversation.Messages = append(m.chat.conversation.Messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: m.system.content,
			})
			m = m.switchToChat()
		case tea.KeyCtrlZ:
			m = m.switchToAI()
		}
	}
	m.system.texting, cmd = m.system.texting.Update(msg)
	return m, cmd
}

func (m model) switchToSystem() model {
	m.state = SYSTEM
	m.system.content = ""
	m.system.texting.Reset() // TODO : is it necessary ?
	return m
}

// CHAT - View to chat with the AI. CTRL+S -> Save. CTRL+Z -> System

func initialChat() chatModel {
	vp := viewport.New(100, 30) // TODO : adapt at size of the terminal
	vp.SetContent(`Welcome to the chat room! Type a message and press Enter to send.`)

	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.SetWidth(30)
	ta.SetHeight(3)

	return chatModel{
		conversation: nil,
		viewport:     vp,
		textarea:     ta,
		messages:     []string{},
	}
}

func (m model) viewChat() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n",
		m.chat.viewport.View(),
		m.chat.textarea.View(),
	)
}

func (m model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.chat.textarea, tiCmd = m.chat.textarea.Update(msg)
	m.chat.viewport, vpCmd = m.chat.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:

			// NOTE : We don't have to add a newline since it's already done by the textarea
			currentMessage := m.chat.textarea.Value()
			m.chat.messages = append(m.chat.messages, m.senderStyle.Render("You: ")+currentMessage)

			finishReason, err := m.chat.conversation.ChatCompletion(m.chat.textarea.Value())
			if err != nil {
				m.err = append(m.err, err)
				return m, nil
			}
			var style lipgloss.Style
			switch finishReason {
			case openai.FinishReasonStop:
				style = m.greenStyle
			case openai.FinishReasonLength:
				style = m.yellowStyle
			case openai.FinishReasonContentFilter:
				style = m.redStyle
			default:
				style = m.blueStyle
			}

			// NOTE : Here we do have to add the newline since the response of openAI have no newline
			botMessage := m.chat.conversation.Messages[len(m.chat.conversation.Messages)-1].Content + "\n"
			m.chat.messages = append(m.chat.messages, style.Render("Bot : ")+botMessage)

			// TODO :  We reload the entiere conversation, it's simpler but could be optimized
			m.chat.viewport.SetContent(strings.Join(m.chat.messages, "\n"))
			m.chat.textarea.Reset()
			m.chat.viewport.GotoBottom()
		case tea.KeyCtrlS:
			if m.chat.conversation.Name != NEWCONV {
				m.save.texting.Placeholder = m.chat.conversation.Name
			}
			m = m.switchToSave()
			return m, nil
		case tea.KeyCtrlZ:
			m = m.switchToSystem()
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) switchToChat() model {
	m.state = CHAT
	m.chat.messages = []string{}
	for _, msg := range m.chat.conversation.Messages {
		var style lipgloss.Style
		var user string
		switch msg.Role {
		// TODO : finishReason is missing from the struct, so the color is not right
		case openai.ChatMessageRoleSystem:
			style = m.greenStyle
			user = "System: "
		case openai.ChatMessageRoleUser:
			style = m.senderStyle
			user = "You: "
		case openai.ChatMessageRoleAssistant:
			style = m.blueStyle
			user = "Bot: " // TODO : the color here and in the update are not coherent

		}
		m.chat.messages = append(m.chat.messages, style.Render(user)+msg.Content)
	}
	m.chat.viewport.SetContent(strings.Join(m.chat.messages, "\n"))
	return m
}

// SAVE - View to save the conversation. -> Conversation

func initialSave() saveModel {
	it := textinput.New()
	it.Placeholder = "Name of the conversation..."
	it.CharLimit = 32
	it.Width = 20
	it.Focus()
	return saveModel{
		texting: it,
		content: "",
	}
}

func (m model) viewSave() string {
	return fmt.Sprintf(
		"Enter system message \n\n%s\n\n%s",
		m.save.texting.View(),
		"(esc to quit)",
	) + "\n"
}

func (m model) updateSave(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.save.content = m.save.texting.Value()
			if m.save.content != "" {
				m.chat.conversation.Name = m.save.content
			}
			err := m.chat.conversation.SaveConversation()
			if err != nil {
				m.err = append(m.err, err)
			}
			m = m.switchToConv()
		case tea.KeyCtrlZ:
			m = m.switchToChat() // TODO : test
		}
	}
	m.save.texting, cmd = m.save.texting.Update(msg)
	return m, cmd
}

func (m model) switchToSave() model {
	m.state = SAVE
	m.save.content = ""
	m.save.texting.Reset() // TODO : is it necessary ?
	return m
}
