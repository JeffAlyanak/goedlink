package n8

import (
	"encoding/binary"
	"fmt"
	"time"
)

type N8 struct {
	Address string
	Port    SerialPortInterface
}

//
// VDC
//

const VDC_DATA_SIZE = 8

type Vdc struct {
	V50 uint16
	V25 uint16
	V12 uint16
	Vbt uint16
}

// NewVdc creates a new Vdc struct from the given data.
func NewVdc(data []uint8) (v *Vdc, err error) {
	if len(data) != VDC_DATA_SIZE {
		err = fmt.Errorf("[NewVdc] invalid data length, expected %d, got %d", VDC_DATA_SIZE, len(data))
		return
	}

	v = &Vdc{
		V50: binary.LittleEndian.Uint16(data[0:2]),
		V25: binary.LittleEndian.Uint16(data[2:4]),
		V12: binary.LittleEndian.Uint16(data[4:6]),
		Vbt: binary.LittleEndian.Uint16(data[6:8]),
	}
	return
}

//
// RTC
//

const RTC_DATA_SIZE = 6

type RtcTime struct {
	Year   uint8
	Month  uint8
	Day    uint8
	Hour   uint8
	Minute uint8
	Second uint8
}

// GetVals returns the RTC time components as a uint8 slice.
//
// Returns a uint8 slice containing the binary data of the RTC time.
func (r RtcTime) GetVals() []uint8 {
	return []uint8{r.Year, r.Month, r.Day, r.Hour, r.Minute, r.Second}
}

// Print prints the RTC date and time in a formatted string.
//
// This method prints the RTC date and time in the format:
// "RTC date: YYYY-MM-DD"
// "RTC time: HH:mm:SS"
func (r RtcTime) Print() {
	fmt.Println("[RTC]")
	fmt.Printf(" Date: 20%02X-%02X-%02X\n", r.Year, r.Month, r.Day)
	fmt.Printf(" Time: %02X:%02X:%02X\n", r.Hour, r.Minute, r.Second)
}

// NewRtcTime creates a new RtcTime struct from the provided time.
func NewRtcTime(dt time.Time) *RtcTime {
	return &RtcTime{
		Year:   decToHex(dt.Year() - 2000),
		Month:  decToHex(int(dt.Month())),
		Day:    decToHex(dt.Day()),
		Hour:   decToHex(dt.Hour()),
		Minute: decToHex(dt.Minute()),
		Second: decToHex(dt.Second()),
	}
}

// NewRtcTimeFromSerial returns new RtcTime from serialized RTC data.
func NewRtcTimeFromSerial(data []uint8) (time *RtcTime, err error) {
	if len(data) != RTC_DATA_SIZE {
		err = fmt.Errorf("[NewRtcTimeFromSerial] invalid data length, expected %d, got %d", RTC_DATA_SIZE, len(data))
		return
	}

	time = &RtcTime{
		Year:   data[0],
		Month:  data[1],
		Day:    data[2],
		Hour:   data[3],
		Minute: data[4],
		Second: data[5],
	}
	return
}

//
// Misc
//

func decToHex(val int) uint8 {
	var hex int
	hex |= (val / 10) << 4
	hex |= (val % 10)
	return (uint8)(hex)
}
