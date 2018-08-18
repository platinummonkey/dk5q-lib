package dk5q_lib

import (
	"github.com/platinummonkey/hid"
	"sync"
	"fmt"
	"runtime"
	"time"
)

// DefaultDasKeyboard is a DasKeyboard Gen 1 HID USB Keyboard implementation.
type DefaultDasKeyboard struct {
	vendorID uint16
	productID uint16
	deviceInterface int
	usage uint16

	device *hid.Device
	deviceInfo *hid.DeviceInfo

	mu sync.Mutex

	sequence byte
}

// NewDefaultDasKeyboard will make a new HID USB Keyboard
func NewDefaultDasKeyboard(vendorID uint16, productID uint16, deviceInterface int, usage uint16) *DefaultDasKeyboard {
	return &DefaultDasKeyboard{
		vendorID: vendorID,
		productID: productID,
		deviceInterface: deviceInterface,
		usage: usage,
	}
}

// Connect to the keyboard
func (d *DefaultDasKeyboard) Connect() (err error) {
	devices := hid.Enumerate(d.vendorID, d.productID)
	if len(devices) == 0 {
		err = fmt.Errorf("no such device found")
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	for _, device := range devices {
		fmt.Printf("Found device: Mfg: %v (%v)\n\tProduct: %v (%v)\n\tSerial: %v\n\tPath: %v\n\tInterface: %v\n\tUsage: %v\n\tUsage Page: %v\n", device.Manufacturer, device.VendorID, device.Product, device.ProductID, device.Serial, device.Path, device.Interface, device.Usage, device.UsagePage)
		if device.Path != "" {
			d.deviceInfo = &device
			if runtime.GOOS == "darwin" {
				if device.Usage == d.usage {
					fmt.Printf("Found matching device, trying to open connection...\n")
					d.device, err = d.deviceInfo.Open()
					if err == nil {
						fmt.Printf("successfully opened device: %v\n", d.device)
						break
					}
				}
			} else { // assume linux/unix/windows
				if d.deviceInterface == device.Interface {
					fmt.Printf("Found matching device, trying to open connection...\n")
					d.device, err = d.deviceInfo.Open()
					if err == nil {
						fmt.Printf("successfully opened device: %v\n", d.device)
						break
					}
				}
			}
		}
	}

	if d.deviceInfo == nil && (d.deviceInterface == 0 && d.usage == 0) {
		err = fmt.Errorf("no matching compatible device found")
		return
	}

	return
}

func (d *DefaultDasKeyboard) Initialize() (err error) {
	fmt.Printf("Initializing keyboard...\n")
	_, err = d.FeatureReport(0x00, InitializePacket(d.deviceInfo))
	return
}

func (d *DefaultDasKeyboard) FreezeEffects() (err error) {
	fmt.Printf("freezing keyboard effects...\n")
	_, err = d.FeatureReport(0x00, FreezePacket(d.deviceInfo))
	return
}

func (d *DefaultDasKeyboard) SetKeyState(state KeyState) (err error) {
	for _, packet := range state.BuildStatePackets(d.deviceInfo) {
		_, err = d.FeatureReport(0x00, packet)
		if err != nil {
			return
		}
	}
	return
}

func (d *DefaultDasKeyboard) Apply() (err error) {
	fmt.Printf("Applying any pending color commands...\n")
	_, err = d.FeatureReport(0x00, TriggerPacket(d.deviceInfo))
	return
}

func (d *DefaultDasKeyboard) SetBrightness(brightness uint8) (err error) {
	fmt.Printf("Applying any pending color commands...\n")
	packet, err := BrightnessPacket(d.deviceInfo, brightness)
	if err != nil {
		return
	}
	_, err = d.FeatureReport(0x00, packet)
	return
}


func (d *DefaultDasKeyboard) GetKeyboardData() (err error, info DasKeyboardFirmwareInfo) {
	fmt.Printf("Getting keyboard data...\n")
	_, err = d.FeatureReport(0x00, FirmwarePacket(d.deviceInfo))
	if err != nil {
		return
	}
	firmwareVersionPacket := make([]byte, 65)
	bytesRead, err := d.device.GetFeatureReport(firmwareVersionPacket)
	if err != nil {
		return
	}
	if bytesRead < 8 {
		err = fmt.Errorf("expected at least 8 bytes")
		return
	}
	fmt.Printf("Got firmware packet: %v\n", firmwareVersionPacket)
	info.MajorVersion = firmwareVersionPacket[4]
	info.MinorVersion = firmwareVersionPacket[5]
	info.PatchVersion = firmwareVersionPacket[6]
	info.RCVersion = firmwareVersionPacket[7]
	info.PacketCount = firmwareVersionPacket[3]
	fmt.Printf("\t Got version %v\n", info)
	return
}

const maxBufSize = 65

func (d *DefaultDasKeyboard) getAndIncrementSequence() byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	sequence := byte(d.sequence)
	if d.sequence == 0xFF {
		d.sequence = 0
	} else {
		d.sequence += 1
	}
	return sequence
}

func (d *DefaultDasKeyboard) FeatureReport(reportID byte, report []byte) (result []byte, err error) {
	buf := make([]byte, 65)
	buf[0] = reportID
	for i, b := range report {
		buf[i+1] = b
	}
	sequence := d.getAndIncrementSequence()
	buf[3] = sequence
	bytesWritten, err := d.device.SendFeatureReport(buf)
	if err != nil {
		fmt.Printf("1. could not write feature report: (%d) %v\n\terr: %v", len(buf), buf, err)
		return
	}
	fmt.Printf("1. wrote %d bytes in feature report: %v\n", bytesWritten, buf)

	buf[2] = 0
	buf[3] = sequence
	bytesRead, err := d.device.GetFeatureReport(buf)
	if err != nil {
		fmt.Printf("2. could not read feature report: (%d) %v\n\terr: %v\n", len(buf), buf, err)
		return
	}
	if bytesRead < 3 {
		err = fmt.Errorf("expected at least 3 bytes, but got %d: %v", bytesRead, buf)
		return
	}

	fmt.Printf("2. read %d bytes in feature report: %v\n", bytesRead, buf)

	if buf[1] != 0x14 {
		err = fmt.Errorf("invalid ack response received: expected 20 but got: %v", buf[1])
		return
	}

	if buf[2] != sequence {
		err = fmt.Errorf("wrong sequence number in response: got: %d, expected %d", buf[2], sequence)
		return
	}

	result = buf

	return
}

func (d *DefaultDasKeyboard) Read() (data []byte, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.device == nil {
		err = fmt.Errorf("not connected")
		return
	}
	data = make([]byte, 65)
	data[0] = 0

	retry := 0
	bytesRead := 0
	for bytesRead != 65 {
		fmt.Printf("trying to read, retry=%d, bytesRead=%d, data=%v\n", retry, bytesRead, data)
		if retry >= 1 {
			if retry > 5 {
				err = fmt.Errorf("maximum retries exceeded")
				return
			}
			time.Sleep(100*time.Millisecond)
		}


		bytesRead, err = d.device.Read(data)
		if bytesRead > 0 && runtime.GOOS == "darwin" {
			bytesRead -= 1
			data = data[1:]
		}
		retry += 1
	}
	return
}

func (d *DefaultDasKeyboard) Write(data []byte) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.device == nil {
		err = fmt.Errorf("not connected")
		return
	}

	bytesWritten := 0
	retry := 0
	for bytesWritten != len(data) {
		if retry >= 1 {
			if retry > 5 {
				err = fmt.Errorf("maximum retries exceeded")
				return
			}
			time.Sleep(100*time.Millisecond)
		}

		bytesWritten, err = d.device.Write(data)
		fmt.Printf("\tgot error: %v\n", err)

		if err == nil {
			return
		}
		retry += 1
	}
	return
}

// Disconnect from the keyboard
func (d *DefaultDasKeyboard) Disconnect() (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.device != nil {
		err = d.device.Close()
	}
	return
}


