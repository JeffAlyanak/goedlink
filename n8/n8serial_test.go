package n8

import (
	"fmt"
	"reflect"
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
	fmt.Println(">>", m.serialBytes[m.currentCall-1])
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
		t.Errorf("Expected address to be '%s', got '%s'", testDeviceName, n8.Address)
	}
	if err == nil {
		t.Errorf("error expected")
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

	err := n8.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = n8.CloseSerial()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBasicSerialRead(t *testing.T) {
	// test setup
	var buf []uint8 = make([]uint8, 3)
	testData := []byte{0x01, 0x02, 0x03, 0x01, 0x02, 0x03}

	mockPort := &MockSerialPort{
		testDataRead: testData,
		serialBytes:  []int{len(testData)},
	}
	n8 := &N8{
		Port: mockPort,
	}

	//
	n, err := n8.Read(buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != len(buf) {
		t.Errorf("bytes read/write doesn't match length of data: %v != %v", n, len(buf))
	}

	//
	mockPort.currentCall = 0
	n, err = n8.Port.Read(buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != len(buf) {
		t.Errorf("bytes read/write doesn't match length of data: %v != %v", n, len(buf))
	}
	if !reflect.DeepEqual(testData, buf) {
		t.Errorf("testData and read buf are not equal: %v != %v", testData, buf)
	}
}

func TestBasicSerialWrite(t *testing.T) {
	buf := []uint8{0x01, 0x02, 0x03}

	mockPort := &MockSerialPort{
		serialBytes: []int{len(buf)},
	}
	n8 := &N8{
		Port: mockPort,
	}

	n, err := n8.Port.Write(buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != len(buf) {
		t.Errorf("bytes read/write doesn't match length of data: %v != %v", n, len(buf))
	}
	if !reflect.DeepEqual(mockPort.testDataWrite, buf) {
		t.Errorf("testData and buf are not equal: %v != %v", mockPort.testDataWrite, buf)
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
				t.Errorf("TxString() error = %v, wantErr %v", err, test.wantErr)
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
		t.Errorf("unexpected error: %v", err)
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
		t.Errorf("unexpected error: %v", err)
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
		t.Errorf("unexpected error: %v", err)
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
				t.Errorf("TxString() error = %v, wantErr %v", err, test.wantErr)
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
		t.Errorf("unexpected error: %v", err)
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
				t.Errorf("TxString() error = %v, wantErr %v", err, test.wantErr)
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
				t.Errorf("TxString() error = %v, wantErr %v", err, test.wantErr)
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
				t.Errorf("TxString() error = %v, wantErr %v", err, test.wantErr)
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
		t.Errorf("unexpected error: %v", err)
	}
	if got != 0x0201 {
		t.Errorf("TestRx16() = %04x, want %04x", got, 0x0201)
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
		t.Errorf("unexpected error: %v", err)
	}
	if got != 0x04030201 {
		t.Errorf("TestRx16() = %04x, want %04x", got, 0x04030201)
	}
}
