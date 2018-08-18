package dk5q_lib

import (
	"github.com/platinummonkey/hid"
	"fmt"
	"bytes"
	"encoding/binary"
)

// Build the Initialization Packet
func InitializePacket(info *hid.DeviceInfo) (packet []byte) {
	// At this point we're really not sure exactly what this does.
	// However, we know it's very important. This is sent by the main service module.
	packet = []byte{
		0x00, 0x13, 0x00, 0x4d, 0x43, 0x49, 0x51, 0x46, 0x49, 0x46, 0x45, 0x44,
		0x4c, 0x48, 0x39, 0x46, 0x34, 0x41, 0x45, 0x43, 0x58, 0x39, 0x31, 0x36,
		0x50, 0x42, 0x44, 0x35, 0x50, 0x33, 0x41, 0x33, 0x30, 0x37, 0x38,
	}
	return
}

func BrightnessPacket(info *hid.DeviceInfo, brightness uint8) (packet []byte, err error) {
	if brightness > 63 || brightness < 0 {
		err = fmt.Errorf("brightness must be between [0,63]")
		return
	}

	order := binary.LittleEndian
	buff := bytes.NewBuffer(packet)
	binary.Write(buff, order, []uint8{0, 43, 0, brightness})
	packet = buff.Bytes()
	return
}

func FirmwarePacket(info *hid.DeviceInfo) (packet []byte) {
	// At this point we're really not sure exactly what this does.
	packet = []byte {
		0x00, 0x11, 0x06, 0x4d, 0x43, 0x49, 0x51, 0x46, 0x49, 0x46, 0x45, 0x44,
		0x4c, 0x48, 0x39, 0x46, 0x34, 0x41, 0x45, 0x43, 0x58, 0x39, 0x31, 0x36, 0x50,
		0x42, 0x44, 0x35, 0x50, 0x33, 0x41, 0x33, 0x30, 0x37, 0x38,
	}
	return
}

func FreezePacket(info *hid.DeviceInfo) (packet []byte) {
	// At this point we're really not sure exactly what this does.
	buf := bytes.NewBuffer(packet)
	order := binary.LittleEndian
	binary.Write(buf, order, []uint8{0, 45, 0, 7})
	binary.Write(buf, order,
		[]uint16{65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535,65535, 65535, 65535})
	packet = buf.Bytes()
	return
}

type StateInfo struct {
	Key uint8
	ColorChannelID uint8
	EffectID uint8
	EffectFlag uint16
	UpMaximumLevel uint16
	UpIncrement uint16
	UpHoldLevel uint16
	UpHoldDelay uint16
	DownMinimumLevel uint16
	DownDecrement uint16
	DownHoldLevel uint16
	DownHoldDelay uint16
	StartDelay uint16
}

func StatePacket(info *hid.DeviceInfo, state StateInfo) (packet []byte) {
	buf := bytes.NewBuffer(packet)
	order := binary.LittleEndian
	if state.Key == 0 {
		state.Key = 151
	}
	binary.Write(buf, order, []uint8{0, 40, 0, state.ColorChannelID, 1, state.Key, state.EffectID})
	binary.Write(buf, order, []uint16{
		state.UpMaximumLevel,
		state.UpIncrement,
		state.UpHoldLevel,
		state.UpHoldDelay,
		state.DownMinimumLevel,
		state.DownDecrement,
		state.DownHoldLevel,
		state.DownHoldDelay,
		state.StartDelay,
		0,
		state.EffectFlag,
	})
	packet = buf.Bytes()
	return
}

func TriggerPacket(info *hid.DeviceInfo) (packet []byte) {
	buf := bytes.NewBuffer(packet)
	order := binary.LittleEndian
	binary.Write(buf, order, []uint8{0, 45, 0, 15})
	binary.Write(buf, order,
		[]uint16{65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535, 65535})
	packet = buf.Bytes()
	return
}

// EffectFlags
const EffectFlagDefaultValue uint16 = 1

func EffectFlagIncrementOnly(value uint16) uint16 {
	return 1
}

func EffectFlagDecrementOnly(value uint16) uint16 {
	return 2
}
func EffectFlagIncrementDecrement(value uint16) uint16 {
	return 25
}
func EffectFlagDecrementIncrement(value uint16) uint16 {
	return 26
}
func EffectFlagTriggerOnApply(value uint16) uint16 {
	return value | 16384
}
func EffectFlagTriggerNow(value uint16) uint16 {
	return value &^ 16384
}
func EffectFlagEnableTransition(value uint16) uint16 {
	return value &^ 4096
}
func EffectFlagDisableTransition(value uint16) uint16 {
	return value | 4096
}