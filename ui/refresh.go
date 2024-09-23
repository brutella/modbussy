package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type RefreshTickMsg struct {
}

// refreshTickMsg sends a RefreshTickMsg message after a duration.
func refreshTickMsg(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(_ time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}
