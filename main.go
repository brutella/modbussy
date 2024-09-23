package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/brutella/modbussy/ui"
	"github.com/simonvetter/modbus"
)

type storage struct {
	Datapoints []*ui.Datapoint         `json:"datapoints"`
	Modbus     *ui.ModbusConfiguration `json:"modbus"`
}

func main() {
	dbFlag := flag.String("db", ".modbussy", "Path to database file")
	transportFlag := flag.String("transport", "", "Transport type (either rtu, tcp, rtuovertcp, or rtuoverudp)")
	addressFlag := flag.String("address", "", "Address of modbus server")
	baudRate := flag.Uint("baudrate", 19200, "RTU Baudrate")
	dataBits := flag.Uint("databits", 8, "RTU Data Bits")
	parity := flag.String("parity", "E", "RTU Parity; either E(ven), N(one), O(dd)")
	stopBits := flag.Uint("stopbits", 1, "RTU Stop Bits")

	flag.Parse()

	// Read the stored data
	stg := storage{
		Datapoints: []*ui.Datapoint{},
		Modbus:     &ui.ModbusConfiguration{},
	}
	buf, err := os.ReadFile(*dbFlag)
	if err == nil {
		json.Unmarshal(buf, &stg)
	}

	if transportFlag != nil && len(*transportFlag) > 0 {
		stg.Modbus.Transport = *transportFlag
	}

	if addressFlag != nil && len(*addressFlag) > 0 {
		stg.Modbus.Addr = *addressFlag
	}
	if baudRate != nil {
		stg.Modbus.BaudRate = *baudRate
	}

	if dataBits != nil {
		stg.Modbus.DataBits = *dataBits
	}

	if parity != nil {
		switch *parity {
		case "E":
			stg.Modbus.Parity = modbus.PARITY_EVEN
		case "O":
			stg.Modbus.Parity = modbus.PARITY_ODD
		case "N":
			stg.Modbus.Parity = modbus.PARITY_NONE
		default:
			log.Fatalf(`invalid parity "%s"`, *parity)
		}
	}

	if stopBits != nil {
		stg.Modbus.StopBits = *stopBits
	}

	for {
		// Prompt modbus configuration
		err := ui.PromptConfig(stg.Modbus)
		if err != nil {
			logError(err)
			os.Exit(1)
		}

		// Connect to modbus
		client, err := modbus.NewClient(stg.Modbus.ClientConfiguration())
		if err != nil {
			logError(err)
			continue
		}
		err = client.Open()
		if err != nil {
			logError(err)
			continue
		}

		// Prompt the data table
		stg.Datapoints, _ = ui.PromptTable(client, stg.Datapoints)
		client.Close()

		// Store the returned data
		buf, err := json.Marshal(stg)
		if err != nil {
			logError(err)
		} else {
			err = os.WriteFile(*dbFlag, buf, 0644)
			if err != nil {
				logError(err)
			}
		}
		os.Exit(1)
	}
}

// logError renders an error based on the current theme.
func logError(err error) {
	fmt.Println(ui.Theme.Focused.ErrorMessage.Render(err.Error()))
}
