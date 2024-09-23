# modbussy

`modbussy` is a command line to work with modbus networks. You can define datapoints which are then read from modbus servers. It's a little utility for modbus.

## Installing

You can install `modbussy` with Homebrew, Go or from Github

```shell
# Homebrew
brew install modbussy

# Install with Go
go install github.com/brutella/modbussy@latest

# Download from Github
...
```

## Usage

Run `modbussy` by executing the command `modbussy`. Easy!

### Configuration

The first screen lets you configure the connection to modbus. You can connect via 
- RTU
- TCP
- RTU via TCP
- or RTU via UDP.

Then specify the address and optionally the data rate, parity, the number of start and stop bits.

### Main UI

The main UI shows a list of datapoints. 
- Press `+` to add a new datapoints and specify the server id, address, name, datatype and flag (readonly, or read-writable).
- You can reload tha list of datapoints by pressing `r`.
- Once you have a list of datapoints, you can monitor with the auto-reload feature by pressing `l`.

### Writing values
- You can write to a datapoints by selecting it in the table and then pressing `w`.
- Enter a new value and choose `Write`.
