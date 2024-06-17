package n8command

import (
	"log"
	"time"

	"forge.rights.ninja/jeff/goedlink/config"
	"forge.rights.ninja/jeff/goedlink/n8serial"
)

func GetStatus(n8 *n8serial.N8) uint16 {
	n8serial.TxCmd(n8, n8serial.CMD_STATUS) // TODO: check for error
	resp := n8serial.Rx16(n8)

	if (resp & 0xff00) != 0xA500 {
		log.Fatalf("Unexpected response code: %v", resp)
	}

	return resp & 0xff
}

func ExitServiceMode(n8 *n8serial.N8) {
	if !isServiceMode(n8) {
		return
	}

	n8serial.TxCmd(n8, n8serial.CMD_RUN_APP)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 100)
		n8serial.CloseSerial(n8)
		n8serial.InitSerial(n8)
	}
	// TODO: throw error/exit if N8 hasn't left service mode after some time
}

func isServiceMode(n8 *n8serial.N8) bool {
	n8serial.TxCmd(n8, n8serial.CMD_GET_MODE)
	resp := n8serial.Rx8(n8)

	return resp == 0xA1
}

func rxData(n8 *n8serial.N8, buf []byte) {
	n8.Port.Read(buf)
}

func memRD(n8 *n8serial.N8, addr uint32, buf []byte, len uint32) {
	if len == 0 {
		log.Fatalf("memRD: No data")
		return
	}

	n8serial.TxCmd(n8, n8serial.CMD_MEM_RD)
	n8serial.Tx32(n8, addr)
	n8serial.Tx32(n8, len)
	n8serial.Tx8(n8, 0)
	rxData(n8, buf)
}

func GetConfig(n8 *n8serial.N8) *config.MapConfig {
	var cfg *config.MapConfig = config.NewMapConfig()

	data := make([]byte, len(cfg.GetBinary()))

	memRD(n8, n8serial.ADDR_CFG, data, (uint32)(len(cfg.GetBinary())))

	return config.NewMapConfigFromBinary(data)
}
