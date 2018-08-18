package dk5q_lib

import "fmt"

// DasKeyboard defines the keyboard interface should model versions change significantly.
type DasKeyboard interface {
	// Connect will connect to the keyboard
	Connect() (err error)
	// Initialize a connection
	Initialize() (err error)
	// FeatureReport
	FeatureReport(report []byte) (data []byte, err error)
	// Read will read data from the keyboard
	Read() (data []byte, err error)
	// Write will write data to the keyboard
	Write(data []byte) (err error)
	// Disconnect will disconnect from the keyboard
	Disconnect() (err error)
}

type DasKeyboardFirmwareInfo struct {
	MajorVersion byte
	MinorVersion byte
	PatchVersion byte
	RCVersion byte
	PacketCount byte
}

func (d DasKeyboardFirmwareInfo) String() string {
	return fmt.Sprintf("%v.%v.%v.%v - (%v)", d.MajorVersion, d.MinorVersion, d.PatchVersion, d.RCVersion, d.PacketCount)
}

// FindDasKeyboard will search for the given keyboard and return it's implementation
func FindDasKeyboard(vendorID uint16, productID uint16, deviceInterface uint16, usage uint16) (DasKeyboard, error) {
	// For now we only have one implementation.
	panic("implement me")
}
