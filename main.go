package main

import (
	"forge.rights.ninja/jeff/goedlink/n8command"
	"forge.rights.ninja/jeff/goedlink/n8serial"
)

var n8 n8serial.N8

func main() {
	n8.Address = "/dev/ttyACM0" // TODO: take this as a command argument

	n8serial.InitSerial(&n8)
	defer n8.Port.Close()

	n8command.ExitServiceMode(&n8)
	config := n8command.GetConfig(&n8)
	config.PrintFull()
}
