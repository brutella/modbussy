package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/simonvetter/modbus"
	"github.com/xiam/to"
)

func PromptTable(client *modbus.ModbusClient, datapoints []*Datapoint) ([]*Datapoint, error) {
	t := NewTable(Theme, client)
	t.SetDatapoints(datapoints)

	// Read all datapoints on launch
	t.refreshAllDatapoints()

	// Focus the table
	t.Focus()

	// Run the program
	p := tea.NewProgram(t)
	_, err := p.Run()

	// Return the new list of datapoints
	return t.Datapoints, err
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type Model struct {
	*table.Model

	KeyMap         KeyMap
	Help           help.Model
	Status         *Status
	MaxColumnWidth int

	Datapoints []*Datapoint
	LastEdited *Datapoint

	needsLayout bool

	modbus *modbus.ModbusClient
}

func NewTable(theme *huh.Theme, client *modbus.ModbusClient) *Model {
	t := table.New()
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)
	t.SetColumns([]table.Column{
		{Title: "#"},
		{Title: "Server ID"},
		{Title: "Address"},
		{Title: "Flags"},
		{Title: "Name"},
		{Title: "Description"},
		{Title: "Value"},
	})

	m := Model{
		Model:          &t,
		KeyMap:         DefaultKeyMap(t.KeyMap),
		Help:           help.New(),
		Status:         NewStatus(theme),
		MaxColumnWidth: 50,
		modbus:         client,
	}
	m.Help.ShowAll = true

	return &m
}

// SetDatapoints sets the datapoints which are shown
func (m *Model) SetDatapoints(datapoints []*Datapoint) {
	m.Datapoints = datapoints
	m.updateRows()
	m.needsLayout = true
}

// newDatapoint returns a new datapoint with the slave id
// set based on existing datapoints.
func (m *Model) newDatapoint() Datapoint {
	new := Datapoint{}
	if m.LastEdited != nil {
		new.SlaveId = m.LastEdited.SlaveId
	} else {
		for _, dp := range m.Datapoints {
			new.SlaveId = dp.SlaveId
			break
		}
	}

	return new
}

func (m *Model) refreshAllDatapoints() {
	for _, dp := range m.Datapoints {
		m.readDatapoint(dp)
	}
}
func (m *Model) readDatapoint(dp *Datapoint) {
	m.modbus.SetUnitId(dp.SlaveId)

	var val any
	var err error
	switch dp.DataType {
	case DataTypeCoil:
		val, err = m.modbus.ReadCoil(dp.Addr)
	case DataTypeBool:
		val, err = m.modbus.ReadDiscreteInput(dp.Addr)
	case DataTypeUint16:
		val, err = m.modbus.ReadRegister(dp.Addr, dp.RegType())
	case DataTypeUint32:
		val, err = m.modbus.ReadUint32(dp.Addr, dp.RegType())
	case DataTypeUint64:
		val, err = m.modbus.ReadUint64(dp.Addr, dp.RegType())
	case DataTypeFloat32:
		val, err = m.modbus.ReadFloat32(dp.Addr, dp.RegType())
	case DataTypeFloat64:
		val, err = m.modbus.ReadFloat64(dp.Addr, dp.RegType())
	}

	if err != nil {
		dp.Err = err
		dp.Value = nil
	} else {
		dp.Err = nil
		dp.Value = val
	}
}

func (m *Model) writeDatapointValue(dp *Datapoint, val any) {
	m.modbus.SetUnitId(dp.SlaveId)

	var err error
	switch dp.DataType {
	case DataTypeBool:
		err = m.modbus.WriteCoil(dp.Addr, to.Bool(val))
	case DataTypeUint16:
		err = m.modbus.WriteRegister(dp.Addr, uint16(to.Uint64(val)))
	case DataTypeUint32:
		err = m.modbus.WriteUint32(dp.Addr, uint32(to.Uint64(val)))
	case DataTypeUint64:
		err = m.modbus.WriteUint64(dp.Addr, to.Uint64(val))
	case DataTypeFloat32:
		err = m.modbus.WriteFloat32(dp.Addr, float32(to.Float64(val)))
	case DataTypeFloat64:
		err = m.modbus.WriteFloat64(dp.Addr, to.Float64(val))
	}

	if err != nil {
		dp.Err = err
		dp.Value = nil
	} else {
		dp.Err = nil
		dp.Value = val
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) SelectedDatapoint() *Datapoint {
	selectedIndex := m.Cursor()

	if selectedIndex >= len(m.Datapoints) {
		return nil
	}

	return m.Datapoints[selectedIndex]
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, tea.ClearScreen
	case RefreshTickMsg:
		m.refreshAllDatapoints()
		m.updateRows()

		if !m.Status.AutoReload {
			return m, tea.ClearScreen
		}

		return m, tea.Sequence(tea.ClearScreen, refreshTickMsg(1*time.Second))

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Write):
			dp := m.SelectedDatapoint()
			if dp == nil {
				break
			}

			val, err := promptWrite(*dp)
			if err == nil {
				m.writeDatapointValue(dp, val)
				m.updateRows()
			}

			return m, tea.ClearScreen

		case key.Matches(msg, m.KeyMap.Refresh):
			m.refreshAllDatapoints()
			m.updateRows()
			return m, tea.ClearScreen

		case key.Matches(msg, m.KeyMap.StopRefresh):
			m.Status.AutoReload = false
			return m, tea.ClearScreen

		case key.Matches(msg, m.KeyMap.RefreshEverySec):
			m.Status.AutoReload = true
			return m, tea.Sequence(tea.ClearScreen, refreshTickMsg(1*time.Second))

		case key.Matches(msg, m.KeyMap.Edit):
			dp := m.SelectedDatapoint()
			if dp == nil {
				break
			}

			edited, err := promptDatapoint(*dp, fmt.Sprintf(`Edit "%s"`, dp.Name))
			if err == nil {

				m.Datapoints[m.Cursor()] = &edited
				m.LastEdited = &edited

				m.updateRows()
			}
			return m, tea.ClearScreen
		case key.Matches(msg, m.KeyMap.Add):
			new, err := promptDatapoint(m.newDatapoint(), "New Datapoint")
			if err == nil {

				m.Datapoints = append(m.Datapoints, &new)

				m.updateRows()
			}
			return m, tea.ClearScreen
		case key.Matches(msg, m.KeyMap.Duplicate):
			dp := m.SelectedDatapoint()
			if dp == nil {
				break
			}
			updated, err := promptDatapoint(*dp, "New Datapoint")
			if err == nil {
				updated.Value = nil

				m.Datapoints = append(m.Datapoints, &updated)

				m.updateRows()
			}
			return m, tea.ClearScreen
		case key.Matches(msg, m.KeyMap.Remove):
			dp := m.SelectedDatapoint()
			if dp == nil {
				break
			}

			i := m.Cursor()
			m.Datapoints = append(m.Datapoints[:i], m.Datapoints[i+1:]...)
			m.updateRows()

			if i < len(m.Rows()) {
				m.SetCursor(i)
			} else if len(m.Rows()) > 0 {
				m.SetCursor(len(m.Rows()))
			}
			return m, tea.ClearScreen
		case key.Matches(msg, m.KeyMap.MoveLineUp):

			cursor := m.Cursor()
			if cursor == 0 {
				break
			}
			up := cursor - 1

			m.Datapoints[up], m.Datapoints[cursor] = m.Datapoints[cursor], m.Datapoints[up]

			m.SetCursor(up)
			m.updateRows()
			return m, tea.ClearScreen
		case key.Matches(msg, m.KeyMap.MoveLineDown):

			n := len(m.Datapoints)

			cursor := m.Cursor()
			if cursor >= n-1 {
				break
			}
			down := cursor + 1

			m.Datapoints[down], m.Datapoints[cursor] = m.Datapoints[cursor], m.Datapoints[down]

			m.SetCursor(down)
			m.updateRows()
			return m, tea.ClearScreen
		default:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}
	table, cmd := m.Model.Update(msg)
	m.Model = &table

	return m, cmd
}

func (m Model) View() string {
	if m.needsLayout {
		m.layoutSubviews()
		m.Model.UpdateViewport()
	}

	m.KeyMap.AutoReload = m.Status.AutoReload
	return baseStyle.Render(m.Model.View()) + "\n" + m.Status.View(m.Model.Width()) + "\n" + m.HelpView() + "\n"
}

func (m Model) HelpView() string {
	return m.Help.View(m.KeyMap)
}

func (m Model) layoutSubviews() {
	cols := m.Columns()
	rows := m.Rows()
	for i, col := range cols {
		maxLen := len(col.Title)
		for _, row := range rows {
			str := row[i]
			if len(str) > maxLen {
				maxLen = len(str)
			}
		}

		cols[i].Width = min(maxLen, m.MaxColumnWidth)
	}
}

func (m *Model) updateRows() {
	rows := make([]table.Row, len(m.Datapoints))
	for i, datapoint := range m.Datapoints {
		rows[i] = append(table.Row{fmt.Sprintf("%d", i+1)}, datapoint.TableRow()...)
	}
	m.SetRows(rows)
}

type KeyMap struct {
	Table           table.KeyMap
	Add             key.Binding
	Remove          key.Binding
	Edit            key.Binding
	Refresh         key.Binding
	RefreshEverySec key.Binding
	StopRefresh     key.Binding
	Write           key.Binding

	Duplicate    key.Binding
	MoveLineUp   key.Binding
	MoveLineDown key.Binding

	AutoReload bool
}

func DefaultKeyMap(km table.KeyMap) KeyMap {
	// Edit the default table keymaps
	// km.LineDown.SetKeys("↓")
	// km.LineDown.SetHelp("↓", "down")
	// km.LineUp.SetKeys("↑")
	// km.LineUp.SetHelp("↑", "up")
	km.GotoBottom.SetEnabled(false)
	km.GotoTop.SetEnabled(false)
	km.HalfPageDown.SetEnabled(false)
	km.HalfPageUp.SetEnabled(false)
	km.PageDown.SetEnabled(false)
	km.PageUp.SetEnabled(false)

	return KeyMap{
		Table: km,
		Write: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "write"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reload"),
		),
		RefreshEverySec: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "auto reload (1s)"),
		),
		StopRefresh: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "stop auto reload"),
		),
		Add: key.NewBinding(
			key.WithKeys("+"),
			key.WithHelp("+", "add"),
		),
		Duplicate: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "duplicate"),
		),
		Remove: key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "remove"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		MoveLineUp: key.NewBinding(
			key.WithKeys("ctrl+k"),
			key.WithHelp("ctrl+k", "line up"),
		),
		MoveLineDown: key.NewBinding(
			key.WithKeys("ctrl+j"),
			key.WithHelp("ctrl+j", "line down"),
		),
	}
}

// ShortHelp implements the KeyMap interface.
func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Table.LineUp, km.Table.LineDown, km.Add, km.Remove, km.Edit, km.Refresh, km.Write}
}

// FullHelp implements the KeyMap interface.
func (km KeyMap) FullHelp() [][]key.Binding {
	upDown := []key.Binding{km.Table.LineUp, km.Table.LineDown, km.MoveLineUp, km.MoveLineDown}
	editing := []key.Binding{km.Add, km.Remove, km.Edit, km.Write, km.Duplicate}
	refresh := []key.Binding{km.Refresh, km.RefreshEverySec}
	if km.AutoReload {
		refresh[1] = km.StopRefresh
	}

	return [][]key.Binding{upDown, editing, refresh}
}
