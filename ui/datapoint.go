package ui

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/simonvetter/modbus"
	"github.com/xiam/to"
)

// DataType represents a data type.
type DataType byte

const (
	DataTypeUint16 DataType = iota
	DataTypeUint32
	DataTypeUint64
	DataTypeFloat32
	DataTypeFloat64
	DataTypeBool
	DataTypeCoil
)

// Flag represents a flag.
type Flag byte

const (
	FlagRead Flag = iota
	FlagReadWrite
)

// Datapoint represents a value from a modbus server.
// It specifies the datatype, flags and unit.
type Datapoint struct {
	SlaveId     uint8    `json:"slaveId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Addr        uint16   `json:"addr"`
	DataType    DataType `json:"data-type"`
	Flag        Flag     `json:"flag"`
	Unit        string   `json:"unit"`

	// Scaling specifies a scaling of the datapoint value.
	// For example if value can be negative, we have to scale
	// the unsigned value to be negative.
	Scaling *Scaling `json:"scaling"`

	// Value is the last read value.
	Value any `json:"-"`

	// Err is not-nil if the last read failed.
	Err error `json:"-"`
}

func (dp Datapoint) RegType() modbus.RegType {
	switch dp.Flag {
	case FlagRead:
		return modbus.INPUT_REGISTER
	case FlagReadWrite:
		return modbus.HOLDING_REGISTER
	}

	return math.MaxUint
}

func (dp Datapoint) fmtValue(val any) string {
	if val == nil {
		return "-"
	}
	if dp.Scaling == nil || !dp.Scaling.Valid() {
		return fmt.Sprintf("%v%s", val, dp.Unit)
	}

	minIn, maxIn, minOut, maxOut := dp.Scaling.Ranges()
	inSize := math.Abs(minIn - maxIn)
	outSize := math.Abs(minOut - maxOut)
	floatVal := to.Float64(val)
	res := (floatVal-minIn)/inSize*outSize + minOut
	return fmt.Sprintf("%0.2f%s", res, dp.Unit)
}

func (dp Datapoint) TableRow() table.Row {
	value := dp.fmtValue(dp.Value)
	if dp.Err != nil {
		value = Theme.Focused.ErrorMessage.Render(dp.Err.Error())
	}

	flags := "R"
	switch dp.Flag {
	case FlagReadWrite:
		flags = "RW"
	}

	return table.Row{
		fmt.Sprintf("%d", dp.SlaveId),
		fmt.Sprintf("%d", dp.Addr),
		flags,
		dp.Name,
		dp.Description,
		value,
	}
}

type Scaling struct {
	MinIn  string `json:"minIn"`
	MaxIn  string `json:"maxIn"`
	MinOut string `json:"minOut"`
	MaxOut string `json:"maxOut"`
}

func (s Scaling) Valid() bool {
	return len(s.MinIn) > 0 && len(s.MaxIn) > 0 && len(s.MinOut) > 0 && len(s.MaxOut) > 0
}
func (dp Scaling) Ranges() (minIn float64, maxIn float64, minOut float64, maxOut float64) {
	minIn = to.Float64(dp.MinIn)
	maxIn = to.Float64(dp.MaxIn)
	minOut = to.Float64(dp.MinOut)
	maxOut = to.Float64(dp.MaxOut)
	return
}

// promptWrite prompts to write a value to a datapoint.
func promptWrite(dp Datapoint) (any, error) {
	write := true
	var val string = "1"
	if dp.Value != nil {
		val = fmt.Sprintf("%v", dp.Value)
	}

	km := huh.NewDefaultKeyMap()
	km.Quit.SetKeys("esc")
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf(`Write value to "%s"`, dp.Name)).
				Value(&val),

			huh.NewConfirm().
				Affirmative("Write").
				Negative("Cancel").
				Value(&write),
		),
	).
		WithKeyMap(km).
		Run()

	if !write {
		return nil, errors.New("canceled")
	}

	return val, err
}

// promptDatapoint shows the editable fields of a datapoint.
func promptDatapoint(dp Datapoint, title string) (Datapoint, error) {
	newScaling := func(minIn, maxIn, minOut, maxOut any) Scaling {
		return Scaling{
			MinIn:  fmt.Sprintf("%v", minIn),
			MaxIn:  fmt.Sprintf("%v", maxIn),
			MinOut: fmt.Sprintf("%v", minOut),
			MaxOut: fmt.Sprintf("%v", maxOut),
		}
	}
	max16 := uint32(math.MaxUint16)
	max32 := uint32(math.MaxUint32)
	var scalings = map[DataType]Scaling{
		DataTypeBool:    newScaling(0, 1, 0, 1),
		DataTypeCoil:    newScaling(0, 1, 0, 1),
		DataTypeUint16:  newScaling(0, max16, 0, max16),
		DataTypeUint32:  newScaling(0, max32, 0, max32),
		DataTypeFloat32: newScaling(0, max32, 0, max32),
		DataTypeFloat64: newScaling(0, max32, 0, max32),
		DataTypeUint64:  newScaling(0, max32, 0, max32),
	}

	if scaling, ok := scalings[dp.DataType]; ok && dp.Scaling != nil {
		scaling.MinIn = dp.Scaling.MinIn
		scaling.MinOut = dp.Scaling.MinOut
		scaling.MaxIn = dp.Scaling.MaxIn
		scaling.MaxOut = dp.Scaling.MaxOut
		scalings[dp.DataType] = scaling
	}

	scaling := scalings[dp.DataType]

	validateInt := (func(s string) error {
		if len(s) == 0 {
			return nil
		}
		_, err := strconv.Atoi(s)
		if err != nil {
			return errors.New("input not a number")
		}
		return nil
	})

	theme := Theme
	theme.Focused.Title = theme.Focused.Title.Width(15).AlignHorizontal(lipgloss.Right)
	theme.Blurred.Title = theme.Blurred.Title.Foreground(theme.Blurred.TextInput.Text.GetForeground()).Width(15).AlignHorizontal(lipgloss.Right)

	km := huh.NewDefaultKeyMap()
	km.Quit.SetKeys("esc")

	old := dp
	save := true
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Server ID").
				Prompt(":").
				Inline(true).
				Accessor(NewNumberAccessor(&dp.SlaveId)).
				WithTheme(theme),

			huh.NewInput().
				Title("Name").
				Prompt(":").
				Placeholder("Enter a name").
				Inline(true).
				Value(&dp.Name).
				WithTheme(theme),

			huh.NewInput().
				Title("Description").
				Prompt(":").
				Placeholder("Enter a description").
				Inline(true).
				Value(&dp.Description).
				WithTheme(theme),

			newIntInput(0, 65_535).
				Title("Address").
				Prompt(":").
				Inline(true).
				Accessor(NewNumberAccessor(&dp.Addr)).
				WithTheme(theme),

			huh.NewSelect[Flag]().
				Title("Flags").
				Inline(true).
				Options(
					huh.NewOption("Read", FlagRead),
					huh.NewOption("Read & Write", FlagReadWrite),
				).
				Validate(func(f Flag) error {
					switch f {
					case FlagRead, FlagReadWrite:
						return nil
					}
					return errors.New("flag required")
				}).
				Value(&dp.Flag).
				WithTheme(theme),

			huh.NewSelect[DataType]().
				Title("Datatype").
				Inline(true).
				Options(
					huh.NewOption("Coil", DataTypeCoil),
					huh.NewOption("Bool", DataTypeBool),
					huh.NewOption("Uint16", DataTypeUint16),
					huh.NewOption("Uint32", DataTypeUint32),
					huh.NewOption("Uint64", DataTypeUint64),
					huh.NewOption("Float32", DataTypeFloat32),
					huh.NewOption("Float64", DataTypeFloat64),
				).
				Value(&dp.DataType).
				WithTheme(theme),

			huh.NewInput().
				Title("Unit").
				Prompt(":").
				Inline(true).
				Value(&dp.Unit).
				WithTheme(theme),

			huh.NewInput().
				Title("Input Min").
				Prompt(":").
				Validate(validateInt).
				Inline(true).
				Value(&scaling.MinIn).
				WithTheme(theme),

			huh.NewInput().
				Title("Max").
				Prompt(":").
				Validate(validateInt).
				Inline(true).
				Value(&scaling.MaxIn).
				WithTheme(theme),

			huh.NewInput().
				Title("Output Min").
				Prompt(":").
				Validate(validateInt).
				Inline(true).
				Value(&scaling.MinOut).
				WithTheme(theme),

			huh.NewInput().
				Title("Max").
				Prompt(":").
				Validate(validateInt).
				Inline(true).
				Value(&scaling.MaxOut).
				WithTheme(theme),

			huh.NewConfirm().
				Key("s").
				Affirmative("Save").
				Negative("Cancel").
				Value(&save).
				WithTheme(theme),
		).
			Title(title), // TODO: Doesn't seem to do anything; see https://github.com/charmbracelet/huh/issues/298
	).
		WithKeyMap(km).
		Run()

	if !save || err != nil {
		return old, errors.New("canceled")
	}
	dp.Scaling = &scaling

	return dp, nil
}
