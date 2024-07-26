package n8

import (
	"fmt"
	"testing"
	"time"
)

// Mocking the Serial Port
type MockSerialPort struct {
	testDataWrite []byte
	testDataRead  []byte
	serialBytes   []int
	currentCall   int
	failAtCall    int
	err           error
}

var MOCK_ERR error = fmt.Errorf("mock error")

func (m *MockSerialPort) Read(buf []byte) (n int, err error) {
	copy(buf, m.testDataRead[:len(buf)])
	m.testDataRead = m.testDataRead[len(buf):]

	m.currentCall++

	if m.currentCall == m.failAtCall {

		return m.serialBytes[m.currentCall-1], m.err
	}

	return m.serialBytes[m.currentCall-1], nil
}

func (m *MockSerialPort) Write(buf []byte) (n int, err error) {
	m.testDataWrite = buf

	m.currentCall++
	if m.currentCall == m.failAtCall {
		return m.serialBytes[m.currentCall-1], m.err
	}

	return m.serialBytes[m.currentCall-1], nil
}

func (m *MockSerialPort) Close() error {
	return nil
}

func TestInitSerial(t *testing.T) {

	// init serial and confirm device name is set
	testDeviceName := "INVALID_SERIAL_DEVICE"
	n8 := &N8{}
	err := n8.InitSerial(testDeviceName, time.Second)
	if n8.Address != testDeviceName {
		t.Errorf("InitSerial() error: expected address to be '%s', got '%s'", testDeviceName, n8.Address)
	}
	if err == nil {
		t.Errorf("Close() error = %v, wantErr True", err)
	}

}

func TestCloseSerial(t *testing.T) {
	// Set up a mock serial port
	mockPort := &MockSerialPort{}

	// Create an instance of N8 with the mock port
	n8 := &N8{
		Address: "",
		Port:    mockPort,
	}

	t.Run("test close", func(t *testing.T) {
		err := n8.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	t.Run("test closeserial", func(t *testing.T) {
		err := n8.CloseSerial()
		if err != nil {
			t.Errorf("CloseSerial() error = %v", err)

		}
	})
}

func TestBasicSerialRead(t *testing.T) {
	tests := []struct {
		name             string
		buf              []uint8
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		dataRead         []byte
		wantErr          bool
		wantReturn       int
	}{
		{"good input", []uint8{1, 2, 3}, 0, nil, []int{1, 3}, []byte{0}, false, 3},
		{"serial read error", []uint8{1, 2, 3}, 1, MOCK_ERR, []int{1, 3}, []byte{0}, true, 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				testDataRead: test.buf,
				serialBytes:  []int{len(test.buf)},
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			got, err := n8.Read(test.buf)

			// test errors and return
			if (err != nil) != test.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, test.wantErr)
				return
			} else if err == nil {
				if got != test.wantReturn {
					t.Errorf("Read() = %04x, want %04x", got, test.wantReturn)
				}
			}
		})
	}
}

func TestBasicSerialWrite(t *testing.T) {
	tests := []struct {
		name             string
		buf              []uint8
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		dataRead         []byte
		wantErr          bool
		wantReturn       int
	}{
		{"good input", []uint8{1, 2, 3}, 0, nil, []int{1, 3}, []byte{0}, false, 3},
		{"serial read error", []uint8{1, 2, 3}, 1, MOCK_ERR, []int{1, 3}, []byte{0}, true, 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				testDataRead: test.buf,
				serialBytes:  []int{len(test.buf)},
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			got, err := n8.Write(test.buf)

			// test errors and return
			if (err != nil) != test.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, test.wantErr)
				return
			} else if err == nil {
				if got != test.wantReturn {
					t.Errorf("Write() = %04x, want %04x", got, test.wantReturn)
				}
			}
		})
	}
}

func TestTxData(t *testing.T) {
	tests := []struct {
		name             string
		testBuf          []uint8
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		wantErr          bool
	}{
		{"good input", []uint8{0x01, 0x02, 0x03}, 0, nil, []int{3}, false},
		{"no input", nil, 0, nil, []int{3}, true},
		{"not all bytes written", []uint8{0x01, 0x02, 0x03}, 0, nil, []int{2}, true},
		{"serial error", []uint8{0x01, 0x02, 0x03}, 1, MOCK_ERR, []int{3}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:  test.serialFailAtCall,
				err:         test.mockError,
				serialBytes: test.serialBytes}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.TxData(test.testBuf)

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("TxData() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestTx8(t *testing.T) {
	var data uint8 = 0x00

	mockPort := &MockSerialPort{
		serialBytes: []int{1}}
	n8 := &N8{
		Port: mockPort,
	}

	err := n8.Tx8(data)
	if err != nil {
		t.Errorf("Tx8() error = %v", err)
	}
}

func TestTx16(t *testing.T) {
	var data uint16 = 0x0000

	mockPort := &MockSerialPort{
		serialBytes: []int{2}}
	n8 := &N8{
		Port: mockPort,
	}

	err := n8.Tx16(data)
	if err != nil {
		t.Errorf("Tx16() error = %v", err)
	}
}

func TestTx32(t *testing.T) {
	var data uint32 = 0x00000000

	mockPort := &MockSerialPort{
		serialBytes: []int{4}}
	n8 := &N8{
		Port: mockPort,
	}

	err := n8.Tx32(data)
	if err != nil {
		t.Errorf("Tx32() error = %v", err)
	}
}

func TestTxCmd(t *testing.T) {
	tests := []struct {
		name             string
		testCmd          uint8
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		wantErr          bool
	}{
		{"good input", CMD_DISK_INIT, 0, nil, []int{4}, false},
		{"not all bytes written", CMD_DISK_INIT, 0, nil, []int{2}, true},
		{"serial error", CMD_DISK_INIT, 1, MOCK_ERR, []int{4}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:  test.serialFailAtCall,
				err:         test.mockError,
				serialBytes: test.serialBytes}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.TxCmd(test.testCmd)

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("TxCmd() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestTxCmdExec(t *testing.T) {
	mockPort := &MockSerialPort{
		serialBytes: []int{1}}
	n8 := &N8{
		Port: mockPort,
	}

	err := n8.TxCmdExec()
	if err != nil {
		t.Errorf("TxCmdExec() error = %v", err)
	}

}

func TestTxString(t *testing.T) {
	tests := []struct {
		name             string
		testString       string
		serialFailAtCall int
		serialBytes      []int
		wantErr          bool
	}{
		{"good input", "validstring", 0, []int{2, 11}, false},
		{"bad length on first call", "validstring", 0, []int{1, 11}, true},
		{"bad length on second call", "validstring", 0, []int{2, 2}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:  test.serialFailAtCall,
				serialBytes: test.serialBytes}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.TxString(test.testString)

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("TxString() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestTxStringFifo(t *testing.T) {
	tests := []struct {
		name             string
		testString       string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		wantErr          bool
	}{
		{"good input", "valid", 0, nil, []int{4, 4, 4, 1, 2, 4, 4, 4, 1, 5}, false},
		{"not all bytes written when sending length", "valid", 0, nil, []int{4, 4, 4, 1, 0, 4, 4, 4, 1, 5}, true},
		{"not all bytes written when sending string", "valid", 0, nil, []int{4, 4, 4, 1, 2, 4, 4, 4, 1, 0}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:  test.serialFailAtCall,
				err:         test.mockError,
				serialBytes: test.serialBytes}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.TxStringFifo(test.testString)

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("TxStringFifo() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestTxDataAck(t *testing.T) {
	tests := []struct {
		name             string
		buf              []uint8
		length           uint32
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		dataRead         []byte
		wantErr          bool
	}{
		{"good input", []uint8{3, 2, 3}, 3, 0, nil, []int{1, 3}, []byte{0}, false},
		{"bad ack", []uint8{3, 2, 3}, 3, 0, nil, []int{1, 3}, []byte{0xff}, true},
		{"serial fail write", []uint8{3, 2, 3}, 3, 2, MOCK_ERR, []int{1, 3}, []byte{0xff}, true},
		{"not all bytes read", []uint8{3, 2, 3}, 2, 0, nil, []int{0, 3}, []byte{0xff}, true},
		{"not all bytes writter", []uint8{3, 2, 3}, 2, 0, nil, []int{1, 0}, []byte{0xff}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
				serialBytes:  test.serialBytes,
				testDataRead: test.dataRead,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			err := n8.TxDataAck(test.buf, test.length)

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("TxDataAck() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestRxData(t *testing.T) {
	tests := []struct {
		name             string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		dataRead         []byte
		wantErr          bool
	}{
		{"good read", 0, nil, []int{1, 1, 1}, []byte{0, 0, 0}, false},
		{"serial error", 1, MOCK_ERR, []int{1, 1, 1}, []byte{0, 0, 0}, true},
		{"byte fails to read", 0, nil, []int{1, 1, 0}, []byte{0, 0, 0}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
				serialBytes:  test.serialBytes,
				testDataRead: test.dataRead,
			}
			n8 := &N8{
				Port: mockPort,
			}
			buf := make([]uint8, len(test.dataRead))

			// function under test
			err := n8.RxData(buf)

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("RxData() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestRx16(t *testing.T) {

	mockPort := &MockSerialPort{
		serialBytes:  []int{1, 1},
		testDataRead: []uint8{0x01, 0x02},
	}
	n8 := &N8{
		Port: mockPort,
	}

	got, err := n8.Rx16()
	if err != nil {
		t.Errorf("Rx16() error = %v, wantErr True", err)
	}
	if got != 0x0201 {
		t.Errorf("Rx16() = %04x, want %04x", got, 0x0201)
	}
}

func TestRx32(t *testing.T) {

	mockPort := &MockSerialPort{
		serialBytes:  []int{1, 1, 1, 1},
		testDataRead: []uint8{0x01, 0x02, 0x03, 0x04},
	}
	n8 := &N8{
		Port: mockPort,
	}

	got, err := n8.Rx32()
	if err != nil {
		t.Errorf("Rx32() error = %v, wantErr True", err)
	}
	if got != 0x04030201 {
		t.Errorf("Rx32() = %04x, want %04x", got, 0x04030201)
	}
}

func TestRxString(t *testing.T) {
	tests := []struct {
		name             string
		testDataRead     []uint8
		serialFailAtCall int
		testError        error
		serialBytes      []int
		wantErr          bool
		wantReturn       string
	}{
		{"good input", []uint8{4, 0, (uint8)('t'), (uint8)('e'), (uint8)('s'), (uint8)('t')}, 0, nil, []int{1, 1, 1, 1, 1, 1}, false, "test"},
		{"length is 0", []uint8{0, 0}, 0, nil, []int{1, 1}, true, "false"},
		{"serial fail on first read", []uint8{4, 0, (uint8)('t'), (uint8)('e'), (uint8)('s'), (uint8)('t')}, 1, MOCK_ERR, []int{1, 1, 1, 1, 1, 1}, true, "test"},
		{"serial fail on second read", []uint8{4, 0, (uint8)('t'), (uint8)('e'), (uint8)('s'), (uint8)('t')}, 2, MOCK_ERR, []int{1, 1, 1, 1, 1, 1}, true, "test"},
		{"serial fail on third read", []uint8{4, 0, (uint8)('t'), (uint8)('e'), (uint8)('s'), (uint8)('t')}, 3, MOCK_ERR, []int{1, 1, 1, 1, 1, 1}, true, "test"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				serialBytes:  test.serialBytes,
				testDataRead: test.testDataRead,
				err:          test.testError,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			got, err := n8.RxString()

			// test errors and return
			if (err != nil) != test.wantErr {
				t.Errorf("RxString() error = %v, wantErr %v", err, test.wantErr)
				return
			} else if err == nil {
				if got != test.wantReturn {
					t.Errorf("RxString() = %s, want %s", got, test.wantReturn)
				}
			}
		})
	}
}

func TestRxFileInfo(t *testing.T) {
	tests := []struct {
		name                 string
		testDataRead         []uint8
		serialFailAtCall     int
		testError            error
		serialBytes          []int
		wantErr              bool
		wantReturnSize       uint32
		wantReturnDate       uint16
		wantReturnTime       uint16
		wantReturnAttributes uint8
		wantReturnName       string
	}{
		{"good input", []uint8{
			0x01,
			0x02,
			0x03,
			0x04,
			0x05,
			0x06,
			0x07,
			0x08,
			0x09,
			0x04,
			0x00,
			uint8('t'),
			uint8('e'),
			uint8('s'),
			uint8('t'),
		},
			0,
			nil,
			[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			false,
			0x04030201,
			0x0605,
			0x0807,
			0x09,
			"test"},
		{"serial fail on first read", []uint8{
			0x01,
			0x02,
			0x03,
			0x04,
			0x05,
			0x06,
			0x07,
			0x08,
			0x09,
			0x04,
			0x00,
			uint8('t'),
			uint8('e'),
			uint8('s'),
			uint8('t'),
		},
			1,
			MOCK_ERR,
			[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			true,
			0x04030201,
			0x0605,
			0x0807,
			0x09,
			"test"},
		{"serial fail on fifth read", []uint8{
			0x01,
			0x02,
			0x03,
			0x04,
			0x05,
			0x06,
			0x07,
			0x08,
			0x09,
			0x04,
			0x00,
			uint8('t'),
			uint8('e'),
			uint8('s'),
			uint8('t'),
		},
			5,
			MOCK_ERR,
			[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			true,
			0x04030201,
			0x0605,
			0x0807,
			0x09,
			"test"},
		{"serial fail on seventh read", []uint8{
			0x01,
			0x02,
			0x03,
			0x04,
			0x05,
			0x06,
			0x07,
			0x08,
			0x09,
			0x04,
			0x00,
			uint8('t'),
			uint8('e'),
			uint8('s'),
			uint8('t'),
		},
			7,
			MOCK_ERR,
			[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			true,
			0x04030201,
			0x0605,
			0x0807,
			0x09,
			"test"},
		{"serial fail on ninth read", []uint8{
			0x01,
			0x02,
			0x03,
			0x04,
			0x05,
			0x06,
			0x07,
			0x08,
			0x09,
			0x04,
			0x00,
			uint8('t'),
			uint8('e'),
			uint8('s'),
			uint8('t'),
		},
			9,
			MOCK_ERR,
			[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			true,
			0x04030201,
			0x0605,
			0x0807,
			0x09,
			"test"},
		{"serial fail on tenth read", []uint8{
			0x01,
			0x02,
			0x03,
			0x04,
			0x05,
			0x06,
			0x07,
			0x08,
			0x09,
			0x04,
			0x00,
			uint8('t'),
			uint8('e'),
			uint8('s'),
			uint8('t'),
		},
			10,
			MOCK_ERR,
			[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			true,
			0x04030201,
			0x0605,
			0x0807,
			0x09,
			"test"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				serialBytes:  test.serialBytes,
				testDataRead: test.testDataRead,
				err:          test.testError,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			size, date, time, attributes, name, err := n8.RxFileInfo()

			// test errors and return
			if (err != nil) != test.wantErr {
				t.Errorf("RxFileInfo() error = %v, wantErr %v", err, test.wantErr)
				return
			} else if err == nil {
				if size != test.wantReturnSize ||
					date != test.wantReturnDate ||
					time != test.wantReturnTime ||
					attributes != test.wantReturnAttributes ||
					name != test.wantReturnName {
					t.Errorf("RxFileInfo() = (%x, %x, %x, %x, %s), want (%x, %x, %x, %x, %s)",
						size,
						date,
						time,
						attributes,
						name,
						test.wantReturnSize,
						test.wantReturnDate,
						test.wantReturnTime,
						test.wantReturnAttributes,
						test.wantReturnName,
					)
				}
			}
		})
	}
}

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name             string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		dataRead         []byte
		wantErr          bool
		wantReturnOk     bool
		wantReturnStatus uint16
	}{
		{"good read", 0, nil, []int{4, 1, 1}, []byte{0xff, 0xa5}, false, true, 0x00ff},
		{"fail first serial write", 1, MOCK_ERR, []int{4, 1, 1}, []byte{0xff, 0xa5}, true, true, 0x00ff},
		{"fail first serial read", 2, MOCK_ERR, []int{4, 1, 1}, []byte{0xff, 0xa5}, true, true, 0x00ff},
		{"invalid status code", 0, nil, []int{4, 1, 1}, []byte{0xff, 0xff}, true, true, 0x00ff},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
				serialBytes:  test.serialBytes,
				testDataRead: test.dataRead,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			ok, status, err := n8.GetStatus()

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, test.wantErr)
				return
			} else if err == nil {
				if !ok || status != test.wantReturnStatus {
					t.Errorf("GetStatus() = (%v, %04x), want (%v, %04x)", ok, status, test.wantReturnOk, test.wantReturnStatus)
				}

			}

		})
	}
}

func TestIsStatusOkay(t *testing.T) {
	tests := []struct {
		name             string
		serialFailAtCall int
		mockError        error
		serialBytes      []int
		dataRead         []byte
		wantErr          bool
		wantReturnOk     bool
		wantReturnStatus uint16
	}{
		{"good read", 0, nil, []int{4, 1, 1}, []byte{0x00, 0xa5}, false, true, 0x0000},
		{"serial read fail", 1, MOCK_ERR, []int{4, 1, 1}, []byte{0x00, 0xa5}, true, true, 0x0000},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// setup mock serial port
			mockPort := &MockSerialPort{
				failAtCall:   test.serialFailAtCall,
				err:          test.mockError,
				serialBytes:  test.serialBytes,
				testDataRead: test.dataRead,
			}
			n8 := &N8{
				Port: mockPort,
			}

			// function under test
			ok, status, err := n8.IsStatusOkay()

			// test errors
			if (err != nil) != test.wantErr {
				t.Errorf("IsStatusOkay() error = %v, wantErr %v", err, test.wantErr)
				return
			} else if err == nil {
				if ok != test.wantReturnOk {
					t.Errorf("IsStatusOkay() = (%v, %04x), want (%v, %04x)", ok, status, test.wantReturnOk, test.wantReturnStatus)
				}

			}

		})
	}
}
