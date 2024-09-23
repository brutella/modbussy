package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type Status struct {
	Text       string
	Err        error
	AutoReload bool

	textStyle       lipgloss.Style
	errStyle        lipgloss.Style
	autoReloadStyle lipgloss.Style
}

func NewStatus(theme *huh.Theme) *Status {
	return &Status{
		textStyle:       theme.Focused.Base,
		errStyle:        theme.Focused.ErrorMessage,
		autoReloadStyle: lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#58F236")).SetString("Auto Reloading"),
	}
}

func (s Status) View(width int) string {
	w := func(str ...string) int {
		w := 0
		for _, s := range str {
			w += lipgloss.Width(s)
		}
		return w
	}
	var text string
	var err string
	var empty string
	var autoReload string

	if s.Text != "" {
		text = s.textStyle.Render(s.Text)
	}

	if s.Err != nil {
		err = s.errStyle.Render(s.Err.Error())
	}

	if s.AutoReload {
		autoReload = s.autoReloadStyle.Render()
	}

	emptySpace := width - w(text, err, autoReload)
	empty = lipgloss.NewStyle().Width(emptySpace).Render()

	return lipgloss.JoinHorizontal(lipgloss.Top, text, err, empty, autoReload)
}
