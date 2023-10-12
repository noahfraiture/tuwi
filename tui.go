package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"
)

type modelSelection struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func (m modelSelection) View() string {
	// The header
	s := "What model do you want ?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func initModel() modelSelection {
	return modelSelection{
		choices:  []string{openai.GPT3Dot5Turbo, openai.GPT4},
		cursor:   0,
		selected: make(map[int]struct{}),
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
