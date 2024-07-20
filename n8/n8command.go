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
func (n8 *N8) SendString(str string) {
	buf := []byte(str)
	length := make([]uint8, 2)
	binary.LittleEndian.PutUint16(length, (uint16)(len(buf)))

	n8.FifoWr(length, 2)
	n8.FifoWr(buf, (uint32)(len(buf)))
}

// GetConfig retrieves the configuration from the N8.
//
// Reads the configuration data from the device's memory and constructs a
// MapConfig object from the binary data.
func (n8 *N8) GetConfig() *MapConfig {
	var cfg *MapConfig = NewMapConfig()

	data := make([]uint8, len(cfg.GetSerialConfig()))

	n8.ReadMemory(ADDR_CFG, data, (uint32)(len(cfg.GetSerialConfig())))

	return NewMapConfigFromBinary(data)
}

// CopyFolder recursively copies a folder on the N8.
func (n8 *N8) CopyFolder(source string, destination string) {
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	if !strings.HasSuffix(destination, "/") {
		destination += "/"
	}

	dirs, err := getDirectories(source)
	if err != nil {
		log.Fatal(err)
	}
	for _, dir := range dirs {
		n8.CopyFolder(dir, destination+filepath.Base(dir))
	}

	files, err := getFiles(source)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		n8.CopyFile(file, destination+filepath.Base(file))
	}
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
func (n8 *N8) CopyFile(source string, destination string) {
	var sourceData []uint8
	var err error

	source = strings.TrimSpace(source)
	destination = strings.TrimSpace(destination)

	if !strings.HasPrefix(strings.ToLower(source), "sd:") {
		fileInfo, err := os.Stat(source)
		if err == nil && fileInfo.IsDir() {
			n8.CopyFolder(source, destination)
			return
		}
	}

	if strings.HasSuffix(destination, "/") || strings.HasSuffix(destination, "\\") {
		destination += filepath.Base(destination)
	}

	if strings.HasPrefix(strings.ToLower(source), "sd:") {
		source = source[3:]
		sourceData = make([]uint8, n8.GetFileInfo(source).Size)

		n8.OpenFile(source, FAT_READ)
		n8.ReadFile(sourceData, (uint32)(len(sourceData)))
		n8.CloseFile()
	} else {
		sourceData, err = os.ReadFile(source)
		if err != nil {
			log.Fatalln("[CopyFile] error reading source:", err)
		}
	}

	if strings.HasPrefix(strings.ToLower(destination), "sd:") {
		destination = destination[3:]

		n8.OpenFile(destination, FAT_CREATE_ALWAYS|FAT_WRITE)
		n8.FileWrite(sourceData, (uint32)(len(sourceData)))
		n8.CloseFile()
	} else {
		err = os.WriteFile(destination, sourceData, 0644)
		if err != nil {
			log.Fatalln("[CopyFile] error writing to destination:", err)
		}
	}
}

// SetConfig sets the configuration on the N8.
//
// Writes the config binary to the N8's memory.
func (n8 *N8) SetConfig(config *MapConfig) {
	buf := config.GetSerialConfig()
	n8.WriteMemory(ADDR_CFG, buf, (uint32)(len(buf)))
}

// Command sends a command to the N8.
//
// Constructs a command buffer with a prefix and the command, and writes
// it to the N8 FIFO.
func (n8 *N8) Command(command uint8) {
	buf := make([]uint8, 2)

	buf[0] = CMD_PREFIX
	buf[1] = command
	n8.FifoWr(buf, (uint32)(len(buf)))
}

// SelectGame selects a game on the N8.
//
// Returns the game index received from the device.
func (n8 *N8) SelectGame(path string) uint16 {
	n8.Command(CMD_SELECT_GAME)
	n8.TxString(path)

	resp := n8.Rx8()
	if resp != 0 {
		log.Fatalf("[SelectGame] game select error: %v", resp)
	}

	time.Sleep(time.Second * 2)

	return n8.Rx16()
}

// getTestMapper retrieves the path for a test mapper.
//
// Reads the MAPROUT.BIN file to determine the mapper, and return a
// path to the correct rbf.
func getTestMapper(mapper uint16) string {
	home := "./maps/"                        // TODO: read these from N8
	maprout, _ := os.ReadFile("MAPROUT.BIN") // TODO: read from N8
	pack := maprout[mapper]

	if pack == 255 && mapper != 255 {
		log.Fatalln("[getTestMapper] mapper is not supported")
	}

	mapPath := home

	if pack < 100 {
		mapPath += "0"
	}
	if pack < 10 {
		mapPath += "0"
	}

	return mapPath + fmt.Sprintf("%d", pack) + ".RBF"
}

// MakeDir creates a directory on the N8.
//
// Trims whitespace from the path, ensures it starts with "sd:",
// creates the directory on the device.
func (n8 *N8) MakeDir(path string) {
	path = strings.TrimSpace(path)

	if !strings.HasPrefix(strings.ToLower(path), "sd:") {
		log.Fatalln("[MakeDir] incorrect dir path: " + path + "\nValid paths must start with sd:")
	}

	n8.mkdir(path[3:])
}

// Halt halts the N8.
func (n8 *N8) Halt() {
	config := NewMapConfig()

	config.MapIndex = 0xff
	n8.SetConfig(config)
	n8.Command(CMD_HALT)

	if n8.Rx8() != 0 {
		log.Fatalln("[Halt] Unexpected response to USB halt")
	}
}

// HaltExit sets the N8 configuration to exit halt mode.
func (n8 *N8) HaltExit() {
	config := NewMapConfig()
	config.MapIndex = 0xff
	config.Ctrl = CTRL_UNLOCK

	n8.SetConfig(config)
}

// MapLoadSDC inits the FPGA with the correct mapper.
//
// Reads map data from `EDN8/MAPROUT.BIN` on N8 SD card,
// then loads the FPGA with the correct `*.RBF` from within
// `EDN8/MAPS/`.
func (n8 *N8) MapLoadSDC(mapId uint8, config *MapConfig) {
	mapRout := make([]uint8, 4096)

	mapPath := "EDN8/MAPS/"

	n8.OpenFile("EDN8/MAPROUT.BIN", FAT_READ)
	n8.ReadFile(mapRout, (uint32)(len(mapRout)))
	n8.CloseFile()

	mapPkg := mapRout[mapId]

	if mapPkg == 0xff && mapId != 0xff {
		var config MapConfig
		config.MapIndex = 255
		config.Ctrl = CTRL_UNLOCK
		n8.FpgaInitFromSD("EDN8/MAPS/255.RBF", &config)
		log.Fatalf("[MapLoadSDC] unsupported mapper: %d", mapId)
	}

	if mapPkg < 100 {
		mapPath += "0"
	}
	if mapPkg < 10 {
		mapPath += "0"
	}

	mapPath += strconv.Itoa((int)(mapPkg)) + ".RBF"
	n8.FpgaInitFromSD(mapPath, config)
}

// Reboot sends a reboot command to the N8.
func (n8 *N8) Reboot() {
	n8.Command(CMD_REBOOT)
	n8.Rx8()

	n8.ExitServiceMode()
}

// LoadOS loads an OS ROM.
//
// Initializes the FPGA with provided OS ROM.
func (n8 *N8) LoadOS(rom *nesrom.NesRom, mapPath string) {
	if mapPath == "" {
		mapPath = getTestMapper(255)
	}

	var config MapConfig
	config.MapIndex = 0xff
	config.Ctrl = CTRL_UNLOCK
	config.Serialize()

	n8.Command(CMD_REBOOT)
	n8.TxCmdExec()

	n8.WriteMemory(rom.GetPrgAddr(), rom.GetPrgData(), rom.GetPrgSize())
	n8.WriteMemory(rom.GetChrAddr(), rom.GetChrData(), rom.GetChrSize())

	n8.GetStatus()

	if mapPath == "" {
		n8.MapLoadSDC(255, &config)
	} else {
		file, err := os.ReadFile(mapPath)
		if err != nil {
			log.Fatalf("[LoadOS] error reading map file %s: %v", mapPath, err)
		}

		n8.FpgaInit(file, &config)
	}
}

// LoadGame loads a new game on the N8.
//
// Creates a `usb_games` directory for USB games and writes the ROM
// and optional mapper `*.RBF` to it. It then selects the game and
// runs it.
func (n8 *N8) LoadGame(romPath string, mapPath string) {
	directory := "usb_games"
	n8.MakeDir("sd:" + directory)

	romDestinationPath := directory + "/" + filepath.Base(romPath)
	fileData, err := os.ReadFile(romPath)
	if err != nil {
		log.Fatalln("[LoadGame] error reading source:", err)
	}

	n8.OpenFile(romDestinationPath, FAT_CREATE_ALWAYS|FAT_WRITE)
	n8.FileWrite(fileData, (uint32)(len(fileData)))
	n8.CloseFile()

	n8.SelectGame(romDestinationPath)

	// mapIndex := n8.SelectGame(romDestinationPath)
	// if mapPath == "" {
	// 	mapPath = getTestMapper(mapIndex)
	// }

	rbfDestinationPath := changeExtension(romDestinationPath, "rbf")

	if mapPath != "" {
		fileData, _ = os.ReadFile(mapPath) // TODO: error checking
		n8.OpenFile(rbfDestinationPath, FAT_WRITE|FAT_CREATE_ALWAYS)
		n8.FileWrite(fileData, (uint32)(len(fileData)))
		n8.CloseFile()
	} else {
		n8.DeleteRecord(rbfDestinationPath)
	}

	n8.Command(CMD_RUN_GAME)
}

//
// Hardware Functions
//

// GetVdc retrieves the VDC state from the N8.
func (n8 *N8) GetVdc() *Vdc {
	buf := make([]uint8, VDC_DATA_SIZE)

	n8.TxCmd(CMD_GET_VDC)
	n8.RxData(buf)

	return NewVdc(buf)
}

// GetRtc retrieves the RTC (Real-Time Clock) time from the N8.
func (n8 *N8) GetRtc() *RtcTime {
	buf := make([]uint8, RTC_DATA_SIZE)

	n8.TxCmd(CMD_RTC_GET)
	n8.RxData(buf)

	return NewRtcTimeFromSerial(buf)
}

// SetRtc sets the RTC (Real-Time Clock) time on the N8.
func (n8 *N8) SetRtc(time time.Time) {
	rtcTime := NewRtcTime(time)

	n8.TxCmd(CMD_RTC_SET)
	n8.TxData(rtcTime.GetVals())
}

//
// Modes
//

// EnterServiceMode switches the N8 to service mode.
//
// Checks if the device is already in service mode. If not, it performs
// a hard reset, waits for the device to boot, and verifies that it has
// successfully entered service mode.
func (n8 *N8) EnterServiceMode() {
	if n8.isServiceMode() {
		return
	}

	n8.TxCmd(CMD_HARD_RESET)
	n8.TxCmdExec()
	n8.bootWait()

	if !n8.isServiceMode() {
		log.Fatalln("[EnterServiceMode] device stuck in app mode")
	}
}

// ExitServiceMode switches the N8 out of service mode.
//
// Checks if the device is currently in service mode. If so, it sends
// the command to switch to app mode.
func (n8 *N8) ExitServiceMode() {
	if !n8.isServiceMode() {
		return
	}

	n8.TxCmd(CMD_RUN_APP)
	n8.bootWait()

	if n8.isServiceMode() {
		log.Fatalln("[ExitServiceMode] device stuck in service mode")
	}
}

// isServiceMode checks if the n8serial device is in service mode.
func (n8 *N8) isServiceMode() bool {
	n8.TxCmd(CMD_GET_MODE)
	resp := n8.Rx8()

	return resp == 0xA1
}

// Recovery performs a recovery operation on the N8.
//
// Checks if the device is in service mode, reads the current core CRC
// from flash, initiates the USB recovery process, verifies the
// recovery status.
func (n8 *N8) Recovery() {
	if !n8.isServiceMode() {
		log.Fatalln("[Recovery] not in service mode")
	}

	n8.Port.Close()
	n8.InitSerial(n8.Address, time.Second*8)

	crc := make([]byte, 4)
	n8.ReadFlash(ADDR_FLA_ICOR, crc, 4)

	n8.TxCmd(CMD_USB_RECOV)
	n8.Tx32(ADDR_FLA_ICOR)
	n8.TxData(crc)

	ok, status := n8.GetStatus()
	if !ok {
		log.Fatalf("[Recovery] status error: %v", status)
	}

	n8.Port.Close()
	n8.InitSerial(n8.Address, time.Second*2)

	if status == 0x0088 {
		log.Fatalln("[Recovery] current core matches to recovery copy")
	} else if status == 0x0000 {
		log.Fatalf("[Recovery] recovery error %02x", status)
	}
}

// bootWait waits for the N8 to boot.
func (n8 *N8) bootWait() {
	for i := 0; i < 10; i++ {

		n8.CloseSerial()
		time.Sleep(time.Millisecond * 100)
		n8.InitSerial(n8.Address, time.Second*2)
		time.Sleep(time.Millisecond * 100)

		ok, _ := n8.GetStatus()
		if ok {
			return
		}
	}

	log.Fatalln("[BootWait] boot timeout")
}

//
// Memory Functions
//

// FifoWr writes data to the FIFO on the N8 device.
//
// Writes the provided data to the FIFO buffer at the specified address.
func (n8 *N8) FifoWr(buf []uint8, length uint32) {
	n8.WriteMemory(ADDR_FIFO, buf, length)
}

// MemoryCrc calculates the CRC of specific memory data on the N8.
//
// Sends a command to calculate the CRC of the specified memory region
// and returns the CRC value received from the device.
func (n8 *N8) MemoryCrc(addr uint32, length uint32) uint32 {
	n8.TxCmd(CMD_MEM_CRC)
	n8.Tx32(addr)
	n8.Tx32(length)
	n8.Tx32(CRC_INIT_VAL)
	n8.TxCmdExec()

	return n8.Rx32()
}

// MemorySet sets a byte value in memory on the N8.
//
// Sends a command to set the specified byte value in the memory region
// starting at the specified address.
func (n8 *N8) MemorySet(addr uint32, val uint8, length uint32) {
	n8.TxCmd(CMD_MEM_SET)
	n8.Tx32(addr)
	n8.Tx32(length)
	n8.Tx8(val)
	n8.TxCmdExec()

	n8.IsStatusOkay()
}

// MemoryTest tests if a specific byte value exists in memory on the N8.
//
// Sends a command to test the memory region starting at the specified address,
// returns 8-bit response.
func (n8 *N8) MemoryTest(addr uint32, val uint8, length uint32) bool {
	n8.TxCmd(CMD_MEM_TST)
	n8.Tx32(addr)
	n8.Tx32(length)
	n8.Tx8(val)
	n8.TxCmdExec()

	return n8.Rx8() != 0
}

// ReadFlash reads data from flash memory on the N8.
//
// Sends a command to read data from flash memory, starting at the specified address,
// into the provided uint8 slice.
func (n8 *N8) ReadFlash(addr uint32, buf []uint8, length uint32) {
	n8.TxCmd(CMD_FLA_RD)
	n8.Tx32(addr)
	n8.Tx32(length)

	n8.RxData(buf)
}

// ReadMemory reads data from memory on the N8.
//
// Reads data from memory in chunks into the provided uint8 slice.
func (n8 *N8) ReadMemory(addr uint32, buf []uint8, length uint32) {
	if length == 0 {
		log.Fatalf("ReadMemory: No data")
		return
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

		n8.TxCmd(CMD_MEM_RD)
		n8.Tx32(addr)
		n8.Tx32(currentChunk)
		n8.TxCmdExec()

		n8.RxData(tempBuf)

		copy(buf[:currentChunk], tempBuf)
		buf = buf[currentChunk:]

		addr += currentChunk
		length -= currentChunk
	}
}

// WriteFlash writes data to flash memory on the N8.
//
// Sends a command to write data to flash memory starting at the specified address,
// verifies the status after operation.
func (n8 *N8) WriteFlash(addr uint32, buf []uint8, length uint32) {
	n8.TxCmd(CMD_FLA_WR)
	n8.Tx32(addr)
	n8.Tx32(length)

	n8.TxDataACK(buf, length)

	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[WriteFlash] status error: %v", resp)
	}
}

// WriteMemory writes data to memory on the N8.
//
// Sends a command to write data to memory starting at the specified address,
// writes the data from the provided uint8 slice.
func (n8 *N8) WriteMemory(addr uint32, buf []uint8, length uint32) {
	if length == 0 {
		log.Fatalf("WriteMemory: No data")
		return
	}

	n8.TxCmd(CMD_MEM_WR)
	n8.Tx32(addr)
	n8.Tx32(length)
	n8.TxCmdExec()

	n8.TxData(buf)
}

//
// FPGA Functions
//

// fpgaPostInit checks the N8 status after FPGA init and writes config to memory
func (n8 *N8) fpgaPostInit(config *MapConfig) {
	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[fpgaPostInit] status error: %v", resp)
	}

	if config != nil {
		n8.WriteMemory(ADDR_CFG, config.GetSerialConfig(), (uint32)(len(config.GetSerialConfig())))
	}
}

// FpgaInit initializes the N8 FPGA.
//
// Initlizes the FPGA with data from `[]uint8`,
// along with initialization data.
func (n8 *N8) FpgaInit(buf []uint8, config *MapConfig) {
	n8.TxCmd(CMD_FPGA_USB)
	n8.Tx32(uint32(len(buf)))

	n8.TxDataACK(buf, uint32(len(buf)))

	n8.fpgaPostInit(config)
}

// FpgaInitFromFlash initializes the N8 FPGA.
//
// Initlizes the FPGA with data from at address in flash,
// along with initialization data.
func (n8 *N8) FpgaInitFromFlash(address uint32, config *MapConfig) {
	n8.TxCmd(CMD_FPGA_FLA)
	n8.Tx32(address)
	n8.TxCmdExec()

	n8.fpgaPostInit(config)
}

// FpgaInitFromSD initializes the N8 FPGA using data from an SD card.
//
// Initlizes the FPGA with data on SD card, along with
// initialization data.
func (n8 *N8) FpgaInitFromSD(path string, config *MapConfig) {
	fileinfo := n8.GetFileInfo(path)

	n8.OpenFile(path, FAT_READ)
	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[FpgaInitFromSD] file open status error: %v", resp)
	}

	n8.TxCmd(CMD_FPGA_SDC)
	n8.Tx32(fileinfo.Size)
	n8.TxCmdExec()

	n8.fpgaPostInit(config)
}
