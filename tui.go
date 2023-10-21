package main

import (
	"github.com/sashabaranov/go-openai"
)

var (
	MODELS = []string{openai.GPT3Dot5Turbo, openai.GPT4}
)

type model struct {
	choices []string
	cursor  int
}

func modelsSelectionModel() model {
	return model{
		choices: []string{openai.GPT3Dot5Turbo, openai.GPT4},
		cursor:  0,
	}
}

/*
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
