package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	usernameInput textinput.Model
	passwordInput textinput.Model
	focused       int
	width, height int
	done          bool
}

var (
	focusedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).Padding(1).BorderForeground(lipgloss.Color("#FFA500"))
	blurredStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1).BorderForeground(lipgloss.Color("#FFFFFF"))
	titleStyle   = lipgloss.NewStyle().Bold(true).PaddingLeft(1)
	minWidth     = 20
	maxWidth     = 50
	minHeight    = 2
	maxHeight    = 5
)

// Initial model setup
func initialModel(width, height int) model {
	usernameInput := textinput.New()
	usernameInput.Placeholder = "Username"
	usernameInput.Focus()
	usernameInput.Width = 30 // This will be adjusted in the View method

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.Width = 30 // This will be adjusted in the View method

	return model{
		usernameInput: usernameInput,
		passwordInput: passwordInput,
		focused:       0, // Start with username input focused
		width:         width,
		height:        height,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down":
			if m.focused == 0 {
				m.focused = 1
				m.passwordInput.Focus()
				m.usernameInput.Blur()
			} else {
				m.focused = 0
				m.usernameInput.Focus()
				m.passwordInput.Blur()
			}
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			m.done = true
		}
	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {

			usernameBox := blurredStyle.Render(
				titleStyle.Render("Username") + "\n" + m.usernameInput.View(),
			)
			passwordBox := blurredStyle.Render(
				titleStyle.Render("Password") + "\n" + m.passwordInput.View(),
			)

			// Measure the height of the boxes by counting lines
			usernameBoxLines := len(strings.Split(usernameBox, "\n"))
			passwordBoxLines := len(strings.Split(passwordBox, "\n"))

			// Calculate top and bottom bounds for each box
			usernameBoxTop, usernameBoxBottom := 1, usernameBoxLines
			passwordBoxTop, passwordBoxBottom := usernameBoxLines+2, usernameBoxLines+2+passwordBoxLines

			if msg.Y >= usernameBoxTop && msg.Y <= usernameBoxBottom {
				m.focused = 0
				m.usernameInput.Focus()
				m.passwordInput.Blur()
			} else if msg.Y >= passwordBoxTop && msg.Y <= passwordBoxBottom {
				m.focused = 1
				m.passwordInput.Focus()
				m.usernameInput.Blur()
			}
		}
	case tea.WindowSizeMsg:
		// Update terminal size
		m.width = msg.Width
		m.height = msg.Height
	}
	// Update only the focused input
	var cmd tea.Cmd
	if m.focused == 0 {
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	} else {
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.done {
		return fmt.Sprintf(
			"Username: %s\nPassword length: %d",
			m.usernameInput.Value(),
			len(m.passwordInput.Value()),
		)
	}

	// Adjust the box sizes based on terminal dimensions
	padding := 2
	boxWidth := m.width - 2*padding
	if boxWidth < minWidth {
		boxWidth = minWidth
	}
	if boxWidth > maxWidth {
		boxWidth = maxWidth
	}

	// Calculate box heights based on terminal height
	boxHeight := (m.height - 10) / 2
	if boxHeight < minHeight {
		boxHeight = minHeight
	}
	if boxHeight > maxHeight {
		boxHeight = maxHeight
	}

	// Render the boxes
	usernameBox := blurredStyle.Render(
		lipgloss.NewStyle().Width(boxWidth).Height(boxHeight).Render(
			titleStyle.Render("Username") + "\n" + m.usernameInput.View(),
		),
	)
	passwordBox := blurredStyle.Render(
		lipgloss.NewStyle().Width(boxWidth).Height(boxHeight).Render(
			titleStyle.Render("Password") + "\n" + m.passwordInput.View(),
		),
	)

	if m.focused == 0 {
		usernameBox = focusedStyle.Render(
			lipgloss.NewStyle().Width(boxWidth).Height(boxHeight).Render(
				titleStyle.Render("Username") + "\n" + m.usernameInput.View(),
			),
		)
	} else {
		passwordBox = focusedStyle.Render(
			lipgloss.NewStyle().Width(boxWidth).Height(boxHeight).Render(
				titleStyle.Render("Password") + "\n" + m.passwordInput.View(),
			),
		)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, usernameBox, passwordBox)
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
}

func clearTerminal() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	clearTerminal()

	// Get initial terminal size
	width, height, err := getTerminalSize()
	if err != nil {
		fmt.Printf("Error getting terminal size: %v", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(width, height), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	clearTerminal()
}

func getTerminalSize() (width, height int, err error) {
	cmd := exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	width, err = strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, 0, err
	}

	cmd = exec.Command("tput", "lines")
	cmd.Stdin = os.Stdin
	out, err = cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	height, err = strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}
