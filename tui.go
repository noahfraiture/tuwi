package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
	"os"
	"time"
)

var (

	// TODO : decide if conversation and DB are store here or in the model
	// I prefer to keep the model simple and strict to the display
	// but it generate global variable
	// Also part of the conversation like message must be display
	db = CouchDB{
		username:         "admin",
		password:         "admin123",
		connectionString: "localhost",
		bucketName:       "conversations",
		scopeName:        "_default",
		collectionName:   "_default",
	}
	currentConversation Conversation

	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type (
	tickMsg  struct{}
	frameMsg struct{}
)

// Run program

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// MAIN MODEL
// TODO : could use factory to remove duplicate code ?

type model struct {
	mAI       modelAI
	mConv     modelConv
	mSystem   modelSystem
	mQuestion modelQuestion
	state     string
	quitting  bool
}

func initialModel() model {
	convModel, err := initialConversationModel()
	if err != nil {
		// TODO : handle error
		return model{}
	}
	return model{
		mAI:       initialModelAI(),
		mConv:     convModel,
		mSystem:   initialSystemModel(),
		mQuestion: initialQuestionModel(),
		state:     "ai",
		quitting:  false,
	}

}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok { // TODO : change that shit
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.state {
	case "conv":
		tmpModel, tmpCmd := m.mConv.Update(msg)
		if m.mConv.choice != nil {
			m.state = "ai"
		}
		return tmpModel, tmpCmd
	case "ai": // choose the model of AI for new conversation
		tmpModel, tmpCmd := m.mAI.Update(msg)
		if m.mAI.choice != nil {
			m.state = "system"
		}
		return tmpModel, tmpCmd
	case "system":
		return m.mSystem.Update(msg)
	case "question":
		return m.mQuestion.Update(msg)
	default:
		// TODO : handle error
		return nil, nil
	}
}

func (m model) View() string {
	if m.quitting {
		return "\n See ya  !\n\n"
	}
	switch m.state {
	case "ai":
		return m.mAI.View()
	case "conv":
		return m.mConv.View()
	case "system":
		return m.mSystem.View()
	case "question":
		return m.mQuestion.View()
	default:
		return "Error happened, "
	}
}

// Model ond method for CONVERSATION CHOICE

type item struct {
	title, desc string
}

type modelConv struct {
	list   list.Model
	choice *item
}

func initialConversationModel() (modelConv, error) {
	ids, err := db.GetDocumentsID()
	if err != nil {
		return modelConv{}, err
	}
	idsItem := make([]list.Item, 2)
	for _, id := range ids {
		idsItem = append(idsItem, item{
			title: id,
			desc:  "",
		})
	}
	return modelConv{list: list.New(idsItem, list.NewDefaultDelegate(), 0, 0)}, nil
}

func (m modelConv) Init() tea.Cmd {
	return nil
}

func (m modelConv) View() string {
	return docStyle.Render(m.list.View())
}

func (m modelConv) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = &i
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// Model and method for the AI selection
// TODO : really duplicate with conversation section, is there a necessity to split them ?

func (i item) FilterValue() string {
	return i.title
}

type modelAI struct {
	list   list.Model
	choice *item
}

func initialModelAI() modelAI {
	aiModels := []list.Item{
		item{title: openai.GPT3Dot5Turbo, desc: "price placeholder"},
		item{title: openai.GPT4, desc: "gpt4 placeholder"},
	}
	return modelAI{
		list: list.New(aiModels, list.NewDefaultDelegate(), 0, 0),
	}
}

func (m modelAI) Init() tea.Cmd {
	return nil
}

func (m modelAI) View() string {
	return docStyle.Render(m.list.View())
}

func (m modelAI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = &i
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit // TODO : quitting bool ? Here or will be detected in main ? Should never be reached
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// Model and method for SYSTEM MESSAGE

type modelSystem struct {
	textInput textinput.Model
	content   string
	err       error
}

func initialSystemModel() modelSystem {
	ti := textinput.New()
	ti.Placeholder = "Your are a helpful assistant."
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 100
	return modelSystem{
		textInput: ti,
	}
}

func (m modelSystem) Init() tea.Cmd {
	return textinput.Blink
}

func (m modelSystem) View() string {
	return fmt.Sprintf(
		"Enter system message \n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func (m modelSystem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.content = m.textInput.Value() // TODO : right command ?
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, nil
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// Model and method for QUESTION

type modelQuestion struct {
	textInput textinput.Model
	content   string
	err       error
}

func initialQuestionModel() modelQuestion {
	ti := textinput.New()
	ti.Placeholder = "How do we cook meth ?"
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 100
	return modelQuestion{
		textInput: ti,
	}
}

func (m modelQuestion) Init() tea.Cmd {
	return textinput.Blink
}

func (m modelQuestion) View() string {
	return fmt.Sprintf(
		"Enter your question \n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func (m modelQuestion) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.content = m.textInput.Value()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, nil
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}
