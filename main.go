package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	input "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	te "github.com/muesli/termenv"
)

type Action int

const (
	Add Action = iota
	Edit
	Delete
	None
)

type model struct {
	choices     []string
	cursor      int
	selected    map[int]struct{}
	beingEdited int
	action      Action
	textInput   input.Model
}

func initialModel() model {
	inputModel := input.NewModel()
	inputModel.Placeholder = "What would you like to do?"
	inputModel.Focus()
	inputModel.CharLimit = 156
	inputModel.Width = 40
	inputModel.Prompt = colorFg(te.String("> "), "2").Bold().String()

	return model{
		choices:   []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
		selected:  make(map[int]struct{}),
		cursor:    0,
		textInput: inputModel,
		action:    None,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func clearAndHideInput(m model) model {
	m.textInput.SetValue("")
	m.action = None

	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.action == Delete {
			switch msg.String() {
			case "y", "Y":
				if len(m.choices) > 1 {
					m.choices = append(m.choices[:m.cursor], m.choices[m.cursor+1:]...)
				} else {
					m.choices = nil
				}
			}

			m.action = None
		} else if m.action == Add || m.action == Edit {
			switch msg.String() {
			case "esc", "ctrl-c":
				m = clearAndHideInput(m)
			case "enter":
				value := m.textInput.Value()

				if len(value) > 0 {
					if m.action == Add {
						m.choices = append(m.choices, value)
					} else if m.action == Edit {
						m.choices[m.cursor] = value
					}
				}

				m = clearAndHideInput(m)
			}

			m.textInput, cmd = m.textInput.Update(msg)
		} else {
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

			case "o", "a":
				m.action = Add

			case "i", "e":
				m.action = Edit
				m.textInput.SetValue(m.choices[m.cursor])
				m.textInput.CursorEnd()

			case "d":
				m.action = Delete
				if len(m.choices) > 1 {
					m.action = Delete
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

	}

	return m, cmd
}

func colorFg(val te.Style, color string) te.Style {
	return val.Foreground(te.ColorProfile().Color(color))
}

func colorBg(val te.Style, color string) te.Style {
	return val.Background(te.ColorProfile().Color(color))
}

func (m model) View() string {
	title := te.String(" Todo List: \n")
	title = colorFg(title, "15")
	title = colorBg(title, "8")
	title = title.Bold()
	s := title.String()

	if len(m.choices) <= 0 {
		s += "The list is empty!"
	} else {
		for i, choice := range m.choices {
			if m.action == Edit && m.cursor == i {
				s += m.textInput.View() + "\n"
			} else {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}

				checked := " "
				if _, ok := m.selected[i]; ok {
					checked = "x"
					choice = te.String(choice).Faint().String()
				}

				s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
			}
		}
	}

	switch m.action {
	case Add:
		s += "\n" + m.textInput.View()
	case Delete:
		s += colorFg(te.String("\nDelete todo? press 'y' to confirm or any other key to cancel."), "9").String()
	}

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
