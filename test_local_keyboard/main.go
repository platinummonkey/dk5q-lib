package main

import (
	usb "github.com/platinummonkey/dk5q-lib"
	"fmt"
)

func fatalIf(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	keymodel, err := usb.NewKeyModelsFromAsset("en-us")
	fatalIf(err)
	fmt.Printf("Got keymodel: %v\n", keymodel.KeyMap)
	keyboard := usb.NewDefaultDasKeyboard(0x24f0, 0x2020, 2, 165)
	if keyboard == nil {
		fatalIf(fmt.Errorf("did not find keyboard"))
	}
	fatalIf(keyboard.Connect())
	fatalIf(keyboard.Initialize())
	err, firmware := keyboard.GetKeyboardData()
	fatalIf(err)
	fmt.Printf("Got firmware data: %v", firmware)
	fatalIf(keyboard.Disconnect())
}