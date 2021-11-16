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
	None Action = iota
	Add
	Edit
	Delete
)

type Todo struct {
	Text    string
	Checked bool
}

type Data struct {
	Todos  []Todo
	Cursor int
}

type Model struct {
	Data        Data
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

func initialModel() *Model {
	return &Model{
		Data: Data{
			Todos: []Todo{
				{Text: "Buy carrots", Checked: false},
				{Text: "Buy celery", Checked: false},
				{Text: "Buy kohlrabi", Checked: false},
			}, Cursor: 0,
		},
		Action: None,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func clearAndHideInput(m Model) Model {
	m.TextInput.SetValue("")
	m.Action = None

	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Action == Delete {
			switch msg.String() {
			case "y", "Y":
				if len(m.Data.Todos) > 1 {
					m.Data.Todos = append(m.Data.Todos[:m.Data.Cursor], m.Data.Todos[m.Data.Cursor+1:]...)
				} else {
					m.Data.Todos = nil
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
						m.Data.Todos = append(m.Data.Todos, Todo{Text: value, Checked: false})
					} else if m.Action == Edit {
						m.Data.Todos[m.Data.Cursor].Text = value
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
				if m.Data.Cursor > 0 {
					m.Data.Cursor--
				}

			case "down", "j":
				if m.Data.Cursor < len(m.Data.Todos)-1 {
					m.Data.Cursor++
				}

			case "o", "a":
				m.Action = Add

			case "i", "e":
				m.Action = Edit
				m.TextInput.SetValue(m.Data.Todos[m.Data.Cursor].Text)
				m.TextInput.CursorEnd()

			case "d":
				m.Action = Delete
				if len(m.Data.Todos) > 1 {
					m.Action = Delete
				}

			case "enter", " ":
				m.Data.Todos[m.Data.Cursor].Checked = !m.Data.Todos[m.Data.Cursor].Checked
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

func (m Model) View() string {
	title := te.String(" Todo List: \n")
	title = colorFg(title, "15")
	title = colorBg(title, "8")
	title = title.Bold()
	s := title.String()

	if len(m.Data.Todos) <= 0 {
		s += "The list is empty! press 'a' or 'o' to add a todo"
	} else {
		for i, todo := range m.Data.Todos {
			if m.Action == Edit && m.Data.Cursor == i {
				s += m.TextInput.View() + "\n"
			} else {
				text := todo.Text
				cursor := " "
				if m.Data.Cursor == i {
					cursor = ">"
				}

				checked := " "
				if m.Data.Todos[i].Checked {
					checked = "x"
					text = te.String(text).Faint().String()
				}

				s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, text)
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

func load(path string) *Model {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	model := Model{}
	err = json.Unmarshal(data, &model.Data)
	if err != nil {
		fmt.Print(err)
		return nil
	}

	return &model
}

func save(m Model) {
	mJson, _ := json.Marshal(m.Data)
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

	return home + "/.tuido/todos.json"
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
