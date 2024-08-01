package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInitialModel(t *testing.T) {
	tests := []struct {
		name                          string
		width, height                 int
		expectedWidth, expectedHeight int
	}{
		{
			name:  "Default dimensions",
			width: 80, height: 20,
			expectedWidth:  30,
			expectedHeight: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel(tt.width, tt.height)
			assert.Equal(t, tt.expectedWidth, m.usernameInput.Width)
			assert.Equal(t, tt.expectedWidth, m.passwordInput.Width)
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name   string
		msg    tea.Msg
		setup  func(*model)
		verify func(*testing.T, model)
	}{
		{
			name: "InitialModel",
			setup: func(m *model) {
				*m = initialModel(80, 20)
			},
			verify: func(t *testing.T, m model) {
				assert.False(t, m.done)
				assert.Equal(t, 0, m.focused)
			},
		},
		{
			name: "Toggle Focus",
			msg:  tea.KeyMsg{Type: tea.KeyDown},
			setup: func(m *model) {
				*m = initialModel(80, 20)
			},
			verify: func(t *testing.T, m model) {
				assert.Equal(t, 1, m.focused)
			},
		},
		{
			name: "Enter Key",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
			setup: func(m *model) {
				*m = initialModel(80, 20)
				m.usernameInput.SetValue("testuser")
				m.passwordInput.SetValue("testpass")
			},
			verify: func(t *testing.T, m model) {
				assert.True(t, m.done)
				assert.Equal(t, "testuser", m.usernameInput.Value())
				assert.Equal(t, "testpass", m.passwordInput.Value())
			},
		},
		{
			name: "Exit with Ctrl+C",
			msg:  tea.KeyMsg{Type: tea.KeyCtrlC},
			setup: func(m *model) {
				*m = initialModel(80, 20)
			},
			verify: func(t *testing.T, m model) {
				newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
				model := newModel.(model)
				msg := cmd()
				assert.Equal(t, tea.Quit(), msg)
				assert.False(t, model.done)
			},
		},
		{
			name: "Mouse Click on Username Field",
			msg: tea.MouseMsg{
				Button: tea.MouseButtonLeft,
				Action: tea.MouseActionPress,
				X:      2, // X-coordinate that should fall within the username field
				Y:      2, // Y-coordinate that should fall within the username field
			},
			setup: func(m *model) {
				*m = initialModel(80, 20)
				m.focused = 1
			},
			verify: func(t *testing.T, m model) {
				newModel, _ := m.Update(tea.MouseMsg{
					Button: tea.MouseButtonLeft,
					Action: tea.MouseActionPress,
					X:      2,
					Y:      2,
				})
				model := newModel.(model)
				assert.Equal(t, 0, model.focused) // Username should be focused
			},
		},
		{
			name: "Mouse Click on Password Field",
			msg: tea.MouseMsg{
				Button: tea.MouseButtonLeft,
				Action: tea.MouseActionPress,
				X:      2, // X-coordinate that should fall within the password field
				Y:      8, // Y-coordinate that should fall within the password field
			},
			setup: func(m *model) {
				*m = initialModel(80, 20)
			},
			verify: func(t *testing.T, m model) {
				newModel, _ := m.Update(tea.MouseMsg{
					Button: tea.MouseButtonLeft,
					Action: tea.MouseActionPress,
					X:      2,
					Y:      8,
				})
				model := newModel.(model)
				assert.Equal(t, 1, model.focused) // Password should be focused
			},
		},
		{
			name: "Terminal Resize",
			msg: tea.WindowSizeMsg{
				Width:  100,
				Height: 30,
			},
			setup: func(m *model) {
				*m = initialModel(80, 20)
			},
			verify: func(t *testing.T, m model) {
				newModel, _ := m.Update(tea.WindowSizeMsg{
					Width:  100,
					Height: 30,
				})
				model := newModel.(model)
				assert.Equal(t, 100, model.width)
				assert.Equal(t, 30, model.height)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel(80, 20)
			if tt.setup != nil {
				tt.setup(&m)
			}
			if tt.msg != nil {
				newModel, _ := m.Update(tt.msg)
				m = newModel.(model)
			}
			if tt.verify != nil {
				tt.verify(t, m)
			}
		})
	}
}

func TestView(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*model)
		verify func(*testing.T, string)
	}{
		{
			name: "View with no input completed",
			setup: func(m *model) {
				*m = initialModel(80, 20)
			},
			verify: func(t *testing.T, view string) {
				assert.Contains(t, view, "Username")
				assert.Contains(t, view, "Password")
			},
		},

		{
			name: "View after input",
			setup: func(m *model) {
				*m = initialModel(80, 20)
				m.usernameInput.SetValue("testuser")
				m.passwordInput.SetValue("testpass")
				m.done = true
			},
			verify: func(t *testing.T, view string) {
				assert.Contains(t, view, "Username: testuser")
				assert.Contains(t, view, "Password length: 8")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel(80, 20)
			if tt.setup != nil {
				tt.setup(&m)
			}
			view := m.View()
			if tt.verify != nil {
				tt.verify(t, view)
			}
		})
	}
}
