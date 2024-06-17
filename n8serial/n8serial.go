package n8serial

import (
	"log"
	"time"

	"github.com/tarm/serial"
)

const ACK_BLOCK_SIZE int = 1024

const ADDR_PRG int = 0x00000000
const ADDR_CHR int = 0x00800000
const ADDR_SRM int = 0x01000000
const ADDR_CFG uint32 = 0x01800000
const ADDR_SSR int = 0x01802000
const ADDR_FIFO int = 0x01810000
const ADDR_FLA_MENU int = 0x00000 //boot fails mos6502 code
const ADDR_FLA_FPGA int = 0x40000 //boot fails fpga code
const ADDR_FLA_ICOR int = 0x80000 //mcu firmware update

const SIZE_PRG int = 0x800000
const SIZE_CHR int = 0x800000
const SIZE_SRM int = 0x40000

const ADDR_OS_PRG int = (ADDR_PRG + 0x7E0000)
const ADDR_OS_CHR int = (ADDR_CHR + 0x7E0000)

const FAT_READ byte = 0x01
const FAT_WRITE byte = 0x02
const FAT_OPEN_EXISTING byte = 0x00
const FAT_CREATE_NEW byte = 0x04
const FAT_CREATE_ALWAYS byte = 0x08
const FAT_OPEN_ALWAYS byte = 0x10
const FAT_OPEN_APPEND byte = 0x30

const CMD_STATUS byte = 0x10
const CMD_GET_MODE byte = 0x11
const CMD_HARD_RESET byte = 0x12
const CMD_GET_VDC byte = 0x13
const CMD_RTC_GET byte = 0x14
const CMD_RTC_SET byte = 0x15
const CMD_FLA_RD byte = 0x16
const CMD_FLA_WR byte = 0x17
const CMD_FLA_WR_SDC byte = 0x18
const CMD_MEM_RD byte = 0x19
const CMD_MEM_WR byte = 0x1A
const CMD_MEM_SET byte = 0x1B
const CMD_MEM_TST byte = 0x1C
const CMD_MEM_CRC byte = 0x1D
const CMD_FPG_USB byte = 0x1E
const CMD_FPG_SDC byte = 0x1F
const CMD_FPG_FLA byte = 0x20
const CMD_FPG_CFG byte = 0x21
const CMD_USB_WR byte = 0x22
const CMD_FIFO_WR byte = 0x23
const CMD_UART_WR byte = 0x24
const CMD_REINIT byte = 0x25
const CMD_SYS_INF byte = 0x26
const CMD_GAME_CTR byte = 0x27
const CMD_UPD_EXEC byte = 0x28

const CMD_DISK_INIT byte = 0xC0
const CMD_DISK_RD byte = 0xC1
const CMD_DISK_WR byte = 0xC2
const CMD_F_DIR_OPN byte = 0xC3
const CMD_F_DIR_RD byte = 0xC4
const CMD_F_DIR_LD byte = 0xC5
const CMD_F_DIR_SIZE byte = 0xC6
const CMD_F_DIR_PATH byte = 0xC7
const CMD_F_DIR_GET byte = 0xC8
const CMD_F_FOPN byte = 0xC9
const CMD_F_FRD byte = 0xCA
const CMD_F_FRD_MEM byte = 0xCB
const CMD_F_FWR byte = 0xCC
const CMD_F_FWR_MEM byte = 0xCD
const CMD_F_FCLOSE byte = 0xCE
const CMD_F_FPTR byte = 0xCF
const CMD_F_FINFO byte = 0xD0
const CMD_F_FCRC byte = 0xD1
const CMD_F_DIR_MK byte = 0xD2
const CMD_F_DEL byte = 0xD3

const CMD_USB_RECOV byte = 0xF0
const CMD_RUN_APP byte = 0xF1

type N8 struct {
	Address string
	Port    *serial.Port
}

func CloseSerial(n8 *N8) {
	n8.Port.Close()
	time.Sleep(time.Millisecond * 100)
}

func InitSerial(n8 *N8) {
	config := &serial.Config{
		Name:        n8.Address,
		Baud:        9600,
		Size:        8,                 // Data bits: 8
		Parity:      serial.ParityNone, // Parity: None
		StopBits:    serial.Stop1,      // Stop bits: 1
		ReadTimeout: time.Second * 1,   // Read timeout: 1 second
	}
	var err error
	n8.Port, err = serial.OpenPort(config)

	if err != nil {
		log.Fatalf("Failed to open serial port: %v", err)
	}
	time.Sleep(time.Millisecond * 100)
}

func TxCmd(n8 *N8, command byte) {
	cmd := make([]byte, 4)
	cmd[0] = byte('+')
	cmd[1] = byte('+' ^ 0xff)
	cmd[2] = command
	cmd[3] = byte(command ^ 0xff)

	_, err := n8.Port.Write(cmd)
	if err != nil {
		log.Fatalf("Failed to write to serial port: %v", err)
	}
}

func TxData(n8 *N8, buf []byte) {
	_, err := n8.Port.Write(buf)
	if err != nil {
		log.Fatalf("Failed to write to serial port: %v", err)
	}
}

func Tx8(n8 *N8, arg int) {
	var buf []byte = make([]byte, 1)
	buf[0] = (byte)(arg)
	TxData(n8, buf)
}

func Tx32(n8 *N8, arg uint32) {
	var buf []byte = make([]byte, 4)

	buf[3] = (byte)(arg >> 24)
	buf[2] = (byte)(arg >> 16)
	buf[1] = (byte)(arg >> 8)
	buf[0] = (byte)(arg)

	TxData(n8, buf)
}

func Rx8(n8 *N8) uint8 {
	buf := make([]byte, 1)

	_, err := n8.Port.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read from serial port: %v", err)
	}

	return (uint8)(buf[0])
}

func Rx16(n8 *N8) uint16 {
	buf := make([]byte, 2)

	_, err := n8.Port.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read from serial port: %v", err)
	}

	return ((uint16)(buf[1])<<8 | (uint16)(buf[0]))
}
