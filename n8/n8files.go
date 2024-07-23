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

const (
	MKDIR_FILE_EXISTS uint16 = 0x08
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
func (n8 *N8) GetFileInfo(path string) (fileInfo *FileInfo, err error) {

	err = n8.TxCmd(CMD_FILE_INFO)
	if err != nil {
		return
	}
	err = n8.TxString(path)
	if err != nil {
		return
	}

	resp, err := n8.Rx8()
	if err != nil {
		return
	}
	if resp != 0 {
		err = fmt.Errorf("[GetFileInfo] unknown status code: %02x", resp)
		return
	}

	size, date, time, attributes, name, err := n8.RxFileInfo()
	if err != nil {
		return
	}

	fileInfo = &FileInfo{
		Size:       size,
		Date:       date,
		Time:       time,
		Attributes: attributes,
		Name:       name}

	return
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
func (n8 *N8) DirRead(maxNameLength uint16) (fileInfo *FileInfo, err error) {
	if maxNameLength == 0 {
		maxNameLength = 0xffff
	}

	err = n8.TxCmd(CMD_FILE_DIR_READ)
	if err != nil {
		return
	}
	err = n8.Tx16(maxNameLength)
	if err != nil {
		return
	}

	resp, err := n8.Rx8()
	if err != nil {
		return
	}
	if resp != 0 {
		err = fmt.Errorf("[DirRead] unknown response code: %04x", resp)
		return
	}

	size, date, time, attributes, name, err := n8.RxFileInfo()
	if err != nil {
		return
	}
	fileInfo.SetFileInfo(size, date, time, attributes, name)

	return
}

// GetDirRecords retrieves specified number of FileInfo entries from the
// currently open directory on the N8.
//
// Sends a command to retrieve a specified number of file information entries
// from the currently open directory on the N8 device.
func (n8 *N8) GetDirRecords(startIndex uint16, amount uint16, maxNameLength uint16) (fileInformation []FileInfo, err error) {
	fileInformation = make([]FileInfo, amount)

	err = n8.TxCmd(CMD_FILE_DIR_GET)
	if err != nil {
		return
	}
	err = n8.Tx16(startIndex)
	if err != nil {
		return
	}
	err = n8.Tx16(amount)
	if err != nil {
		return
	}
	err = n8.Tx16(maxNameLength)
	if err != nil {
		return
	}

	var i uint16
	for i = 0; i < amount; i++ {
		resp, err := n8.Rx8()
		if err != nil {
			return fileInformation, err
		}
		if resp != 0 {
			err = fmt.Errorf("[GetDirRecords] unknown response code: %04x", resp)
		}

		var fileInfo FileInfo
		size, date, time, attributes, name, err := n8.RxFileInfo()
		if err != nil {
			return fileInformation, err
		}

		fileInfo.SetFileInfo(size, date, time, attributes, name)

		fileInformation[i] = fileInfo
	}

	return fileInformation, nil
}

// FileCrc calculates the CRC of a file on the N8.
//
// Sends a file CRC command to the device with the specified length,
// calculates the CRC value.
func (n8 *N8) FileCrc(length uint32) (crc uint32, err error) {
	err = n8.TxCmd(CMD_FILE_CRC)
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

	resp, err := n8.Rx8()
	if err != nil {
		return
	}
	if resp != 0 {
		log.Fatalf("[FileCrc] unknown status code: %04x", resp)
	}

	return n8.Rx32()
}

//
// Non-WR File Operations
//

// OpenFile opens a file on the N8.
//
// Sends a file open command to the device with the specified path and access mode.
func (n8 *N8) OpenFile(path string, mode uint8) (err error) {
	err = n8.TxCmd(CMD_FILE_OPEN)
	if err != nil {
		return
	}
	err = n8.Tx8(mode)
	if err != nil {
		return
	}
	err = n8.TxString(path)
	if err != nil {
		return
	}

	return nil
}

// CloseFile closes the currently open file on the N8.
//
// Sends a file close command to the device to close the currently open file.
func (n8 *N8) CloseFile() (err error) {
	err = n8.TxCmd(CMD_FILE_CLOSE)
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[CloseFile] unknown response code: %v", resp)
		return
	}

	return nil
}

// FileSetPointer sets the file pointer to a specified address.
//
// Sends command to the N8 to set the file pointer to the specified address.
func (n8 *N8) FileSetPointer(address uint32) (err error) {
	err = n8.TxCmd(CMD_FILE_PTR)
	if err != nil {
		return
	}
	err = n8.Tx32(address)
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[FileSetPointer] response code error: %v", resp)
		return
	}

	return nil
}

// DirOpen opens a directory on the N8 for reading.
//
// Sends a command to open the specified directory on the N8 device.
func (n8 *N8) DirOpen(path string) (err error) {
	err = n8.TxCmd(CMD_FILE_DIR_OPEN)
	if err != nil {
		return
	}
	err = n8.TxString(path)
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[DirOpen] status error: %v", resp)
		return
	}

	return
}

// DirLoad loads a directory listing from the N8.
//
// Sends a command to load a directory listing from specified path.
func (n8 *N8) DirLoad(path string, sorted uint8) (err error) {
	err = n8.TxCmd(CMD_FILE_DIR_LD)
	if err != nil {
		return
	}
	err = n8.Tx8(sorted)
	if err != nil {
		return
	}
	err = n8.TxString(path)
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[DirLoad] status error: %v", resp)
		return
	}

	return
}

// GetDirSize retrieves number of entries in currently open directory on the N8.
//
// Sends command to retrieve number of entries in the currently open directory,
// returns number of entries as `uint16`.
func (n8 *N8) GetDirSize() (size uint16, err error) {
	err = n8.TxCmd(CMD_FILE_DIR_SIZE)
	if err != nil {
		return
	}

	return n8.Rx16()
}

//
// File Read Operations
//

// ReadFile reads data from a file on the N8.
//
// Reads data from a file in chunks of up to 4096 bytes.
func (n8 *N8) ReadFile(buf []uint8, length uint32) (err error) {
	err = n8.TxCmd(CMD_FILE_READ)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}

	const chunkSize uint32 = 0x1000
	for length > 0 {
		currentChunk := chunkSize

		if length < chunkSize {
			currentChunk = length
		}
		tempBuf := make([]uint8, currentChunk)

		resp, err := n8.Rx8()
		if err != nil {
			return err
		}
		if resp != 0 {
			err = fmt.Errorf("[ReadFile] unknown response code: %v", resp)
			return err
		}

		err = n8.RxData(tempBuf)
		if err != nil {
			return err
		}
		copy(buf[:currentChunk], tempBuf)
		buf = buf[currentChunk:]

		length -= currentChunk
	}

	return
}

// ReadFileFromMemory reads file data from memory on the N8.
//
// Reads file data from memory in chunks of up to 4096 bytes.
func (n8 *N8) ReadFileFromMemory(address uint32, length uint32) (err error) {
	const chunkSize uint32 = 0x1000
	for length > 0 {
		currentChunk := chunkSize
		if length < chunkSize {
			currentChunk = length
		}

		err = n8.TxCmd(CMD_FILE_READ_MEM)
		if err != nil {
			return
		}
		err = n8.Tx32(address)
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

		ok, resp, err := n8.IsStatusOkay()
		if err != nil {
			return err
		}
		if !ok {
			err = fmt.Errorf("[ReadFileFromMemory] unknown status code: %v", resp)
		}

		length -= currentChunk
		address += currentChunk
	}

	return
}

//
// File Write Operations
//

// FileWrite writes data to a file on the N8.
//
// Sends a file write command to the device along with the data to be written.
func (n8 *N8) FileWrite(buf []uint8, length uint32) (err error) {
	err = n8.TxCmd(CMD_FILE_WRITE)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}
	err = n8.TxDataACK(buf, length)
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[FileWrite] unknown status code: %v", resp)
		return
	}

	return
}

// FileWriteFromMemory writes data from memory to a file on the N8.
//
// Writes data from memory to a file in chunks of up to 4096 bytes.
func (n8 *N8) FileWriteFromMemory(address uint32, length uint32) (err error) {
	const chunkSize uint32 = 0x1000
	for length > 0 {
		currentChunk := chunkSize
		if length < chunkSize {
			currentChunk = length
		}

		err = n8.TxCmd(CMD_FILE_WRITE_MEM)
		if err != nil {
			return
		}
		err = n8.Tx32(address)
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

		ok, resp, err := n8.IsStatusOkay()
		if err != nil {
			return err
		}
		if !ok {
			err = fmt.Errorf("[FileWriteFromMemory] unknown status code: %v", resp)
			return err
		}

		length -= currentChunk
		address += currentChunk
	}

	return
}

// mkdir creates a new directory on the N8 device.
//
// Sends a command to create a new directory at the specified path on the N8 device.
func (n8 *N8) mkdir(path string) (err error) {
	err = n8.TxCmd(CMD_FILE_DIR_MK)
	if err != nil {
		return
	}
	err = n8.TxString(path)
	if err != nil {
		return
	}

	ok, status, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		if status == MKDIR_FILE_EXISTS {
			return
		}

		err = fmt.Errorf("[mkdir] unknown status code returned: %v", status)
		return
	}

	return
}

// DeleteRecord deletes a file or directory from the N8.
//
// Sends a file delete command to the device with the specified path.
func (n8 *N8) DeleteRecord(path string) (err error) {
	err = n8.TxCmd(CMD_FILE_DEL)
	if err != nil {
		return
	}
	err = n8.TxString(path)
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		if resp == DELETE_FILE_NOT_FOUND {
			return
		}
		log.Fatalf("[DeleteRecord] unknown resonse code: %v", resp)
	}

	return
}

//
// Disk Operations
//

// DiskInit initializes the disk on the N8.
//
// Sends a disk initialization command to the device and checks response.
func (n8 *N8) DiskInit() (err error) {
	err = n8.TxCmd(CMD_DISK_INIT)
	if err != nil {
		return
	}

	ok, resp, err := n8.IsStatusOkay()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("[DiskInit] status response error: %v", resp)
		return
	}

	return
}

// DiskRead reads data from the disk into the provided buffer.
//
// Sends a command to read data from the disk starting at the specified address,
// reads it in chunks of up to 512 bytes until the specified length is read.
func (n8 *N8) DiskRead(buf []uint8, address uint32, length uint32) (err error) {
	err = n8.TxCmd(CMD_DISK_READ)
	if err != nil {
		return
	}
	err = n8.Tx32(address)
	if err != nil {
		return
	}
	err = n8.Tx32(length)
	if err != nil {
		return
	}

	var i uint32
	for i = 0; i < length; i++ {
		resp, err := n8.Rx8()
		if err != nil {
			return err
		}
		if resp != 0 {
			log.Fatalf("[DiskRead] unknown response code: %v", resp)
		}

		tempBuf := make([]uint8, 512)
		err = n8.RxData(tempBuf)
		if err != nil {
			return err
		}

		copy(buf[:512], tempBuf)
		buf = buf[512:]
	}

	return
}

//
// Misc
//

// changeExtension changes the file extension of a given file path.
func changeExtension(path string, newExt string) (newPath string) {
	return filepath.Join(filepath.Dir(path), strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))+"."+newExt)
}
