package n8

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"forge.rights.ninja/jeff/goedlink/nesrom"
)

const CMD_PREFIX uint8 = 0x2A      // '*'
const CMD_TEST uint8 = 0x74        // 't'
const CMD_REBOOT uint8 = 0x72      // 'r'
const CMD_HALT uint8 = 0x68        // 'h'
const CMD_SELECT_GAME uint8 = 0x6E // 'n'
const CMD_RUN_GAME uint8 = 0x73    // 's'

//
// General Functions
//

// SendString sends a string to N8.
//
// Converts the string to a byte slice, convers its length to a little-endian
// byte slice, writes both to N8 FIFO.
func (n8 *N8) SendString(str string) (err error) {
	buf := []byte(str)
	length := make([]uint8, 2)
	binary.LittleEndian.PutUint16(length, (uint16)(len(buf)))

	err = n8.FifoWr(length, 2)
	if err != nil {
		return
	}
	err = n8.FifoWr(buf, (uint32)(len(buf)))
	if err != nil {
		return
	}

	return nil
}

// GetConfig retrieves the configuration from the N8.
//
// Reads the configuration data from the device's memory and constructs a
// MapConfig object from the binary data.
func (n8 *N8) GetConfig() (cfg *MapConfig, err error) {
	cfg = NewMapConfig()

	data := make([]uint8, len(cfg.GetSerialConfig()))

	err = n8.ReadMemory(ADDR_CFG, data, (uint32)(len(cfg.GetSerialConfig())))
	if err != nil {
		return
	}

	return NewMapConfigFromBinary(data), nil
}

// CopyFolder recursively copies a folder on the N8.
func (n8 *N8) CopyFolder(source string, destination string) (err error) {
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	if !strings.HasSuffix(destination, "/") {
		destination += "/"
	}

	dirs, err := getDirectories(source)
	if err != nil {
		return
	}
	for _, dir := range dirs {
		err = n8.CopyFolder(dir, destination+filepath.Base(dir))
		if err != nil {
			return
		}
	}

	files, err := getFiles(source)
	if err != nil {
		return
	}
	for _, file := range files {
		n8.CopyFile(file, destination+filepath.Base(file))
	}

	return
}

// getFiles returns files in a directory as a []string.
func getFiles(src string) ([]string, error) {
	files, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, filepath.Join(src, file.Name()))
		}
	}

	return fileList, nil
}

// getFiles returns directories in a directory as a []string.
func getDirectories(src string) ([]string, error) {
	files, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, filepath.Join(src, file.Name()))
		}
	}

	return dirs, nil
}

// CopyFolder copies a file on the N8.
//
// Prefix the source or destination string with `sd:` to specify
// a location on the N8 SD card.
func (n8 *N8) CopyFile(source string, destination string) (err error) {
	var sourceData []uint8

	source = strings.TrimSpace(source)
	destination = strings.TrimSpace(destination)

	if !strings.HasPrefix(strings.ToLower(source), "sd:") {
		fileInfo, err := os.Stat(source)
		if err == nil && fileInfo.IsDir() {
			n8.CopyFolder(source, destination)
			return err
		}
	}

	if strings.HasSuffix(destination, "/") || strings.HasSuffix(destination, "\\") {
		destination += filepath.Base(destination)
	}

	if strings.HasPrefix(strings.ToLower(source), "sd:") {
		source = source[3:]
		fileInfo, err := n8.GetFileInfo(source)
		if err != nil {
			return err
		}

		sourceData = make([]uint8, fileInfo.Size)

		err = n8.OpenFile(source, FAT_READ)
		if err != nil {
			return err
		}
		err = n8.ReadFile(sourceData, (uint32)(len(sourceData)))
		if err != nil {
			return err
		}
		err = n8.CloseFile()
		if err != nil {
			return err
		}
	} else {
		sourceData, err = os.ReadFile(source)
		if err != nil {
			return
		}
	}

	if strings.HasPrefix(strings.ToLower(destination), "sd:") {
		destination = destination[3:]

		err = n8.OpenFile(destination, FAT_CREATE_ALWAYS|FAT_WRITE)
		if err != nil {
			return
		}
		err = n8.FileWrite(sourceData, (uint32)(len(sourceData)))
		if err != nil {
			return
		}
		err = n8.CloseFile()
		if err != nil {
			return
		}
	} else {
		err = os.WriteFile(destination, sourceData, 0644)
		if err != nil {
			return
		}
	}

	return
}

// SetConfig sets the configuration on the N8.
//
// Writes the config binary to the N8's memory.
func (n8 *N8) SetConfig(config *MapConfig) (err error) {
	buf := config.GetSerialConfig()
	return n8.WriteMemory(ADDR_CFG, buf, (uint32)(len(buf)))
}

// Command sends a command to the N8.
//
// Constructs a command buffer with a prefix and the command, and writes
// it to the N8 FIFO.
func (n8 *N8) Command(command uint8) (err error) {
	buf := make([]uint8, 2)
	buf[0] = CMD_PREFIX
	buf[1] = command

	return n8.FifoWr(buf, (uint32)(len(buf)))
}

// SelectGame selects a game on the N8.
//
// Returns the game index received from the device.
func (n8 *N8) SelectGame(path string) (index uint16, err error) {
	n8.Command(CMD_SELECT_GAME)
	n8.TxStringFifo(path)

	resp, err := n8.Rx8()
	if err != nil {
		return
	}
	if resp != 0 {
		err = fmt.Errorf("[SelectGame] CMD_SELECT_GAME returned bad status code: %v", resp)
		return
	}

	return n8.Rx16()
}

// getTestMapper retrieves the path for a test mapper.
//
// Reads the MAPROUT.BIN file to determine the mapper, and return a
// path to the correct rbf.
func getTestMapper(mapper uint16) (path string, err error) {
	mapPath := "./maps/"
	maprout, err := os.ReadFile("MAPROUT.BIN")
	if err != nil {
		return
	}

	pack := maprout[mapper]
	if pack == 255 && mapper != 255 {
		log.Fatalln("[getTestMapper] mapper is not supported")
	}

	if pack < 100 {
		mapPath += "0"
	}
	if pack < 10 {
		mapPath += "0"
	}

	return mapPath + fmt.Sprintf("%d", pack) + ".RBF", nil
}

// MakeDir creates a directory on the N8.
//
// Trims whitespace from the path, ensures it starts with "sd:",
// creates the directory on the device.
func (n8 *N8) MakeDir(path string) (err error) {
	path = strings.TrimSpace(path)

	if !strings.HasPrefix(strings.ToLower(path), "sd:") {
		err = fmt.Errorf("[MakeDir] incorrect dir path: " + path + "\nValid paths must start with sd:")
		return
	}

	err = n8.mkdir(path[3:])
	if err != nil {
		return
	}

	return nil
}

// Halt halts the N8.
func (n8 *N8) Halt() (err error) {
	config := NewMapConfig()

	config.MapIndex = 0xff
	n8.SetConfig(config)
	n8.Command(CMD_HALT)

	resp, err := n8.Rx8()
	if err != nil {
		return
	}
	if resp != 0 {
		return fmt.Errorf("[Halt] unexpected response to USB halt")
	}

	return
}

// HaltExit sets the N8 configuration to exit halt mode.
func (n8 *N8) HaltExit() (err error) {
	config := NewMapConfig()
	config.MapIndex = 0xff
	config.Ctrl = CTRL_UNLOCK

	return n8.SetConfig(config)
}

// MapLoadSDC inits the FPGA with the correct mapper.
//
// Reads map data from `EDN8/MAPROUT.BIN` on N8 SD card,
// then loads the FPGA with the correct `*.RBF` from within
// `EDN8/MAPS/`.
func (n8 *N8) MapLoadSDC(mapId uint8, config *MapConfig) (err error) {
	mapRout := make([]uint8, 4096)

	mapPath := "EDN8/MAPS/"

	err = n8.OpenFile("EDN8/MAPROUT.BIN", FAT_READ)
	if err != nil {
		return
	}
	err = n8.ReadFile(mapRout, (uint32)(len(mapRout)))
	if err != nil {
		return
	}
	err = n8.CloseFile()
	if err != nil {
		return
	}

	mapPkg := mapRout[mapId]

	if mapPkg == 0xff && mapId != 0xff {
		var config MapConfig
		config.MapIndex = 255
		config.Ctrl = CTRL_UNLOCK
		n8.FpgaInitFromSD("EDN8/MAPS/255.RBF", &config)
		log.Fatalf("[MapLoadSDC] unsupported mapper: %d", mapId) // TODO: determine why we're doing the above if we're going to throw a fatal error
	}

	if mapPkg < 100 {
		mapPath += "0"
	}
	if mapPkg < 10 {
		mapPath += "0"
	}
	mapPath += strconv.Itoa((int)(mapPkg)) + ".RBF"

	return n8.FpgaInitFromSD(mapPath, config)
}

// Reboot sends a reboot command to the N8.
func (n8 *N8) Reboot() (err error) {
	err = n8.Command(CMD_REBOOT)
	if err != nil {
		return
	}
	resp, err := n8.Rx8()
	if err != nil {
		return
	}
	if resp != 0 {
		err = fmt.Errorf("[Reboot] unknown response code: %04x", resp)
		return
	}

	return n8.ExitServiceMode()
}

// LoadOS loads an OS ROM.
//
// Initializes the FPGA with provided OS ROM.
func (n8 *N8) LoadOS(rom *nesrom.NesRom, mapPath string) (err error) {
	if mapPath == "" {
		mapPath, err = getTestMapper(255)
		if err != nil {
			return
		}
	}

	var config MapConfig
	config.MapIndex = 0xff
	config.Ctrl = CTRL_UNLOCK
	config.Serialize()

	err = n8.Command(CMD_REBOOT)
	if err != nil {
		return
	}
	err = n8.TxCmdExec()
	if err != nil {
		return
	}
	err = n8.WriteMemory(rom.GetPrgAddr(), rom.GetPrgData(), rom.GetPrgSize())
	if err != nil {
		return
	}
	err = n8.WriteMemory(rom.GetChrAddr(), rom.GetChrData(), rom.GetChrSize())
	if err != nil {
		return
	}

	_, _, err = n8.GetStatus()
	if err != nil { // TODO: determine if I need to actually check status here
		return
	}

	if mapPath == "" {
		if n8.MapLoadSDC(255, &config) != nil {
			return
		}
	} else {
		file, err := os.ReadFile(mapPath)
		if err != nil {
			return err
		}

		err = n8.FpgaInit(file, &config)
		if err != nil {
			return err
		}
	}

	return
}

// LoadGame loads a new game on the N8.
//
// Creates a `usb_games` directory for USB games and writes the ROM
// and optional mapper `*.RBF` to it. It then selects the game and
// runs it.
func (n8 *N8) LoadGame(romPath string, mapPath string) (err error) {
	directory := "usb_games"
	err = n8.MakeDir("sd:" + directory)
	if err != nil {
		return
	}

	romDestinationPath := directory + "/" + filepath.Base(romPath)
	fileData, err := os.ReadFile(romPath)
	if err != nil {
		return
	}

	err = n8.OpenFile(romDestinationPath, FAT_CREATE_ALWAYS|FAT_WRITE)
	if err != nil {
		return
	}
	err = n8.FileWrite(fileData, (uint32)(len(fileData)))
	if err != nil {
		return
	}
	err = n8.CloseFile()
	if err != nil {
		return
	}

	_, err = n8.SelectGame(romDestinationPath)
	if err != nil {
		return
	}
	// mapIndex, err := n8.SelectGame(romDestinationPath)
	// if err != nil {
	// 	return
	// }
	// if mapPath == "" {
	// 	mapPath = getTestMapper(mapIndex)
	// }

	rbfDestinationPath := changeExtension(romDestinationPath, "rbf")

	if mapPath != "" {
		fileData, err = os.ReadFile(mapPath)
		if err != nil {
			return
		}
		err = n8.OpenFile(rbfDestinationPath, FAT_WRITE|FAT_CREATE_ALWAYS)
		if err != nil {
			return
		}
		err = n8.FileWrite(fileData, (uint32)(len(fileData)))
		if err != nil {
			return
		}
		err = n8.CloseFile()
		if err != nil {
			return
		}
	} else {
		err = n8.DeleteRecord(rbfDestinationPath)
		if err != nil {
			return
		}
	}

	return n8.Command(CMD_RUN_GAME)
}

//
// Hardware Functions
//

// GetVdc retrieves the VDC state from the N8.
func (n8 *N8) GetVdc() (v *Vdc, err error) {
	buf := make([]uint8, VDC_DATA_SIZE)

	err = n8.TxCmd(CMD_GET_VDC)
	if err != nil {
		return
	}
	err = n8.RxData(buf)
	if err != nil {
		return
	}

	return NewVdc(buf)
}

// GetRtc retrieves the RTC (Real-Time Clock) time from the N8.
func (n8 *N8) GetRtc() (time *RtcTime, err error) {
	buf := make([]uint8, RTC_DATA_SIZE)

	err = n8.TxCmd(CMD_RTC_GET)
	if err != nil {
		return
	}
	err = n8.RxData(buf)
	if err != nil {
		return
	}

	return NewRtcTimeFromSerial(buf)
}

// SetRtc sets the RTC (Real-Time Clock) time on the N8.
func (n8 *N8) SetRtc(time time.Time) (err error) {
	rtcTime := NewRtcTime(time)

	err = n8.TxCmd(CMD_RTC_SET)
	if err != nil {
		return
	}

	return n8.TxData(rtcTime.GetVals())
}

//
// Modes
//

// EnterServiceMode switches the N8 to service mode.
//
// Checks if the device is already in service mode. If not, it performs
// a hard reset, waits for the device to boot, and verifies that it has
// successfully entered service mode.
func (n8 *N8) EnterServiceMode() (err error) {
	isServiceMode, err := n8.isServiceMode()
	if err != nil {
		return err
	}
	if isServiceMode {
		return nil
	}

	err = n8.TxCmd(CMD_HARD_RESET)
	if err != nil {
		return
	}
	err = n8.TxCmdExec()
	if err != nil {
		return
	}
	err = n8.bootWait()
	if err != nil {
		return
	}

	isServiceMode, err = n8.isServiceMode()
	if err != nil {
		return err
	}
	if isServiceMode {
		err = fmt.Errorf("[EnterServiceMode] device stuck in app mode")
		return
	}

	return
}

// ExitServiceMode switches the N8 out of service mode.
//
// Checks if the device is currently in service mode. If so, it sends
// the command to switch to app mode.
func (n8 *N8) ExitServiceMode() (err error) {
	isServicemode, err := n8.isServiceMode()
	if err != nil {
		return err
	}
	if !isServicemode {
		return nil
	}

	err = n8.TxCmd(CMD_RUN_APP)
	if err != nil {
		return
	}
	err = n8.bootWait()
	if err != nil {
		return
	}

	isServicemode, err = n8.isServiceMode()
	if err != nil {
		return err
	}

	if isServicemode {
		err = fmt.Errorf("[ExitServiceMode] device stuck in service mode")
		return
	}

	return nil
}

// isServiceMode checks if the n8serial device is in service mode.
func (n8 *N8) isServiceMode() (b bool, err error) {
	err = n8.TxCmd(CMD_GET_MODE)
	if err != nil {
		return
	}

	resp, err := n8.Rx8()
	return resp == 0xA1, err
}

// Recovery performs a recovery operation on the N8.
//
// Checks if the device is in service mode, reads the current core CRC
// from flash, initiates the USB recovery process, verifies the
// recovery status.
func (n8 *N8) Recovery() (err error) {
	isServiceMode, err := n8.isServiceMode()
	if err != nil {
		return
	}
	if isServiceMode {
		err = fmt.Errorf("[Recovery] not in service mode")
		return
	}

	err = n8.Port.Close()
	if err != nil {
		return
	}
	err = n8.InitSerial(n8.Address, time.Second*8)
	if err != nil {
		return
	}

	crc := make([]byte, 4)
	err = n8.ReadFlash(ADDR_FLA_ICOR, crc, 4)
	if err != nil {
		return
	}

	err = n8.TxCmd(CMD_USB_RECOV)
	if err != nil {
		return
	}
	err = n8.Tx32(ADDR_FLA_ICOR)
	if err != nil {
		return
	}
	err = n8.TxData(crc)
	if err != nil {
		return
	}

	ok, status, err := n8.GetStatus()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[Recovery] unknown status code: %04x", status)
		return
	}

	err = n8.Port.Close()
	if err != nil {
		return
	}
	err = n8.InitSerial(n8.Address, time.Second*2)
	if err != nil {
		return
	}

	if status == 0x0088 {
		log.Fatalln("[Recovery] current core matches to recovery copy")
	} else if status == 0x0000 {
		log.Fatalf("[Recovery] recovery error %02x", status)
	}

	return
}

// bootWait waits for the N8 to boot.
func (n8 *N8) bootWait() (err error) {
	for i := 0; i < 10; i++ {

		err = n8.CloseSerial()
		if err != nil {
			return
		}
		time.Sleep(time.Millisecond * 100)
		err = n8.InitSerial(n8.Address, time.Second*2)
		if err != nil {
			return
		}
		time.Sleep(time.Millisecond * 100)

		ok, _, err := n8.GetStatus()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}

	return fmt.Errorf("[BootWait] boot timeout")
}

//
// Memory Functions
//

// FifoWr writes data to the FIFO on the N8 device.
//
// Writes the provided data to the FIFO buffer at the specified address.
func (n8 *N8) FifoWr(buf []uint8, length uint32) error {
	return n8.WriteMemory(ADDR_FIFO, buf, length)
}

// MemoryCrc calculates the CRC of specific memory data on the N8.
//
// Sends a command to calculate the CRC of the specified memory region
// and returns the CRC value received from the device.
func (n8 *N8) MemoryCrc(addr uint32, length uint32) (crc uint32, err error) {
	err = n8.TxCmd(CMD_MEM_CRC)
	if err != nil {
		return
	}
	err = n8.Tx32(addr)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}
	err = n8.Tx32(CRC_INIT_VAL)
	if err != nil {
		return
	}
	err = n8.TxCmdExec()
	if err != nil {
		return
	}

	return n8.Rx32()
}

// MemorySet sets a byte value in memory on the N8.
//
// Sends a command to set the specified byte value in the memory region
// starting at the specified address.
func (n8 *N8) MemorySet(addr uint32, val uint8, length uint32) (err error) {
	err = n8.TxCmd(CMD_MEM_SET)
	if err != nil {
		return
	}
	err = n8.Tx32(addr)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}
	err = n8.Tx8(val)
	if err != nil {
		return
	}
	err = n8.TxCmdExec()
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[MemorySet] unknown response code: %04x", resp)
	}

	return
}

// MemoryTest tests if a specific byte value exists in memory on the N8.
//
// Sends a command to test the memory region starting at the specified address,
// returns 8-bit response.
func (n8 *N8) MemoryTest(addr uint32, val uint8, length uint32) (b bool, err error) {
	err = n8.TxCmd(CMD_MEM_TST)
	if err != nil {
		return
	}
	err = n8.Tx32(addr)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}
	err = n8.Tx8(val)
	if err != nil {
		return
	}
	err = n8.TxCmdExec()
	if err != nil {
		return
	}

	resp, err := n8.Rx8()
	return resp != 0, err
}

// ReadFlash reads data from flash memory on the N8.
//
// Sends a command to read data from flash memory, starting at the specified address,
// into the provided uint8 slice.
func (n8 *N8) ReadFlash(addr uint32, buf []uint8, length uint32) (err error) {
	err = n8.TxCmd(CMD_FLA_RD)
	if err != nil {
		return
	}
	err = n8.Tx32(addr)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}

	return n8.RxData(buf)
}

// ReadMemory reads data from memory on the N8.
//
// Reads data from memory in chunks into the provided uint8 slice.
func (n8 *N8) ReadMemory(addr uint32, buf []uint8, length uint32) (err error) {
	if length == 0 {
		return fmt.Errorf("[ReadMemory] length cannot be 0")
	}

	// I've seen some issues when reading too much at a time, so I'm
	// reading a max of 32 uint8s at a time.
	const chunkSize uint32 = 0x20
	for length > 0 {
		currentChunk := chunkSize
		if length < chunkSize {
			currentChunk = length
		}
		tempBuf := make([]uint8, currentChunk)

		// transmit command & params
		err = n8.TxCmd(CMD_MEM_RD)
		if err != nil {
			return
		}
		err = n8.Tx32(addr)
		if err != nil {
			return
		}
		err = n8.Tx32(currentChunk)
		if err != nil {
			return
		}
		err = n8.TxCmdExec()
		if err != nil {
			return
		}

		// receive data
		err = n8.RxData(tempBuf)
		if err != nil {
			return
		}

		copy(buf[:currentChunk], tempBuf)
		buf = buf[currentChunk:]

		addr += currentChunk
		length -= currentChunk
	}

	return
}

// WriteFlash writes data to flash memory on the N8.
//
// Sends a command to write data to flash memory starting at the specified address,
// verifies the status after operation.
func (n8 *N8) WriteFlash(addr uint32, buf []uint8, length uint32) (err error) {
	err = n8.TxCmd(CMD_FLA_WR)
	if err != nil {
		return
	}
	err = n8.Tx32(addr)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}
	err = n8.TxDataAck(buf, length)
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[WriteFlash] unknown status code: %04x", resp)
	}

	return
}

// WriteMemory writes data to memory on the N8.
//
// Sends a command to write data to memory starting at the specified address,
// writes the data from the provided uint8 slice.
func (n8 *N8) WriteMemory(addr uint32, buf []uint8, length uint32) (err error) {
	if length == 0 {
		return fmt.Errorf("[WriteMemory] length of data was 0")
	}

	err = n8.TxCmd(CMD_MEM_WR)
	if err != nil {
		return
	}
	err = n8.Tx32(addr)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}
	err = n8.TxCmdExec()
	if err != nil {
		return
	}
	err = n8.TxData(buf)
	if err != nil {
		return
	}

	return
}

//
// FPGA Functions
//

// fpgaPostInit checks the N8 status after FPGA init and writes config to memory
func (n8 *N8) fpgaPostInit(config *MapConfig) (err error) {
	if config == nil {
		err = fmt.Errorf("[fpgaPostInit] config is invalid (it's nil)")
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[fpgaPostInit] unknown status code: %v", resp)
		return
	}

	return n8.WriteMemory(ADDR_CFG, config.GetSerialConfig(), (uint32)(len(config.GetSerialConfig())))
}

// FpgaInit initializes the N8 FPGA.
//
// Initlizes the FPGA with data from `[]uint8`,
// along with initialization data.
func (n8 *N8) FpgaInit(buf []uint8, config *MapConfig) (err error) {
	err = n8.TxCmd(CMD_FPGA_USB)
	if err != nil {
		return
	}
	err = n8.Tx32(uint32(len(buf)))
	if err != nil {
		return
	}
	err = n8.TxDataAck(buf, uint32(len(buf)))
	if err != nil {
		return
	}

	return n8.fpgaPostInit(config)
}

// FpgaInitFromFlash initializes the N8 FPGA.
//
// Initlizes the FPGA with data from at address in flash,
// along with initialization data.
func (n8 *N8) FpgaInitFromFlash(address uint32, config *MapConfig) (err error) {
	err = n8.TxCmd(CMD_FPGA_FLA)
	if err != nil {
		return
	}
	err = n8.Tx32(address)
	if err != nil {
		return
	}
	err = n8.TxCmdExec()
	if err != nil {
		return
	}

	return n8.fpgaPostInit(config)
}

// FpgaInitFromSD initializes the N8 FPGA using data from an SD card.
//
// Initlizes the FPGA with data on SD card, along with
// initialization data.
func (n8 *N8) FpgaInitFromSD(path string, config *MapConfig) (err error) {
	fileinfo, err := n8.GetFileInfo(path)
	if err != nil {
		return
	}

	err = n8.OpenFile(path, FAT_READ)
	if err != nil {
		return
	}
	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[FpgaInitFromSD] unknown status code: %04x", resp)
		return
	}

	err = n8.TxCmd(CMD_FPGA_SDC)
	if err != nil {
		return
	}
	err = n8.Tx32(fileinfo.Size)
	if err != nil {
		return
	}
	err = n8.TxCmdExec()
	if err != nil {
		return
	}

	return n8.fpgaPostInit(config)
}
