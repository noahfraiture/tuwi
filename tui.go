package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/text/language/display"
	"io"
	"time"
)

var (

	// TODO : decide if conversation and DB are store here or in the model
	// I prefer to keep the model simple and strict to the display
	// but it generate global variable
	// Also part of the conversation like message must be display
	db = couchDB{
		username: "admin",
		password: "admin123",
		connectionString: "localhost",
		bucketName: "conversations",
		scopeName: "_default",
		collectionName: "_default",
	}
	currentConversation conversation

	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type (
	tickMsg struct{}
	frameMsg struct{}
)

// MAIN SECTION

type model struct {
	mList modelList
	mInput modelInput
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}


// Model and method for the AI selection

type item struct {
	title, desc string
}

func (i item) FilterValue() string {
	return i.title
}

type modelList struct {
	list list.Model
}

func initialModelAI() modelList {
	aiModels := []list.Item{
		item{title: openai.GPT3Dot5Turbo, desc: "price placeholder"},
		item{title: openai.GPT4, desc: "gpt4 placeholder"},
	}
	return modelList{
		list: list.New(aiModels, list.NewDefaultDelegate(), 0, 0),
	}
}

func (m modelList) Init() tea.Cmd {
	return nil
}

func (m modelList) View() string {
	return docStyle.Render(m.list.View())
}

func (m modelList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
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


// Model ond method for CONVERSATION LOAD


func initialConversationtModel() (modelList, error) {
	ids, err := db.getDocumentsID()
	if err != nil {
		return modelList{}, err
	}
	idsItem := make([]list.Item, 2)
	for _, id := range ids {
		idsItem = append(idsItem, item{
			title: id,
			desc:  "",
		})
	}
	return modelList{list:
		list.New(idsItem, list.NewDefaultDelegate(), 0, 0)
	}, nil
}

// Model and method for SYSTEM MESSAGE

type modelInput struct {
	textInput textinput.Model
	err error
}

func (m modelInput) View() string {
	return fmt.Sprintf(
		"Enter what you must enter\n\n%s\n\n%s",
		// TODO split model and function for different message ?
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func (m modelInput) Init() tea.Cmd {
	return textinput.Blink
}

func initialSystemModel() modelInput {
	ti := textinput.New()
	ti.Placeholder = "How do you do ?"
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 100
	return modelInput{
		textInput: ti,
	}
}

func (m modelInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, nil
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}
