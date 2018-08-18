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
	UpIncrementDelay uint16
	UpHoldLevel uint16
	UpHoldDelay uint16
	DownMinimumLevel uint16
	DownDecrement uint16
	DownDecrementDelay uint16
	DownHoldLevel uint16
	DownHoldDelay uint16
	StartDelay uint16
}

func NewStateInfo(keyID uint8, colorChannelID uint8) *StateInfo {
	return &StateInfo{
		Key: keyID,
		EffectID: 0x00,
		EffectFlag: EffectFlagDefaultValue,
		ColorChannelID: colorChannelID,
	}
}

type KeyState struct {
	redState *StateInfo
	greenState *StateInfo
	blueState *StateInfo
	key *KeyModel
}

func NewKeyState(key *KeyModel) KeyState {
	keyChannels := key.RGBChannels()
	return KeyState{
		key: key,
		redState: NewStateInfo(uint8(key.LEDID()), keyChannels[0]),
		greenState: NewStateInfo(uint8(key.LEDID()), keyChannels[0]),
		blueState: NewStateInfo(uint8(key.LEDID()), keyChannels[0]),
	}
}

func (s *KeyState) SetFromColorRGB(red uint16, green uint16, blue uint16) {
	s.redState.DownHoldLevel = red
	s.greenState.DownHoldLevel = green
	s.blueState.DownHoldLevel = blue
}

func (s *KeyState) SetToColorRGB(red uint16, green uint16, blue uint16) {
	s.redState.UpHoldLevel = red
	s.greenState.UpHoldLevel = green
	s.blueState.UpHoldLevel = blue
}

func (s *KeyState) SetUpMaximum(red uint16, green uint16, blue uint16) {
	s.redState.UpMaximumLevel = red
	s.greenState.UpMaximumLevel = green
	s.blueState.UpMaximumLevel = blue
}

func (s *KeyState) SetDownMinimum(red uint16, green uint16, blue uint16) {
	s.redState.DownMinimumLevel = red
	s.greenState.DownMinimumLevel = green
	s.blueState.DownMinimumLevel = blue
}

func (s *KeyState) SetUpHoldDelay(delay uint16) {
	s.redState.UpHoldDelay = delay
	s.greenState.UpHoldDelay = delay
	s.blueState.UpHoldDelay = delay
}

func (s *KeyState) SetDownHoldDelay(delay uint16) {
	s.redState.DownHoldDelay = delay
	s.greenState.DownHoldDelay = delay
	s.blueState.DownHoldDelay = delay
}

func (s *KeyState) SetUpIncrement(increment uint16) {
	s.redState.UpIncrement = increment
	s.greenState.UpIncrement = increment
	s.blueState.UpIncrement = increment
}

func (s *KeyState) SetDownDecrement(decrement uint16) {
	s.redState.DownDecrement = decrement
	s.greenState.DownHoldDelay = decrement
	s.blueState.DownHoldDelay = decrement
}

func (s *KeyState) SetUpIncrementDelay(delay uint16) {
	s.redState.UpIncrementDelay = delay
	s.greenState.UpIncrementDelay = delay
	s.blueState.UpIncrementDelay = delay
}

func (s *KeyState) SetDownDecrementDelay(delay uint16) {
	s.redState.DownDecrementDelay = delay
	s.greenState.DownDecrementDelay = delay
	s.blueState.DownDecrementDelay = delay
}

func (s *KeyState) SetStartDelay(delay uint16) {
	s.redState.StartDelay = delay
	s.greenState.StartDelay = delay
	s.blueState.StartDelay = delay
}

func (s *KeyState) SetMoveUp(delay uint16) {
	s.redState.EffectFlag = EffectFlagIncrementOnly(s.redState.EffectFlag)
	s.greenState.EffectFlag = EffectFlagIncrementOnly(s.greenState.EffectFlag)
	s.blueState.EffectFlag = EffectFlagIncrementOnly(s.blueState.EffectFlag)
}

func (s *KeyState) SetMoveDown(delay uint16) {
	s.redState.EffectFlag = EffectFlagDecrementOnly(s.redState.EffectFlag)
	s.greenState.EffectFlag = EffectFlagDecrementOnly(s.greenState.EffectFlag)
	s.blueState.EffectFlag = EffectFlagDecrementOnly(s.blueState.EffectFlag)
}

func (s *KeyState) SetTransition(delay uint16) {
	s.redState.EffectFlag = EffectFlagIncrementDecrement(s.redState.EffectFlag)
	s.greenState.EffectFlag = EffectFlagIncrementDecrement(s.greenState.EffectFlag)
	s.blueState.EffectFlag = EffectFlagIncrementDecrement(s.blueState.EffectFlag)
}

func (s *KeyState) SetTransitionReverse(delay uint16) {
	s.redState.EffectFlag = EffectFlagDecrementIncrement(s.redState.EffectFlag)
	s.greenState.EffectFlag = EffectFlagDecrementIncrement(s.greenState.EffectFlag)
	s.blueState.EffectFlag = EffectFlagDecrementIncrement(s.blueState.EffectFlag)
}

func (s *KeyState) SetApplyImmediately(delay uint16) {
	s.redState.EffectFlag = EffectFlagTriggerNow(s.redState.EffectFlag)
	s.greenState.EffectFlag = EffectFlagTriggerNow(s.greenState.EffectFlag)
	s.blueState.EffectFlag = EffectFlagTriggerNow(s.blueState.EffectFlag)
}

func (s *KeyState) SetApplyDelayed(delay uint16) {
	s.redState.EffectFlag = EffectFlagTriggerOnApply(s.redState.EffectFlag)
	s.greenState.EffectFlag = EffectFlagTriggerOnApply(s.greenState.EffectFlag)
	s.blueState.EffectFlag = EffectFlagTriggerOnApply(s.blueState.EffectFlag)
}

func (s *KeyState) SetToHardwareProfile() {
	s.redState.EffectID = 0x00
	s.greenState.EffectID = 0x00
	s.blueState.EffectID = 0x00
}

func (s *KeyState) BuildStatePackets(info *hid.DeviceInfo) (packets [][]byte) {
	packets = make([][]byte, 3)
	packets[0] = StatePacket(info, *s.redState)
	packets[1] = StatePacket(info, *s.greenState)
	packets[2] = StatePacket(info, *s.blueState)
	return packets
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
		state.UpIncrementDelay,
		state.UpHoldLevel,
		state.UpHoldDelay,
		state.DownMinimumLevel,
		state.DownDecrement,
		state.DownDecrementDelay,
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