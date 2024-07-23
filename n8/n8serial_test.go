package n8

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// Mocking the Serial Port
type MockSerialPort struct {
	testDataIn  []byte
	testDataOut []byte
	bytesIn     []int
	bytesOut    []int
	currentCall int
	failAtCall  int
	err         error
}

func (m *MockSerialPort) Read(buf []byte) (n int, err error) {
	copy(buf, m.testDataOut)

	m.currentCall++

	fmt.Println(">>>>", m.currentCall, m.failAtCall)

	if m.currentCall == m.failAtCall {

		return m.bytesOut[m.currentCall-1], m.err
	}

	return m.bytesOut[m.currentCall-1], nil
}

func (m *MockSerialPort) Write(buf []byte) (n int, err error) {
	m.testDataIn = buf

	m.currentCall++
	if m.currentCall == m.failAtCall {
		return m.bytesIn[m.currentCall-1], m.err
	}

	return m.bytesIn[m.currentCall-1], nil
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
	testData := []byte{0x01, 0x02, 0x03}

	mockPort := &MockSerialPort{
		testDataOut: testData,
		bytesOut:    []int{len(testData)},
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
		bytesIn: []int{len(buf)},
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
	if !reflect.DeepEqual(mockPort.testDataIn, buf) {
		t.Errorf("testData and buf are not equal: %v != %v", mockPort.testDataIn, buf)
	}
}

func TestTxData(t *testing.T) {

	// Test a valid write of 3 bytes
	goodWriteBuf := []uint8{0x01, 0x02, 0x03}
	mockPort := &MockSerialPort{
		bytesIn: []int{len(goodWriteBuf)}}
	n8 := &N8{
		Port: mockPort,
	}

	err := n8.TxData(goodWriteBuf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(mockPort.testDataIn, goodWriteBuf) {
		t.Errorf("testData and buf are not equal: %v != %v", mockPort.testDataIn, goodWriteBuf)
	}

	// Test handling of error from serial port
	mockPort.currentCall = 0
	mockPort.failAtCall = 1
	mockPort.err = fmt.Errorf("mock error")

	err = n8.TxData(goodWriteBuf)
	if err == nil {
		t.Errorf("expected an error")
	}

	// Test a valid write of 3 bytes with mismatch in bytes written into the port
	mockPort.currentCall = 0
	mockPort.failAtCall = 0
	mockPort.bytesIn = []int{0}

	err = n8.TxData(goodWriteBuf)
	if err == nil {
		t.Errorf("expected an error")
	}

	// Test an invalid write of 0 bytes
	var badWriteBuf []uint8
	mockPort = &MockSerialPort{
		bytesIn: []int{len(badWriteBuf)}}
	n8 = &N8{
		Port: mockPort,
	}

	err = n8.TxData(badWriteBuf)
	if err == nil {
		t.Error("expected an error with bad data, but got none")
	}
}

func TestTx8(t *testing.T) {
	var data uint8 = 0x00

	mockPort := &MockSerialPort{
		bytesIn: []int{1}}
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
		bytesIn: []int{2}}
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
		bytesIn: []int{4}}
	n8 := &N8{
		Port: mockPort,
	}

	err := n8.Tx32(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTxCmd(t *testing.T) {
	cmd := CMD_DISK_INIT

	mockPort := &MockSerialPort{
		bytesIn: []int{4}}
	n8 := &N8{
		Port: mockPort,
	}

	// test known good command write
	err := n8.TxCmd(cmd)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test handling of error from serial port
	mockPort.currentCall = 0
	mockPort.failAtCall = 1
	mockPort.err = fmt.Errorf("mock error")
	err = n8.TxCmd(cmd)
	if err == nil {
		t.Errorf("expected an error")
	}

	// Test a valid command but with unexpected number of bytes written
	mockPort.currentCall = 0
	mockPort.failAtCall = 0
	mockPort.bytesIn = []int{2}
	mockPort.err = nil

	err = n8.TxCmd(cmd)
	if err == nil {
		t.Errorf("expected an error")
	}

}

func TestTxCmdExec(t *testing.T) {
	mockPort := &MockSerialPort{
		bytesIn: []int{1}}
	n8 := &N8{
		Port: mockPort,
	}

	err := n8.TxCmdExec()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

}

func TestTxString(t *testing.T) {
	testString := "whee"

	mockPort := &MockSerialPort{
		bytesIn: []int{2, 4}}
	n8 := &N8{
		Port: mockPort,
	}

	// test with known good value
	err := n8.TxString(testString)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

}
