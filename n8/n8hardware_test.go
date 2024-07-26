package n8

import (
	"reflect"
	"testing"
	"time"
)

func TestNewVdc(t *testing.T) {
	tests := []struct {
		name       string
		buf        []uint8
		wantErr    bool
		wantReturn *Vdc
	}{
		{
			"good input",
			[]uint8{1, 2, 3, 4, 5, 6, 7, 8},
			false,
			&Vdc{0x0201, 0x0403, 0x0605, 0x0807},
		},
		{
			"invalid data length",
			[]uint8{1, 2},
			true,
			&Vdc{0x0201, 0x0403, 0x0605, 0x0807},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// function under test
			got, err := NewVdc(test.buf)

			// test errors and return
			if (err != nil) != test.wantErr {
				t.Errorf("NewVdc() error = %v, wantErr %v", err, test.wantErr)
				return
			} else if err == nil {

				if !reflect.DeepEqual(got, test.wantReturn) {
					t.Errorf("NewVdc() = %04x, want %04x", got, test.wantReturn)
				}
			}
		})
	}
}

func TestRtcTimeGetVals(t *testing.T) {
	tests := []struct {
		name       string
		rtcTime    *RtcTime
		wantErr    bool
		wantReturn []uint8
	}{
		{
			"good input",
			&RtcTime{1, 2, 3, 4, 5, 6},
			false,
			[]uint8{1, 2, 3, 4, 5, 6},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// function under test
			got := test.rtcTime.GetVals()

			// test return
			if !reflect.DeepEqual(got, test.wantReturn) {
				t.Errorf("RtcTime.GetVals() = %04x, want %04x", got, test.wantReturn)
			}
		})
	}
}

func TestNewRtcTime(t *testing.T) {

	tests := []struct {
		name       string
		time       string
		wantErr    bool
		wantReturn *RtcTime
	}{
		{
			"good input",
			"2001-02-03 04:05:06",
			false,
			&RtcTime{1, 2, 3, 4, 5, 6},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			time, _ := time.Parse("2006-01-02 15:04:05", test.time)

			// function under test
			got := NewRtcTime(time)

			// test return
			if !reflect.DeepEqual(got, test.wantReturn) {
				t.Errorf("NewRtcTime() = %04x, want %04x", got, test.wantReturn)
			}
		})
	}
}

func TestNewRtcTimeFromSerial(t *testing.T) {
	tests := []struct {
		name       string
		buf        []uint8
		wantErr    bool
		wantReturn *RtcTime
	}{
		{
			"good input",
			[]uint8{1, 2, 3, 4, 5, 6},
			false,
			&RtcTime{1, 2, 3, 4, 5, 6},
		},
		{
			"invalid data length",
			[]uint8{1, 2},
			true,
			&RtcTime{1, 2, 3, 4, 5, 6},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// function under test
			got, err := NewRtcTimeFromSerial(test.buf)

			// test errors and return
			if (err != nil) != test.wantErr {
				t.Errorf("NewRtcTimeFromSerial() error = %v, wantErr %v", err, test.wantErr)
				return
			} else if err == nil {

				if !reflect.DeepEqual(got, test.wantReturn) {
					t.Errorf("NewRtcTimeFromSerial() = %04x, want %04x", got, test.wantReturn)
				}
			}
		})
	}
}
