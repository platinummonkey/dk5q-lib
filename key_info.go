package dk5q_lib

import (
	"os"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"regexp"
)

type LED struct {
	RedChannel byte
	BlueChannel byte
	GreenChannel byte
	Zone int
	ID int
}

func NewLED(ID int) LED {
	// figure out which zone we're in. Why is this so crazy?
	if (14 <= ID  && ID <= 17) || (34 <= ID && ID <= 40) || (58 <= ID && ID <= 63) || (81 <= ID && ID <= 90) || (106 <= ID && ID <= 111) || (130 <= ID && ID <= 135) || (155 <= ID && ID <= 160) {
		// zone 2
		// color1 = blue
		// color2 = red
		// color3 = green
		return LED{ID: ID, RedChannel: 1, GreenChannel: 2, BlueChannel: 0, Zone: 2}
	} else if (18 <= ID && ID <= 23) || (41 <= ID && ID <= 47) || (64 <= ID && ID <= 71) || (91 <= ID && ID <= 95) || (115 <= ID && ID <= 119) || (137 <= ID && ID <= 143) || (161 <= ID && ID <= 167) || (ID == 191) || (193 <= ID && ID <= 215) {
		// zone 3
		// color1 = green
		// color2 = blue
		// color3 = red
		return LED{ID: ID, RedChannel: 2, GreenChannel: 0, BlueChannel: 1, Zone: 3}
	} else {
		// zone 1
		// color1 = red
		// color2 = green
		// color3 = blue
		return LED{ID: ID, RedChannel: 0, GreenChannel: 1, BlueChannel: 2, Zone: 1}
	}
}

type KeyCoordinates struct {
	X float32
	Y float32
}

type KeyModelInterface interface {
	LEDID() int
	LED() LED
	LEDZone() int
	RGBChannels() [3]byte
	Description() string
	ShortName() string
	TopLeftCoordinates() KeyCoordinates
	Width() float32
	Height() float32
}

type KeyModel struct {
	led LED
	description string
	shortName string
	topLeftCoordinates KeyCoordinates
	width float32
	height float32
}

func NewKeyModel(ledID int, description string, shortName string, topLeftCoordinates KeyCoordinates, width float32, height float32) KeyModel {
	return KeyModel{
		led: NewLED(ledID),
		description: description,
		shortName: shortName,
		topLeftCoordinates: topLeftCoordinates,
		width: width,
		height: height,
	}
}

func (k KeyModel) LEDID() int {
	return k.led.ID
}

func (k KeyModel) LED() LED {
	return k.led
}

func (k KeyModel) LEDZone() int {
	return k.led.Zone
}

func (k KeyModel) RGBChannels() [3]byte {
	var rgb [3]byte
	rgb[0] = k.led.RedChannel
	rgb[1] = k.led.GreenChannel
	rgb[2] = k.led.BlueChannel
	return rgb
}

func (k KeyModel) Description() string {
	return k.description
}

func (k KeyModel) ShortName() string {
	return k.shortName
}

func (k KeyModel) TopLeftCoordinates() KeyCoordinates {
	return k.topLeftCoordinates
}

func (k KeyModel) Width() float32 {
	return k.width
}

func (k KeyModel) Height() float32 {
	return k.height
}

type jsonKeyDef struct {
	LEDIDs []int `json:"ledIds"`
	Description string `json:"description"`
	ShortName string `json:"shortName"`
	TopLeftCoordinates map[string]float32 `json:"topLeftCoordinates"`
	Width float32 `json:"width"`
	Height float32 `json:"height"`
}

type KeyMap struct {
	Keys []KeyModel
	KeyMap map[string]KeyModel
}

func NewKeyModelsFromAsset(layout string) (KeyMap, error) {
	keys := make([]KeyModel, 0)
	filename := fmt.Sprintf("./keyboard_layouts/%s.json", layout)
	_, err := os.Stat(filename)
	switch {
	case err == nil:
		raw, err := ioutil.ReadFile(filename)
		if err != nil {
			return KeyMap{}, err
		}
		var rawDefs []jsonKeyDef
		err = json.Unmarshal(raw, &rawDefs)
		if err != nil {
			return KeyMap{}, err
		}
		for _, def := range rawDefs {
			keys = append(keys, NewKeyModel(def.LEDIDs[0], def.Description, def.ShortName, KeyCoordinates{X: def.TopLeftCoordinates["x"], Y: def.TopLeftCoordinates["y"]}, def.Width, def.Height))
		}
		keyMap := KeyMap{
			Keys: keys,
			KeyMap: make(map[string]KeyModel, 0),
		}
		re := regexp.MustCompile(`\W`)
		for _, key := range keys {
			keyName := re.ReplaceAllString(key.Description(), "")
			if keyName != "" {
				keyMap.KeyMap[keyName] = key
			}
		}

		return keyMap, nil
	case os.IsNotExist(err):
		return KeyMap{}, fmt.Errorf("no such format found")
	default:
		return KeyMap{}, err
	}
}