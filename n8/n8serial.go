package n8

import (
	"encoding/binary"
	"fmt"
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

// SerialPortInterface abstracts the serial port to allow for
// testing with a mock interface.
type SerialPortInterface interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() (err error)
}

func (n8 *N8) Read(buf []byte) (bytesRead int, err error) {
	return n8.Port.Read(buf)
}
func (n8 *N8) Write(buf []byte) (bytesWritter int, err error) {
	return n8.Port.Write(buf)
}
func (n8 *N8) Close() (err error) {
	return n8.Port.Close()
}

// CloseSerial closes serial connection.
//
// Waits 100ms after closing to avoid issues trying to
// reconnect to quickly.
func (n8 *N8) CloseSerial() (err error) {
	err = n8.Port.Close()
	time.Sleep(time.Millisecond * 100)
	return
}

// InitSerial configures and opens serial connection.
//
// Waits 100ms after opening to avoid issues trying to
// reconnect to quickly.
func (n8 *N8) InitSerial(device string, timeout time.Duration) (err error) {
	n8.Address = device

	config := &serial.Config{
		Name:        n8.Address,
		Baud:        9600,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
		ReadTimeout: timeout,
	}

	n8.Port, err = serial.OpenPort(config)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %v", err)
	}

	time.Sleep(time.Millisecond * 100)

	return nil
}

//
// Serial Transmit
//

// TxData sends an arbitrary stream of data to the N8.
func (n8 *N8) TxData(buf []uint8) (err error) {
	if len(buf) == 0 {
		return fmt.Errorf("[TxData] no data to write")
	}

	n, err := n8.Write(buf)
	if err != nil {
		return
	}
	if n != len(buf) {
		return fmt.Errorf("[TxData] not all bytes written to serial port: expected %v, wrote %v", len(buf), n)
	}
	return nil
}

// Tx8 sends 8 bits to the N8.
func (n8 *N8) Tx8(arg uint8) (err error) {
	var buf []uint8 = make([]uint8, 1)
	buf[0] = (uint8)(arg)

	return n8.TxData(buf)
}

// Tx8 sends 16 bits to the N8.
func (n8 *N8) Tx16(arg uint16) (err error) {
	var buf []uint8 = make([]uint8, 2)
	binary.LittleEndian.PutUint16(buf[:], arg)

	return n8.TxData(buf)
}

// Tx32 sends 32 bits to the N8.
func (n8 *N8) Tx32(arg uint32) (err error) {
	var buf []uint8 = make([]uint8, 4)
	binary.LittleEndian.PutUint32(buf[:], arg)

	return n8.TxData(buf)
}

// TxCmd sends a serial command to the N8.
//
// `n8.TxCmdExec()` is generally called after this to execute the command.
func (n8 *N8) TxCmd(command uint8) (err error) {
	cmd := make([]uint8, 4)
	cmd[0] = uint8('+')
	cmd[1] = uint8('+' ^ 0xff)
	cmd[2] = command
	cmd[3] = uint8(command ^ 0xff)

	n, err := n8.Write(cmd)
	if err != nil {
		return
	}
	if n != len(cmd) {
		return fmt.Errorf("[TxData] not all bytes written to serial port: expected %v, wrote %v", len(cmd), n)
	}

	return nil
}

// TxCmdExec sends an `EXEC` command to the N8.
//
// This is generally used after `n8.TxCmd()`.
func (n8 *N8) TxCmdExec() (err error) {
	return n8.Tx8(CMD_EXEC)
}

// TxString sends string data to the N8.
//
// First, two bytes are sent indicating the length of the
// string in bytes. Next, the string itself is transmited.
func (n8 *N8) TxString(str string) (err error) {
	err = n8.Tx16((uint16)(len(str)))
	if err != nil {
		return
	}
	err = n8.TxData(([]uint8)(str))
	if err != nil {
		return
	}

	return nil
}

// TxStringFifo sends string data to the N8 FIFO.
//
// First, two bytes are sent indicating the length of the
// string in bytes. Next, the string itself is transmited.
func (n8 *N8) TxStringFifo(str string) (err error) {
	data := []byte(str)
	dataLength := make([]byte, 2)

	binary.LittleEndian.PutUint16(dataLength, uint16(len(data)))

	err = n8.FifoWr(dataLength, 2)
	if err != nil {
		return
	}
	return n8.FifoWr(data, (uint32)(len(data)))
}

// TxDataAck sends data in blocks with acks for each one.
//
// Sends data in blocks up to 1024 bytes long, checking the N8 status
// after each block is transmitted.
func (n8 *N8) TxDataAck(buf []uint8, length uint32) (err error) {
	var offset uint32 = 0
	var block uint32 = ACK_BLOCK_SIZE

	for length > 0 {
		if block > length {
			block = length
		}

		resp, err := n8.Rx8()
		if err != nil || resp != 0 {
			return fmt.Errorf("[TxDataAck] bad ack: %02x", resp)
		}
		err = n8.TxData(buf[offset : offset+block])
		if err != nil {
			return err
		}

		length -= block
		offset += block
	}

	return nil
}

//
// Serial Read
//

// RxData reads data from the serial port into the provided buffer.
//
// It reads one uint8 at a time as reading too quickly causes issues.
func (n8 *N8) RxData(buf []uint8) (err error) {
	var bytesRead int
	for remaining := len(buf); remaining > 0; {
		tinyBuf := make([]uint8, 1)

		n, err := n8.Read(tinyBuf)
		if err != nil || n != 1 {
			return err
		}

		copy(buf[len(buf)-remaining:], tinyBuf)

		bytesRead++
		remaining--
	}

	if bytesRead != len(buf) {
		return fmt.Errorf("[TxData] not all bytes read from serial port: expected %v, read %v", len(buf), bytesRead)
	}

	return nil
}

// Rx8 reads 8 bits from the N8.
func (n8 *N8) Rx8() (resp uint8, err error) {
	buf := make([]uint8, 1)
	err = n8.RxData(buf)
	if err != nil {
		return
	}

	return (uint8)(buf[0]), nil
}

// Rx16 reads 16 bits from the N8.
func (n8 *N8) Rx16() (resp uint16, err error) {
	buf := make([]uint8, 2)
	err = n8.RxData(buf)
	if err != nil {
		return
	}

	return binary.LittleEndian.Uint16(buf), nil
}

// Rx32 reads 32 bits from the N8.
func (n8 *N8) Rx32() (resp uint32, err error) {
	buf := make([]uint8, 4)
	err = n8.RxData(buf)
	if err != nil {
		return
	}

	return binary.LittleEndian.Uint32(buf), nil
}

// RxString reads string data from the N8.
//
// First, two bytes are received indicating the length of the
// string in bytes. Next, the string itself is read.
func (n8 *N8) RxString() (resp string, err error) {
	len, err := n8.Rx16()
	if err != nil {
		return
	}

	buf := make([]uint8, len)

	err = n8.RxData(buf)
	if err != nil {
		return
	}

	return string(buf), nil
}

// RxFileInfo reads serialized FileInfo data from the N8.
func (n8 *N8) RxFileInfo() (size uint32, date uint16, time uint16, attributes uint8, name string, err error) {

	size, err = n8.Rx32()
	if err != nil {
		return
	}
	date, err = n8.Rx16()
	if err != nil {
		return
	}
	time, err = n8.Rx16()
	if err != nil {
		return
	}
	attributes, err = n8.Rx8()
	if err != nil {
		return
	}
	name, err = n8.RxString()
	if err != nil {
		return
	}

	return
}

//
// Misc
//

// GetStatus gets the N8 status.
//
// The high nibble should be 0xa5 if the status code was received
// successfully. The low nibble indicates the status.
func (n8 *N8) GetStatus() (isOkay bool, statusCode uint16, err error) {
	n8.TxCmd(CMD_STATUS)
	resp, err := n8.Rx16()
	if err != nil {
		return
	}

	if (resp & 0xff00) != 0xa500 { // high nibble should be a5
		err = fmt.Errorf("[GetStatus] response %04x not a valid status code", resp)
		return false, resp, err
	}

	return true, resp & 0x00ff, nil // low nibble returned as status code
}

// IsStatusOkay checks status code returned by the N8.
//
// A code of `0` is Ok. Other codes indicate specific errors.
func (n8 *N8) IsStatusOkay() (isOkay bool, statusCode uint16, err error) {
	ok, status, err := n8.GetStatus()
	if err != nil {
		return
	}
	if !ok {
		fmt.Printf("[isStatusOkay] status is not okay: %04x\n", status)
	}

	return status == 0, status, nil
}
