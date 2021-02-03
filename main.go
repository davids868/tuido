package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	input "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	te "github.com/muesli/termenv"
	"io/ioutil"
	"os"
)

type Action int

const (
	Add Action = iota
	Edit
	Delete
	None
)

type model struct {
	Todos       []string
	Cursor      int
	Selected    map[int]struct{}
	BeingEdited int
	Action      Action
	TextInput   input.Model
	FilePath    string
}

func initTextInput() input.Model {
	inputModel := input.NewModel()
	inputModel.Placeholder = "What would you like to do?"
	inputModel.Focus()
	inputModel.CharLimit = 156
	inputModel.Width = 40
	inputModel.Prompt = colorFg(te.String("> "), "2").Bold().String()

	return inputModel
}

func initialModel() *model {
	return &model{
		Todos:    []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
		Selected: make(map[int]struct{}),
		Cursor:   0,
		Action:   None,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func clearAndHideInput(m model) model {
	m.TextInput.SetValue("")
	m.Action = None

	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Action == Delete {
			switch msg.String() {
			case "y", "Y":
				if len(m.Todos) > 1 {
					m.Todos = append(m.Todos[:m.Cursor], m.Todos[m.Cursor+1:]...)
				} else {
					m.Todos = nil
				}
			}

			m.Action = None
		} else if m.Action == Add || m.Action == Edit {
			switch msg.String() {
			case "esc", "ctrl-c":
				m = clearAndHideInput(m)
			case "enter":
				value := m.TextInput.Value()

				if len(value) > 0 {
					if m.Action == Add {
						m.Todos = append(m.Todos, value)
					} else if m.Action == Edit {
						m.Todos[m.Cursor] = value
					}
				}

				m = clearAndHideInput(m)
			}

			m.TextInput, cmd = m.TextInput.Update(msg)
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				save(m)
				return m, tea.Quit

			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
				}

			case "down", "j":
				if m.Cursor < len(m.Todos)-1 {
					m.Cursor++
				}

			case "o", "a":
				m.Action = Add

			case "i", "e":
				m.Action = Edit
				m.TextInput.SetValue(m.Todos[m.Cursor])
				m.TextInput.CursorEnd()

			case "d":
				m.Action = Delete
				if len(m.Todos) > 1 {
					m.Action = Delete
				}

			case "enter", " ":
				_, ok := m.Selected[m.Cursor]
				if ok {
					delete(m.Selected, m.Cursor)
				} else {
					m.Selected[m.Cursor] = struct{}{}
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

	if len(m.Todos) <= 0 {
		s += "The list is empty! press 'a' or 'o' to add a todo"
	} else {
		for i, todo := range m.Todos {
			if m.Action == Edit && m.Cursor == i {
				s += m.TextInput.View() + "\n"
			} else {
				cursor := " "
				if m.Cursor == i {
					cursor = ">"
				}

				checked := " "
				if _, ok := m.Selected[i]; ok {
					checked = "x"
					todo = te.String(todo).Faint().String()
				}

				s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, todo)
			}
		}
	}

	switch m.Action {
	case Add:
		s += "\n" + m.TextInput.View()
	case Delete:
		s += colorFg(te.String("\nDelete todo? press 'y' to confirm or any other key to cancel."), "9").String()
	}

	return s
}

func load(path string) *model {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	model := model{}
	err = json.Unmarshal(data, &model)
	if err != nil {
		fmt.Print(err)
		return nil
	}

	return &model
}

func save(m model) {
	mJson, _ := json.Marshal(m)
	err := ioutil.WriteFile(m.FilePath, mJson, 0644)

	if err != nil {
		fmt.Print(err)
	}
}

func get_path() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	home, _ := os.UserHomeDir()
	os.Mkdir(home+"/.tuido", os.ModePerm)

	return home + "/.tuido/todos"
}

func main() {
	path := get_path()

	model := load(path)
	if model == nil {
		model = initialModel()
	}
	model.TextInput = initTextInput()
	model.FilePath = path

	p := tea.NewProgram(*model)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
