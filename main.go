package main

import (
    "fmt"
    "os"
	"time"

    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	game *Game
	inputHandler *TeaInputHandler
}

type frameMsg struct{}

func animate() tea.Cmd {
	return tea.Tick(time.Second/24, func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

func (m model) Init() tea.Cmd {
    return animate()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.game.Tick(1.0)

    switch msg := msg.(type) {
		case tea.KeyMsg:
			m.inputHandler.UpdatePressedKeys(msg.String())
			
			switch msg.String() {
				case "ctrl+c", "q":
					return m, tea.Quit
			}
		default:
			m.inputHandler.UpdatePressedKeys("")
			return m, animate()
	}

    // Return the updated model to the Bubble Tea runtime for processing.
    // Note that we're not returning a command.
    return m, nil
}

func (m model) View() string {
	// Return contents of game frame buffer
    return m.game.Draw()
}



/////////////////////////////////////////////////
// Tea Input Handler
type TeaInputHandler struct {
	pressedKeys string
}
func (t *TeaInputHandler) UpdatePressedKeys(pressedKeys string) {
	t.pressedKeys = pressedKeys
}
func (t *TeaInputHandler) IsKeyPressed(s string) bool {
	switch t.pressedKeys {
	case s:
		return true
	default:
		return false
	}
}

func main() {
	inputHandler := &TeaInputHandler{}
	InitGame(inputHandler)

	
	m := model{
		game: GetGame(),
		inputHandler: inputHandler,
	}

    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
}
