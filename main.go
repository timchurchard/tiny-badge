package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/uc8151"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

// Preferences
const (
	// [DEFAULT, MEDIUM, FAST, TURBO]
	// DEFAULT is the best quality, but takes the longest to draw
	// TURBO is very fast, but *by far* the lowest quality
	EPDSPEED = uc8151.FAST
)

// Globals!
var (
	led     machine.Pin
	display uc8151.Device
)

// Color vars
var (
	WHITE = color.RGBA{0, 0, 0, 0}
	BLACK = color.RGBA{1, 1, 1, 255}
)

func main() {
	led = machine.LED
	led.Configure(
		machine.PinConfig{Mode: machine.PinOutput},
	)

	button_a := machine.BUTTON_A
	button_a.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	button_b := machine.BUTTON_B
	button_b.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	button_c := machine.BUTTON_C
	button_c.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 12000000,
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
	})

	display = uc8151.New(
		machine.SPI0,
		machine.EPD_CS_PIN,
		machine.EPD_DC_PIN,
		machine.EPD_RESET_PIN,
		machine.EPD_BUSY_PIN,
	)
	display.Configure(uc8151.Config{
		Blocking: true,
		Rotation: uc8151.ROTATION_270,
		Speed:    EPDSPEED,
	})
	// displayWidth, displayHeight := display.Size()

	drawScreenA()

	for {
		led.Set(false)

		if button_a.Get() {
			led.Set(true)
			drawScreenA()
		} else if button_b.Get() {
			led.Set(true)
			drawScreenB()
		} else if button_c.Get() {
			led.Set(true)
			drawScreenC()
		}

		time.Sleep(time.Second / 4)
	}
}

func drawScreenA() {
	display.ClearBuffer()
	display.ClearDisplay()

	tinyfont.WriteLineRotated(&display, &freemono.Bold24pt7b, 8, 30, "AAA", BLACK, tinyfont.NO_ROTATION)

	display.Display()
}

func drawScreenB() {
	display.ClearBuffer()
	display.ClearDisplay()

	tinyfont.WriteLineRotated(&display, &freemono.Bold24pt7b, 8, 30, "BBB", BLACK, tinyfont.NO_ROTATION)

	display.Display()
}

func drawScreenC() {
	display.ClearBuffer()
	display.ClearDisplay()

	tinyfont.WriteLineRotated(&display, &freemono.Bold24pt7b, 8, 30, "CCC", BLACK, tinyfont.NO_ROTATION)

	display.Display()
}
