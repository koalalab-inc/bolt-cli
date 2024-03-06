package pkg

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Choice struct {
	Label  string
	Value  interface{}
	Parent *Choice
}

type Model struct {
	Choices  []*Choice        // items on the to-do list
	cursor   int              // which to-do list item our cursor is pointing at
	Selected map[int]struct{} // which to-do items are selected
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			m.Selected = make(map[int]struct{})
			return m, tea.Quit

		case "ctrl+a":
			// Toggle all items
			selectedIndices := len(m.Selected)
			if selectedIndices > 0 {
				m.Selected = make(map[int]struct{})
			} else {
				for i := range m.Choices {
					m.Selected[i] = struct{}{}
				}
			}

		// The "up" keys move the cursor up
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.Choices)-1 {
				m.cursor++
			}

		// The spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case " ":
			childrenIndices := []int{}
			choice := m.Choices[m.cursor]
			for j, c := range m.Choices {
				if c.Parent == choice {
					childrenIndices = append(childrenIndices, j)
				}
			}
			_, ok := m.Selected[m.cursor]
			if ok {
				delete(m.Selected, m.cursor)
				for _, child := range childrenIndices {
					delete(m.Selected, child)
				}
			} else {
				m.Selected[m.cursor] = struct{}{}
				for _, child := range childrenIndices {
					m.Selected[child] = struct{}{}
				}
			}
			siblingIndices := []int{}
			parentIndex := -1
			if choice.Parent != nil {
				for j, c := range m.Choices {
					if c.Parent == choice.Parent {
						siblingIndices = append(siblingIndices, j)
					}
					if c == choice.Parent {
						parentIndex = j
					}
				}
			}
			allSiblingSelected := true
			for _, siblingIndex := range siblingIndices {
				if _, ok := m.Selected[siblingIndex]; !ok {
					allSiblingSelected = false
					break
				}
			}
			if parentIndex != -1 && len(siblingIndices) > 1 {
				if allSiblingSelected {
					m.Selected[parentIndex] = struct{}{}
				} else {
					delete(m.Selected, parentIndex)
				}
			}

		case "enter":
			// If the user hits the "enter" key, we'll return a command
			// that tells the Bubble Tea runtime to quit.
			return m, tea.Quit
		}

	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	// The header
	s := "\033[96m⚡⚡Select the workflows/jobs that you want to fasten using Bolt?⚡⚡\033[0m\n"
	s += "--------------------------------------------------------------------------------\n"
	s += "\033[33mUse the arrow keys: up(↑) down(↓) to move and <space> to select.\033[0m\n"
	s += "\033[33mUse <ctrl+a> to toggle all.\033[0m\n\n"
	s += "NOTE: Only the jobs that are not fastened with bolt will be shown.\n\n"

	// Iterate over our choices
	for i, choice := range m.Choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.Selected[i]; ok {
			checked = "\033[92m✔\033[0m" // selected!
		}

		// Add indentation for child items
		indentation := ""
		if choice.Parent != nil {
			indentation = "  "
		}

		// Render the row
		if m.cursor == i {
			s += fmt.Sprintf("%s%s [%s] \033[107m\033[30m%s\033[0m\n", indentation, cursor, checked, choice.Label)
		} else {
			s += fmt.Sprintf("%s%s [%s] %s\n", indentation, cursor, checked, choice.Label)
		}
	}

	// The footer
	s += "\n\033[33mPress <enter> to proceed\033[0m\n"
	s += "\033[33mPress <q> or <ctrl+c> to quit.\033[0m\n"
	s += "--------------------------------------------------------------------------------\n"

	// Send the UI for rendering
	return s
}
