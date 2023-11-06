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
	"os"
	"strings"
	"time"
)

const (
	AI     = "ai"
	CONV   = "conv"
	SYSTEM = "system"
	CHAT   = "chat"
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

		db          CouchDB
		docStyle    lipgloss.Style
		senderStyle lipgloss.Style
		err         error
		state       string
		quitting    bool
	}

	convModel struct {
		list   list.Model
		choice *Conversation
	}

	aiModel struct {
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
		conversation Conversation
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
		bucketName:       "test",
		scopeName:        "_default",
		collectionName:   "_default",
	}
	return model{
		conv:   initialConv(&db),
		ai:     initialAI(),
		system: initialSystem(),
		chat:   initialChat(),

		db: db,

		docStyle:    lipgloss.NewStyle().Margin(1, 2),
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		state:       CONV,
		quitting:    false,
	}

}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Update TODO : it seems that display two list one after the other doesn't work
// ai -> conv or conv -> ai is same shit
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		}
		break
	case tea.WindowSizeMsg:
		h, v := m.docStyle.GetFrameSize()
		m.conv.list.SetSize(msg.Width-h, msg.Height-v)
		m.ai.list.SetSize(msg.Width-h, msg.Height-v)
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
	default:
		return "State doesn't exist\n"
	}
}

func initialConv(db *CouchDB) convModel {
	if listConv, err := db.GetConversations(); err != nil {
		return convModel{
			list:   list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
			choice: nil,
		}
	} else {
		listItemConv := make([]list.Item, len(listConv))
		for i, conv := range listConv {
			listItemConv[i] = itemConv(conv)
		}
		return convModel{
			list:   list.New(listItemConv, list.NewDefaultDelegate(), 0, 0),
			choice: nil,
		}
	}
}

func (m model) viewConv() string {
	return m.docStyle.Render(m.conv.list.View())
}

func (m model) updateConv(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyEnter {
		if i, ok := m.conv.list.SelectedItem().(itemConv); ok {
			m.state = AI
			m.conv.choice = (*Conversation)(&i) // TODO : does the conversation loose data ?
		}
	}

	var cmd tea.Cmd
	m.conv.list, cmd = m.conv.list.Update(msg)
	return m, cmd
}

func initialAI() aiModel {
	return aiModel{
		list: list.New([]list.Item{
			aiVersion{title: "ai1",
				desc: "placeholder"},
			aiVersion{title: "ai2",
				desc: "placeholder"},
		}, list.NewDefaultDelegate(), 0, 0),
		choice: nil,
	}
}

func (m model) viewAI() string {
	return m.docStyle.Render(m.ai.list.View())
}

func (m model) updateAI(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyEnter {
		if i, ok := m.ai.list.SelectedItem().(aiVersion); ok {
			m.state = SYSTEM
			m.ai.choice = &i
		}
	}

	var cmd tea.Cmd
	m.ai.list, cmd = m.ai.list.Update(msg)
	return m, cmd
}

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
			m.system.content = m.system.texting.Value()
			m.state = CHAT
		}
	case error:
		m.err = msg
		return m, nil
	}
	m.system.texting, cmd = m.system.texting.Update(msg)
	return m, cmd
}

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
		conversation: Conversation{},
		viewport:     vp,
		textarea:     ta,
		messages:     []string{},
	}
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
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.chat.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.chat.messages = append(m.chat.messages, m.senderStyle.Render("You: ")+m.chat.textarea.Value())
			m.chat.viewport.SetContent(strings.Join(m.chat.messages, "\n"))
			m.chat.textarea.Reset()
			m.chat.viewport.GotoBottom()
		}

	// We handle errors just like any other message
	case error:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) viewChat() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.chat.viewport.View(),
		m.chat.textarea.View(),
	) + "\n\n"
}
