package n8

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

const ACK_BLOCK_SIZE uint32 = 0x0400

const (
	ADDR_CFG      uint32 = 0x01800000
	ADDR_SSR      uint32 = 0x01802000
	ADDR_FIFO     uint32 = 0x01810000
	ADDR_FLA_MENU uint32 = 0x00000000 //boot fails mos6502 code
	ADDR_FLA_FPGA uint32 = 0x00040000 //boot fails fpga code
	ADDR_FLA_ICOR uint32 = 0x00080000 //mcu firmware update
)

const (
	SIZE_PRG uint32 = 0x800000
	SIZE_CHR uint32 = 0x800000
	SIZE_SRM uint32 = 0x040000
)

const (
	CMD_EXEC           uint8 = 0x00
	CMD_STATUS         uint8 = 0x10
	CMD_GET_MODE       uint8 = 0x11
	CMD_HARD_RESET     uint8 = 0x12
	CMD_GET_VDC        uint8 = 0x13
	CMD_RTC_GET        uint8 = 0x14
	CMD_RTC_SET        uint8 = 0x15
	CMD_FLA_RD         uint8 = 0x16
	CMD_FLA_WR         uint8 = 0x17
	CMD_FLA_WR_SDC     uint8 = 0x18
	CMD_MEM_RD         uint8 = 0x19
	CMD_MEM_WR         uint8 = 0x1A
	CMD_MEM_SET        uint8 = 0x1B
	CMD_MEM_TST        uint8 = 0x1C
	CMD_MEM_CRC        uint8 = 0x1D
	CMD_FPGA_USB       uint8 = 0x1E
	CMD_FPGA_SDC       uint8 = 0x1F
	CMD_FPGA_FLA       uint8 = 0x20
	CMD_FPGA_CFG       uint8 = 0x21
	CMD_USB_WR         uint8 = 0x22
	CMD_FIFO_WR        uint8 = 0x23
	CMD_UART_WR        uint8 = 0x24
	CMD_REINIT         uint8 = 0x25
	CMD_SYS_INF        uint8 = 0x26
	CMD_GAME_CTR       uint8 = 0x27
	CMD_UPD_EXEC       uint8 = 0x28
	CMD_DISK_INIT      uint8 = 0xC0
	CMD_DISK_READ      uint8 = 0xC1
	CMD_DISK_WRITE     uint8 = 0xC2
	CMD_FILE_DIR_OPEN  uint8 = 0xC3
	CMD_FILE_DIR_READ  uint8 = 0xC4
	CMD_FILE_DIR_LD    uint8 = 0xC5
	CMD_FILE_DIR_SIZE  uint8 = 0xC6
	CMD_FILE_DIR_PATH  uint8 = 0xC7
	CMD_FILE_DIR_GET   uint8 = 0xC8
	CMD_FILE_OPEN      uint8 = 0xC9
	CMD_FILE_READ      uint8 = 0xCA
	CMD_FILE_READ_MEM  uint8 = 0xCB
	CMD_FILE_WRITE     uint8 = 0xCC
	CMD_FILE_WRITE_MEM uint8 = 0xCD
	CMD_FILE_CLOSE     uint8 = 0xCE
	CMD_FILE_PTR       uint8 = 0xCF
	CMD_FILE_INFO      uint8 = 0xD0
	CMD_FILE_CRC       uint8 = 0xD1
	CMD_FILE_DIR_MK    uint8 = 0xD2
	CMD_FILE_DEL       uint8 = 0xD3
	CMD_USB_RECOV      uint8 = 0xF0
	CMD_RUN_APP        uint8 = 0xF1
)

//
// General Serial
//

// CloseSerial closes serial connection.
//
// Waits 100ms after closing to avoid issues trying to
// reconnect to quickly.
func (n8 *N8) CloseSerial() {
	n8.Port.Close()
	time.Sleep(time.Millisecond * 100)
}

// InitSerial configures and opens serial connection.
//
// Waits 100ms after opening to avoid issues trying to
// reconnect to quickly.
func (n8 *N8) InitSerial(device string, timeout time.Duration) {
	n8.Address = device

	config := &serial.Config{
		Name:        n8.Address,
		Baud:        9600,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
		ReadTimeout: timeout,
	}

	var err error
	n8.Port, err = serial.OpenPort(config)
	if err != nil {
		log.Fatalf("Failed to open serial port: %v", err)
	}

	time.Sleep(time.Millisecond * 100)
}

//
// Serial Transmit
//

// TxData sends an arbitrary stream of data to the N8.
func (n8 *N8) TxData(buf []uint8) {
	_, err := n8.Port.Write(buf)
	if err != nil {
		log.Fatalf("Failed to write to serial port: %v", err)
	}
}

// Tx8 sends 8 bits to the N8.
func (n8 *N8) Tx8(arg uint8) {
	var buf []uint8 = make([]uint8, 1)
	buf[0] = (uint8)(arg)
	n8.TxData(buf)
}

// Tx8 sends 16 bits to the N8.
func (n8 *N8) Tx16(arg uint16) {
	var buf []uint8 = make([]uint8, 2)
	binary.LittleEndian.PutUint16(buf[:], arg)

	n8.TxData(buf)
}

// Tx32 sends 32 bits to the N8.
func (n8 *N8) Tx32(arg uint32) {
	var buf []uint8 = make([]uint8, 4)
	binary.LittleEndian.PutUint32(buf[:], arg)

	n8.TxData(buf)
}

// TxCmd sends a serial command to the N8.
//
// `n8.TxCmdExec()` is generally called after this to execute the command.
func (n8 *N8) TxCmd(command uint8) {
	cmd := make([]uint8, 4)
	cmd[0] = uint8('+')
	cmd[1] = uint8('+' ^ 0xff)
	cmd[2] = command
	cmd[3] = uint8(command ^ 0xff)

	_, err := n8.Port.Write(cmd)
	if err != nil {
		log.Fatalf("Failed to write to serial port: %v", err)
	}
}

// TxCmdExec sends an `EXEC` command to the N8.
//
// This is generally used after `n8.TxCmd()`.
func (n8 *N8) TxCmdExec() {
	n8.Tx8(CMD_EXEC)
}

// TxString sends string data to the N8.
//
// First, two bytes are sent indicating the length of the
// string in bytes. Next, the string itself is transmited.
func (n8 *N8) TxString(str string) {
	n8.Tx16((uint16)(len(str)))
	n8.TxData(([]uint8)(str))
}

// TxDataACK sends data in blocks with acks for each one.
//
// Sends data in blocks up to 1024 bytes long, checking the N8 status
// after each block is transmitted.
func (n8 *N8) TxDataACK(buf []uint8, length uint32) {
	var offset uint32 = 0
	var block uint32 = ACK_BLOCK_SIZE

	for length > 0 {
		if block > length {
			block = length
		}

		resp := n8.Rx8()
		if resp != 0 {

			log.Fatalf("[TxDataACK] bad ack: %02x", resp)
		}

		n8.TxData(buf[offset : offset+block])

		length -= block
		offset += block
	}
}

//
// Serial Read
//

// RxData reads data from the serial port into the provided buffer.
//
// It reads one uint8 at a time as reading too quickly causes issues.
func (n8 *N8) RxData(buf []uint8) {
	for remaining := len(buf); remaining > 0; {
		tinyBuf := make([]uint8, 1)

		n8.Port.Read(tinyBuf)
		copy(buf[len(buf)-remaining:], tinyBuf)

		remaining--
	}
}

// Rx8 reads 8 bits from the N8.
func (n8 *N8) Rx8() uint8 {
	buf := make([]uint8, 1)
	n8.RxData(buf)

	return (uint8)(buf[0])
}

// Rx16 reads 16 bits from the N8.
func (n8 *N8) Rx16() uint16 {
	buf := make([]uint8, 2)
	n8.RxData(buf)

	return binary.LittleEndian.Uint16(buf)
}

// Rx32 reads 32 bits from the N8.
func (n8 *N8) Rx32() uint32 {
	buf := make([]uint8, 4)
	n8.RxData(buf)

	return binary.LittleEndian.Uint32(buf)
}

// RxString reads string data from the N8.
//
// First, two bytes are received indicating the length of the
// string in bytes. Next, the string itself is read.
func (n8 *N8) RxString() string {
	len := n8.Rx16()
	buf := make([]uint8, len)
	n8.RxData(buf)
	return string(buf)
}

// RxFileInfo reads serialized FileInfo data from the N8.
func (n8 *N8) RxFileInfo() (uint32, uint16, uint16, uint8, string) {
	return n8.Rx32(), n8.Rx16(), n8.Rx16(), n8.Rx8(), n8.RxString()
}

//
// Misc
//

// GetStatus gets the N8 status.
//
// The high nibble should be 0xa5 if the status code was received
// successfully. The low nibble indicates the status.
func (n8 *N8) GetStatus() (bool, uint16) {
	n8.TxCmd(CMD_STATUS)
	resp := n8.Rx16()

	if (resp & 0xff00) != 0xa500 { // high nibble should be a5
		return false, resp
	}

	return true, resp & 0x00ff // low nibble returned as status code
}

// IsStatusOkay checks status code returned by the N8.
//
// A code of `0` is Ok. Other codes indicate specific errors.
func (n8 *N8) IsStatusOkay() (bool, uint16) {
	ok, resp := n8.GetStatus()
	if !ok {
		fmt.Printf("[isStatusOkay] could not read status: %04x\n", resp)
	}

	return resp == 0, resp
}
