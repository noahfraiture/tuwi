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
	AI      = "ai"
	CONV    = "conv"
	SYSTEM  = "system"
	CHAT    = "chat"
	SAVE    = "save"
	NEWCONV = "newconv"
)

// TODO : decide if conversation and DB are store here or in the model
// I prefer to keep the model simple and strict to the display
// but it generate global variable
// Also part of the conversation like message must be display

type (
	tickMsg struct{}

	// MAIN MODEL
	// TODO : could use factory to remove duplicate code ?
	model struct {
		conv   convModel
		ai     aiModel
		system systemModel
		chat   chatModel
		save   saveModel

		db          CouchDB
		senderStyle lipgloss.Style
		redStyle    lipgloss.Style
		greenStyle  lipgloss.Style
		yellowStyle lipgloss.Style
		blueStyle   lipgloss.Style
		err         error
		state       string
		quitting    bool
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
)

func (conv itemConv) Title() string {
	return conv.Name
}
func (conv itemConv) Description() string {
	return conv.ID
}
func (conv itemConv) FilterValue() string {
	return conv.Name
}

func (i aiVersion) Title() string       { return i.title }
func (i aiVersion) Description() string { return i.desc }
func (i aiVersion) FilterValue() string { return i.title }

// Run program
// TODO : add comments everywhere

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initialModel() model {
	db := CouchDB{
		username:         "admin",
		password:         "admin123",
		connectionString: "localhost",
		bucketName:       "conversations",
		scopeName:        "_default",
		collectionName:   "_default",
	}
	return model{
		conv:   initialConv(&db),
		ai:     initialAI(),
		system: initialSystem(),
		chat:   initialChat(),
		save:   initialSave(),

		db: db,

		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		redStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		greenStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		yellowStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		blueStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("4")), // TODO : is it blue ?
		state:       CONV,
		quitting:    false,
	}

}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlS:
			fmt.Println("Saving...")
		}
		break
	case tea.WindowSizeMsg:
		m.windowConv(msg)
		m.windowAI(msg)
		break
	}

	switch m.state {
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
		m.err = errors.New("State doesn't exist\n")
		println(m.err)
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

// CONVERSATION

func initialConv(db *CouchDB) convModel {
	conv := convModel{
		style:  lipgloss.NewStyle().Margin(1, 2),
		choice: nil,
	}
	if listConv, err := db.GetConversations(); err != nil {
		conv.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	} else {
		listItemConv := make([]list.Item, len(listConv)+1)
		listItemConv[0] = itemConv(Conversation{
			ID:        NEWCONV,
			Model:     "",
			Name:      "New conversation",
			Messages:  nil,
			HasChange: false,
		})
		for i, conv := range listConv {
			listItemConv[i+1] = itemConv(conv)
		}
		conv.list = list.New(listItemConv, list.NewDefaultDelegate(), 0, 0)
	}
	return conv
}

func (m model) viewConv() string {
	return m.conv.style.Render(m.conv.list.View()) + "\n"
}

func (m model) windowConv(msg tea.WindowSizeMsg) {
	h, v := m.conv.style.GetFrameSize()
	m.conv.list.SetSize(msg.Width-h, msg.Height-v)
}

func (m model) updateConv(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If the user choose the NEWCONV conversation, we create a new one
	// Else we replace the existing one by the chosen one and go directly to the chat

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
			if i, ok := m.conv.list.SelectedItem().(itemConv); ok {
				if i.ID == NEWCONV {
					m.state = AI
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
				} else {
					m.state = CHAT
					m.chat.conversation = (*Conversation)(&i)
				}
				m.conv.choice = (*Conversation)(&i) // TODO : does the conversation loose data ?
			}
		}
	}

	var cmd tea.Cmd
	m.conv.list, cmd = m.conv.list.Update(msg)
	return m, cmd
}

// AI

func initialAI() aiModel {
	return aiModel{
		list: list.New([]list.Item{
			aiVersion{title: openai.GPT4,
				desc: "placeholder"},
			aiVersion{title: openai.GPT3Dot5Turbo,
				desc: "placeholder"},
		}, list.NewDefaultDelegate(), 0, 0),
		style:  lipgloss.NewStyle().Margin(1, 2),
		choice: nil,
	}
}

func (m model) windowAI(msg tea.WindowSizeMsg) {
	h, v := m.ai.style.GetFrameSize()
	m.ai.list.SetSize(msg.Width-h, msg.Height-v)
}

func (m model) viewAI() string {
	return m.ai.style.Render(m.ai.list.View()) + "\n"
}

func (m model) updateAI(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyEnter {
		if i, ok := m.ai.list.SelectedItem().(aiVersion); ok {
			m.state = SYSTEM
			m.ai.choice = &i
			m.chat.conversation.Model = i.title
		}
	}

	var cmd tea.Cmd
	m.ai.list, cmd = m.ai.list.Update(msg)
	return m, cmd
}

// SYSTEM

func initialSystem() systemModel {
	it := textinput.New()
	it.Placeholder = "You are a helpful assistant"
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
			m.state = CHAT
			m.system.content = m.system.texting.Value()
			if m.system.content == "" {
				m.system.content = "You are a helpful assistant"
			}
			m.chat.conversation.Messages = append(m.chat.conversation.Messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: m.system.content,
			})
		}
	case error: // TODO : handle error in others function
		m.err = msg
		return m, nil
	}
	m.system.texting, cmd = m.system.texting.Update(msg)
	return m, cmd
}

// CHAT

func initialChat() chatModel {
	vp := viewport.New(30, 30) // TODO : adapt at size of the terminal
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
		"%s\n\n%s",
		m.chat.viewport.View(),
		m.chat.textarea.View(),
	) + "\n\n"
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
			m.chat.messages = append(m.chat.messages, m.senderStyle.Render("You: ")+m.chat.textarea.Value())

			finishReason, err := m.chat.conversation.ChatCompletion(m.chat.textarea.Value())
			if err != nil {
				m.err = err
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
			m.chat.messages = append(m.chat.messages, style.Render("Bot : ")+m.chat.conversation.Messages[len(m.chat.conversation.Messages)-1].Content)
			m.chat.viewport.SetContent(strings.Join(m.chat.messages, "\n"))
			m.chat.textarea.Reset()
			m.chat.viewport.GotoBottom()
		case tea.KeyCtrlS:
			m.state = SAVE
			if m.chat.conversation.Name != NEWCONV {
				m.save.texting.Placeholder = m.chat.conversation.Name
			}
			return m, nil
		}
	}
	// TODO : Handle error in type

	return m, tea.Batch(tiCmd, vpCmd)
}

// SAVE

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
			m.state = CONV
			// TODO : should retrieve conversations from DB
			m.save.content = m.save.texting.Value()
			if m.save.content != "" {
				m.chat.conversation.Name = m.save.content
			}
			err := m.db.StoreConversation(m.chat.conversation)
			if err != nil {
				m.err = err
			}
		}
	case error: // TODO : handle error in others function
		m.err = msg
		return m, nil
	}
	m.save.texting, cmd = m.save.texting.Update(msg)
	return m, cmd
}
