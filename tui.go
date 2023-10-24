package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"
)

var (
	MODELS = []string{openai.GPT3Dot5Turbo, openai.GPT4}

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
)

type model struct {
	choices        []string
	cursor         int
	currentMessage string
	lastResponse   string
	aiModel        string
	systemMessage  string
	convID         string
}

func modelsSelectionModel() model {
	return model{
		choices: MODELS,
		cursor:  0,
	}
}

func (m model) Init() tea.Cmd { return nil }

/
func historySelectionModel() model {
	// TODO : create and handle DB
	return model{
		choices: loadHistory(),
		cursor:  0,
	}
}

func (m modelSelection) Init() tea.Cmd {
	return nil
}

func (m modelSelection) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}
	return m, nil
}
*/
