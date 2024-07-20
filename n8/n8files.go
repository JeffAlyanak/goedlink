package n8

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

const CRC_INIT_VAL uint32 = 0x0000
const (
	FAT_READ          uint8 = 0x01
	FAT_WRITE         uint8 = 0x02
	FAT_OPEN_EXISTING uint8 = 0x00
	FAT_CREATE_NEW    uint8 = 0x04
	FAT_CREATE_ALWAYS uint8 = 0x08
	FAT_OPEN_ALWAYS   uint8 = 0x10
	FAT_OPEN_APPEND   uint8 = 0x30
)

const DELETE_FILE_NOT_FOUND uint16 = 0x04

type FileInfo struct {
	Size       uint32
	Date       uint16
	Time       uint16
	Attributes uint8
	Name       string
}

//
// File Information
//

// GetFileInfo retrieves information about a file from the N8.
//
// Sends a file info command to the device with the specified path,
// retrieves data and returns pointer to `FileInfo`.
func (n8 *N8) GetFileInfo(path string) *FileInfo {

	n8.TxCmd(CMD_FILE_INFO)
	n8.TxString(path)

	resp := n8.Rx8()
	if resp != 0 {
		log.Fatalf("[GetFileInfo] file access error: %02x", resp)
	}

	size, date, time, attributes, name := n8.RxFileInfo()
	return &FileInfo{
		Size:       size,
		Date:       date,
		Time:       time,
		Attributes: attributes,
		Name:       name}
}

// SetFileInfo sets the attributes of the FileInfo struct.
func (fileInfo *FileInfo) SetFileInfo(size uint32, date uint16, time uint16, attributes uint8, name string) {
	fileInfo.Size = size
	fileInfo.Date = date
	fileInfo.Time = time
	fileInfo.Attributes = attributes
	fileInfo.Name = name
}

// DirRead reads the file info from the currently open directory on the N8.
//
// Sends a command to read file info from the currently open directory.
func (n8 *N8) DirRead(maxNameLength uint16) *FileInfo {
	if maxNameLength == 0 {
		maxNameLength = 0xffff
	}

	n8.TxCmd(CMD_FILE_DIR_READ)
	n8.Tx16(maxNameLength)

	resp := n8.Rx8()
	if resp != 0 {
		log.Fatalf("[DirRead] dir read error: %v", resp)
	}

	var fileInfo FileInfo
	size, date, time, attributes, name := n8.RxFileInfo()
	fileInfo.SetFileInfo(size, date, time, attributes, name)

	return &fileInfo
}

// GetDirRecords retrieves specified number of FileInfo entries from the
// currently open directory on the N8.
//
// Sends a command to retrieve a specified number of file information entries
// from the currently open directory on the N8 device.
func (n8 *N8) GetDirRecords(startIndex uint16, amount uint16, maxNameLength uint16) []FileInfo {
	fileInformation := make([]FileInfo, amount)

	n8.TxCmd(CMD_FILE_DIR_GET)
	n8.Tx16(startIndex)
	n8.Tx16(amount)
	n8.Tx16(maxNameLength)

	var i uint16
	for i = 0; i < amount; i++ {
		resp := n8.Rx8()
		if resp != 0 {
			log.Fatalf("[GetDirRecords] dir read error: %v", resp)
		}

		var fileInfo FileInfo
		size, date, time, attributes, name := n8.RxFileInfo()
		fileInfo.SetFileInfo(size, date, time, attributes, name)

		fileInformation[i] = fileInfo
	}

	return fileInformation
}

// FileCrc calculates the CRC of a file on the N8.
//
// Sends a file CRC command to the device with the specified length,
// calculates the CRC value.
func (n8 *N8) FileCrc(length uint32) uint32 {
	n8.TxCmd(CMD_FILE_CRC)
	n8.Tx32(length)
	n8.Tx32(CRC_INIT_VAL)

	resp := n8.Rx8()

	if resp != 0 {
		log.Fatalf("[FileCrc] disk read error: %v", resp)
	}

	return n8.Rx32()
}

//
// Non-WR File Operations
//

// OpenFile opens a file on the N8.
//
// Sends a file open command to the device with the specified path and access mode.
func (n8 *N8) OpenFile(path string, mode uint8) {
	n8.TxCmd(CMD_FILE_OPEN)
	n8.Tx8(mode)
	n8.TxString(path)
}

// CloseFile closes the currently open file on the N8.
//
// Sends a file close command to the device to close the currently open file.
func (n8 *N8) CloseFile() {
	n8.TxCmd(CMD_FILE_CLOSE)

	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[CloseFile] response code error: %v", resp)
	}
}

// FileSetPointer sets the file pointer to a specified address.
//
// Sends command to the N8 to set the file pointer to the specified address.
func (n8 *N8) FileSetPointer(address uint32) {
	n8.TxCmd(CMD_FILE_PTR)
	n8.Tx32(address)

	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[FileSetPointer] response code error: %v", resp)
	}
}

// DirOpen opens a directory on the N8 for reading.
//
// Sends a command to open the specified directory on the N8 device.
func (n8 *N8) DirOpen(path string) {
	n8.TxCmd(CMD_FILE_DIR_OPEN)
	n8.TxString(path)

	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[DirOpen] status error: %v", resp)
	}
}

// DirLoad loads a directory listing from the N8.
//
// Sends a command to load a directory listing from specified path.
func (n8 *N8) DirLoad(path string, sorted uint8) {
	n8.TxCmd(CMD_FILE_DIR_LD)
	n8.Tx8(sorted)
	n8.TxString(path)

	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[DirLoad] status error: %v", resp)
	}
}

// GetDirSize retrieves number of entries in currently open directory on the N8.
//
// Sends command to retrieve number of entries in the currently open directory,
// returns number of entries as `uint16`.
func (n8 *N8) GetDirSize() uint16 {
	n8.TxCmd(CMD_FILE_DIR_SIZE)

	return n8.Rx16()
}

//
// File Read Operations
//

// ReadFile reads data from a file on the N8.
//
// Reads data from a file in chunks of up to 4096 bytes.
func (n8 *N8) ReadFile(buf []uint8, length uint32) {
	n8.TxCmd(CMD_FILE_READ)
	n8.Tx32(length)

	const chunkSize uint32 = 0x1000
	for length > 0 {
		currentChunk := chunkSize

		if length < chunkSize {
			currentChunk = length
		}
		tempBuf := make([]uint8, currentChunk)

		resp := n8.Rx8()
		if resp != 0 {
			log.Fatalf("[ReadFile] file read error: %v", resp)
		}

		n8.RxData(tempBuf)
		copy(buf[:currentChunk], tempBuf)
		buf = buf[currentChunk:]

		length -= currentChunk
	}
}

// ReadFileFromMemory reads file data from memory on the N8.
//
// Reads file data from memory in chunks of up to 4096 bytes.
func (n8 *N8) ReadFileFromMemory(address uint32, length uint32) {
	const chunkSize uint32 = 0x1000
	for length > 0 {
		currentChunk := chunkSize
		if length < chunkSize {
			currentChunk = length
		}

		n8.TxCmd(CMD_FILE_READ_MEM)
		n8.Tx32(address)
		n8.Tx32(currentChunk)
		n8.TxCmdExec()

		ok, resp := n8.IsStatusOkay()
		if !ok {
			log.Fatalf("[ReadFileFromMemory] file read error: %v", resp)
		}

		length -= currentChunk
		address += currentChunk
	}
}

//
// File Write Operations
//

// FileWrite writes data to a file on the N8.
//
// Sends a file write command to the device along with the data to be written.
func (n8 *N8) FileWrite(buf []uint8, length uint32) {
	n8.TxCmd(CMD_FILE_WRITE)
	n8.Tx32(length)
	n8.TxDataACK(buf, length)

	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[FileWrite] file write error: %v", resp)
	}
}

// FileWriteFromMemory writes data from memory to a file on the N8.
//
// Writes data from memory to a file in chunks of up to 4096 bytes.
func (n8 *N8) FileWriteFromMemory(address uint32, length uint32) {
	const chunkSize uint32 = 0x1000
	for length > 0 {
		currentChunk := chunkSize
		if length < chunkSize {
			currentChunk = length
		}

		n8.TxCmd(CMD_FILE_WRITE_MEM)
		n8.Tx32(address)
		n8.Tx32(currentChunk)
		n8.TxCmdExec()

		ok, resp := n8.IsStatusOkay()
		if !ok {
			log.Fatalf("[FileWriteFromMemory] file write error: %v", resp)
		}

		length -= currentChunk
		address += currentChunk
	}
}

// mkdir creates a new directory on the N8 device.
//
// Sends a command to create a new directory at the specified path on the N8 device.
func (n8 *N8) mkdir(path string) {
	n8.TxCmd(CMD_FILE_DIR_MK)
	n8.TxString(path)

	ok, status := n8.IsStatusOkay()
	if !ok {
		if status == 8 {
			return // directory already exists, no action needed
		}
		log.Fatalf("[mkdir] status error: %v", status)
	}
}

// DeleteRecord deletes a file or directory from the N8.
//
// Sends a file delete command to the device with the specified path.
func (n8 *N8) DeleteRecord(path string) {
	n8.TxCmd(CMD_FILE_DEL)
	n8.TxString(path)

	ok, resp := n8.IsStatusOkay()
	if !ok {
		if resp == DELETE_FILE_NOT_FOUND {
			fmt.Printf("[DeleteRecord] could not delete %s, file doesn't exist.\n", path)
			return
		}
		log.Fatalf("[DeleteRecord] status response error: %v", resp)
	}
}

//
// Disk Operations
//

// DiskInit initializes the disk on the N8.
//
// Sends a disk initialization command to the device and checks response.
func (n8 *N8) DiskInit() {
	n8.TxCmd(CMD_DISK_INIT)

	ok, resp := n8.IsStatusOkay()
	if !ok {
		log.Fatalf("[DiskInit] status response error: %v", resp)
	}
}

// DiskRead reads data from the disk into the provided buffer.
//
// Sends a command to read data from the disk starting at the specified address,
// reads it in chunks of up to 512 bytes until the specified length is read.
func (n8 *N8) DiskRead(buf []uint8, address uint32, length uint32) {
	n8.TxCmd(CMD_DISK_READ)
	n8.Tx32(address)
	n8.Tx32(length)

	var i uint32
	for i = 0; i < length; i++ {
		resp := n8.Rx8()
		if resp != 0 {
			log.Fatalf("[DiskRead] disk read error: %v", resp)
		}

		tempBuf := make([]uint8, 512)
		n8.RxData(tempBuf)

		copy(buf[:512], tempBuf)
		buf = buf[512:]
	}
}

//
// Misc
//

// changeExtension changes the file extension of a given file path.
func changeExtension(path string, newExt string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	newPath := filepath.Join(filepath.Dir(path), base+"."+newExt)

	return newPath
}
