package n8

import (
	"reflect"
	"testing"
)

func TestSendString(t *testing.T) {
	tests := []struct {
		name             string
		testString       string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		wantErr          bool
	}{
		{"good write", "test", 0, nil, []int{4, 4, 4, 1, 2, 4, 4, 4, 1, 4}, false},
		{"serial write 1 fail", "test", 1, MOCK_ERR, []int{4, 4, 4, 1, 2, 4, 4, 4, 1, 4}, true},
		{"serial write 6 fail", "test", 6, MOCK_ERR, []int{4, 4, 4, 1, 2, 4, 4, 4, 1, 4}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:  test.serialFailAtCall,
				err:         test.mockError,
				serialBytes: test.serialBytes,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.SendString(test.testString)

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("SendString() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestGetConfig(t *testing.T) {

	mockConfig := &MapConfig{
		serialConfig: make([]uint8, CONFIG_BASE+16),
	}

	tests := []struct {
		name             string
		testString       string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		readData         []uint8
		wantErr          bool
		wantConfig       []uint8
	}{
		{
			"good read",
			"test",
			0,
			nil,
			[]int{4, 4, 4, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 4, 4, 4, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			mockConfig.GetSerialConfig(),
			false,
			mockConfig.GetSerialConfig(),
		},
		{
			"serial fail on read",
			"test",
			5,
			MOCK_ERR,
			[]int{4, 4, 4, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 4, 4, 4, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			mockConfig.GetSerialConfig(),
			true,
			mockConfig.GetSerialConfig(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
				testDataRead: test.readData,
				serialBytes:  test.serialBytes,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			gotConfig, err := n8.GetConfig()

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v, serial %v", err, test.wantErr, mockPort.currentCall)
				return
			} else if err == nil {
				if !reflect.DeepEqual(gotConfig.GetSerialConfig(), test.wantConfig) {
					t.Errorf("GetConfig() = %04x, want %04x", gotConfig.GetSerialConfig(), test.wantConfig)
				}
			}
		})
	}
}

func TestCopyFolder(t *testing.T) {
	tests := []struct {
		name             string
		testString       string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		readData         []uint8
		wantErr          bool
	}{
		{
			"good read",
			"test",
			0,
			nil,
			[]int{4, 1, 4, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 4, 4, 4, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			[]uint8{0},
			false,
		}, // TODO: more test cases
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
				testDataRead: test.readData,
				serialBytes:  test.serialBytes,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.CopyFolder("./", "sd:test")

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("CopyFolder() error = %v, wantErr %v, serial %v", err, test.wantErr, mockPort.currentCall)
				return
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name             string
		testString       string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		readData         []uint8
		wantErr          bool
	}{
		{
			"good read",
			"test",
			0,
			nil,
			[]int{4, 1, 2, 4, 4, 4, 1, 1024, 1, 1024, 1, 1024, 1, 1024, 1, 1024, 1, 430, 4, 1, 1, 4, 4, 1, 1},
			[]uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa5, 0x00, 0xa5},
			false,
		},
		{
			"back tx ack",
			"test",
			0,
			nil,
			[]int{4, 1, 2, 4, 4, 4, 1, 1024, 1, 1024, 1, 1024, 1, 1024, 1, 1024, 1, 430, 4, 1, 1, 4, 4, 1, 1},
			[]uint8{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa5, 0x00, 0xa5},
			true,
		},
		{
			"not all bytes sent",
			"test",
			0,
			nil,
			[]int{4, 1, 2, 4, 4, 4, 1, 0, 1, 1024, 1, 1024, 1, 1024, 1, 1024, 1, 430, 4, 1, 1, 4, 4, 1, 1},
			[]uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa5, 0x00, 0xa5},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
				testDataRead: test.readData,
				serialBytes:  test.serialBytes,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.CopyFile("../README.md", "sd:test")

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("CopyFolder() error = %v, wantErr %v, serial %v", err, test.wantErr, mockPort.currentCall)
				return
			}
		})
	}
}

func TestSetConfig(t *testing.T) {
	mockConfig := &MapConfig{
		serialConfig: make([]uint8, CONFIG_BASE+16),
	}

	tests := []struct {
		name             string
		testString       string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		readData         []uint8
		wantErr          bool
		wantConfig       []uint8
	}{
		{
			"good read",
			"test",
			0,
			nil,
			[]int{4, 4, 4, 1, 48},
			mockConfig.GetSerialConfig(),
			false,
			mockConfig.GetSerialConfig(),
		},
		{
			"serial fail on read",
			"test",
			5,
			MOCK_ERR,
			[]int{4, 4, 4, 1, 48},
			mockConfig.GetSerialConfig(),
			true,
			mockConfig.GetSerialConfig(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
				testDataRead: test.readData,
				serialBytes:  test.serialBytes,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.SetConfig(mockConfig)

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v, serial %v", err, test.wantErr, mockPort.currentCall)
				return
			}
		})
	}
}
