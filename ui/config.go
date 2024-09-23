package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/simonvetter/modbus"
)

// ModbusConfiguration represents a
// modbus client configuration
type ModbusConfiguration struct {
	Transport string `json:"transport,omitempty"`
	Addr      string `json:"address,omitempty"`
	BaudRate  uint   `json:"baudrate,omitempty"`
	DataBits  uint   `json:"databits,omitempty"`
	Parity    uint   `json:"parity,omitempty"`
	StopBits  uint   `json:"stopbits,omitempty"`
}

func (c *ModbusConfiguration) ClientConfiguration() *modbus.ClientConfiguration {
	cfg := &modbus.ClientConfiguration{}
	cfg.URL = fmt.Sprintf("%s://%s", c.Transport, c.Addr)
	cfg.DataBits = c.DataBits
	cfg.StopBits = c.StopBits
	cfg.Parity = c.Parity
	cfg.Speed = c.BaudRate

	return cfg
}

// PromptConfig prompts to the user to configurate
// the modbus connection.
func PromptConfig(cfg *ModbusConfiguration) error {
	addressInput := huh.NewInput()
	addressInput.Skip()
	stopBitsInput := newIntInput(1, 2)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose the transport").
				Options(
					huh.NewOption("TCP", "tcp"),
					huh.NewOption("RTU (Serial)", "rtu"),
					huh.NewOption("RTU over TCP", "rtuovertcp"),
					huh.NewOption("RTU over UDP", "rtuoverudp"),
				).
				Value(&cfg.Transport).
				Validate(func(s string) error {
					switch cfg.Transport {
					case "tcp", "rtuovertcp", "rtuoverudp":
						cfg.Addr = "localhost:502"
					default:
						cfg.Addr = "/dev/ttyUSB0"
					}

					// call updated to update the input value
					addressInput.Value(&cfg.Addr)
					return nil
				}),
			addressInput.
				Title("Enter the address").
				PlaceholderFunc(func() string {
					switch cfg.Transport {
					case "tcp", "rtuovertcp", "rtuoverudp":
						return "hostname-or-ip-address:502"
					}
					return "/dev/ttyUSB0"
				}, &cfg.Transport).
				SuggestionsFunc(func() []string {
					switch cfg.Transport {
					case "tcp", "rtuovertcp", "rtuoverudp":
						return []string{"localhost:502"}
					case "rtu":
						return []string{"/dev/ttyUSB0"}
					}
					return []string{}
				}, &cfg.Transport).
				Value(&cfg.Addr),
		),
		huh.NewGroup(
			huh.NewSelect[uint]().
				Title("Select a baud rate").
				Options(
					huh.NewOption("19.200 kBit/s", uint(19_200)),
					huh.NewOption("9.600 kBit/s", uint(9600)),
					huh.NewOption("4.800 kBit/s", uint(4800)),
					huh.NewOption("2.400 kBit/s", uint(2400)),
					huh.NewOption("1.200 kBit/s", uint(1200)),
					huh.NewOption("600 kBit/s", uint(600)),
					huh.NewOption("300 kBit/s", uint(300)),
				).
				Value(&cfg.BaudRate),

			huh.NewSelect[uint]().
				Title("Choose the parity").
				Options(
					huh.NewOption("None", modbus.PARITY_NONE),
					huh.NewOption("Even", modbus.PARITY_EVEN),
					huh.NewOption("Odd", modbus.PARITY_ODD),
				).
				Validate(func(val uint) error {
					switch val {
					case modbus.PARITY_NONE:
						cfg.StopBits = 2
					default:
						cfg.StopBits = 1
					}
					stopBitsInput.Accessor(NewNumberAccessor(&cfg.StopBits))
					return nil
				}).
				Value(&cfg.Parity),

			newIntInput(5, 8).
				Title("Enter the number of data bits").
				Accessor(NewNumberAccessor(&cfg.DataBits)),
			stopBitsInput.
				Title("Enter the number of stop bits").
				Accessor(NewNumberAccessor(&cfg.StopBits)),
		).
			Title("Modus over RTU Configuration").
			WithHideFunc(func() bool {
				return cfg.Transport != "rtu"
			}),
	).Run()
	return err
}
