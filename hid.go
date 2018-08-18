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

// At this point we're really not sure exactly what this does.
// However, we know it's very important. This is sent by the main service module.
var initializePacket = []byte{0x00, 0x13, 0x00, 0x4d, 0x43, 0x49, 0x51, 0x46, 0x49, 0x46, 0x45, 0x44,
	0x4c, 0x48, 0x39, 0x46, 0x34, 0x41, 0x45, 0x43, 0x58, 0x39, 0x31, 0x36,
	0x50, 0x42, 0x44, 0x35, 0x50, 0x33, 0x41, 0x33, 0x30, 0x37, 0x38}

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
	firmwareVersionPacket, err := d.FeatureReport(0x00, FirmwarePacket(d.deviceInfo))
	if err != nil {
		return
	}
	fmt.Printf("Got firmware packet: %v\n", firmwareVersionPacket)
	startIdx := 2
	info.MajorVersion = firmwareVersionPacket[startIdx+1]
	info.MinorVersion = firmwareVersionPacket[startIdx+2]
	info.PatchVersion = firmwareVersionPacket[startIdx+3]
	info.RCVersion = firmwareVersionPacket[startIdx+4]
	info.PacketCount = firmwareVersionPacket[startIdx]
	fmt.Printf("\t Got version %v\n", info)
	return
}

const maxBufSize = 65

func (d *DefaultDasKeyboard) FeatureReport(reportID uint16, report []byte) (result []byte, err error) {
	buf := make([]byte, 0)
	buf = append(buf, byte(reportID))
	buf = append(buf, report[1:]...)
	//buf = append(buf, report...)
	for len(buf) < maxBufSize {
		buf = append(buf, 0)
	}
	d.mu.Lock()
	sequence := byte(d.sequence)
	d.mu.Unlock()
	buf[3] = sequence
	bufCopy := make([]byte, len(buf))
	for i, b := range buf {
		bufCopy[i] = b
	}
	bytesWritten, err := d.device.GetFeatureReport(buf)
	if err != nil {
		fmt.Printf("could not write feature report: (%d) %v\n\tfinal buf: (%d) %v\n", len(bufCopy), bufCopy, len(buf), buf)
		return
	}
	fmt.Printf("wrote %d bytes in feature report: (%d) %v\n\tfinal buf: (%d) %v\n", bytesWritten, len(bufCopy), bufCopy, len(buf), buf)
	d.mu.Lock()
	if d.sequence == 0xFF {
		d.sequence = 0x00
	} else {
		d.sequence += 1
	}
	d.mu.Unlock()

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

		//now := time.Now()
		bytesWritten, err = d.device.Write(data)
		fmt.Printf("\tgot error: %v\n", err)

		//duration := time.Now().Sub(now).Nanoseconds()
		//if duration < 10000 {
		//	// missing device errors will return almost immediately
		//	err = fmt.Errorf("feature report errored to quickly, likely the device is removed: %v", err)
		//	return
		//}
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


