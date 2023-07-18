package main

import (
	"embed"
	"fmt"
	"image/color"
	"machine"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"golang.org/x/image/bmp"
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

	screenA = 'A'
	screenB = 'B'
	screenC = 'C'
)

// Globals!
var (
	led     machine.Pin
	display uc8151.Device

	//go:embed fs/smile.bmp fs/contact.txt fs/bitcoin.txt
	fsFiles embed.FS

	// Color vars
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

	button_up := machine.BUTTON_UP
	button_up.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	button_down := machine.BUTTON_DOWN
	button_down.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

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

	buttonPressedABC := false
	buttonPressedUpDown := false
	screen := screenA
	updown := false

	drawScreenA(updown)

	for {
		buttonPressedABC = button_a.Get() || button_b.Get() || button_c.Get()
		buttonPressedUpDown = button_up.Get() || button_down.Get()

		if button_up.Get() || button_down.Get() {
			updown = !updown
		}

		if button_a.Get() {
			screen = screenA
		} else if button_b.Get() {
			screen = screenB
		} else if button_c.Get() {
			screen = screenC
		}

		if buttonPressedABC || buttonPressedUpDown {
			fmt.Printf("buttonPressedABC: %v, buttonPressedUpDown: %v, updown: %v\n", buttonPressedABC, buttonPressedUpDown, updown)
			led.Set(true)

			if buttonPressedABC {
				// Ensure switching (A,B,C) screen always starts on down screen
				updown = false
			}

			switch screen {
			case screenA:
				drawScreenA(updown)
			case screenB:
				drawScreenB(updown)
			case screenC:
				drawScreenC(updown)
			}
		}

		led.Set(false)
		time.Sleep(time.Second / 4)
	}
}

func drawScreenA(alt bool) {
	display.ClearBuffer()
	display.ClearDisplay()

	drawBitmap("fs/smile.bmp", 168+12, 0)
	drawContactTxt("fs/contact.txt", alt)

	display.Display()
}

func drawScreenB(alt bool) {
	display.ClearBuffer()
	display.ClearDisplay()

	title := "Bitcoin!"

	address, comment := getBitcoinAddress("fs/bitcoin.txt", alt)
	if address == "" {
		return
	}

	drawQRCode(address, 0, 0)

	tinyfont.WriteLine(&display, &freemono.Bold12pt7b, 128, 30, title, BLACK)

	tinyfont.WriteLine(&display, &freemono.Regular9pt7b, 128, 55, address[0:6], BLACK)
	tinyfont.WriteLine(&display, &freemono.Regular9pt7b, 128+(9*7), 55, "...", BLACK)
	tinyfont.WriteLine(&display, &freemono.Regular9pt7b, 128+(9*7)+(9*4), 55, address[len(address)-6:], BLACK)

	tinyfont.WriteLine(&display, &freemono.Regular9pt7b, 128, 80, comment, BLACK)

	display.Display()
}

func getBitcoinAddress(fn string, alt bool) (string, string) {
	addressBytes, err := fsFiles.ReadFile(fn)
	if err != nil {
		fmt.Printf("ERROR: Unable to open file: %v\n", err)
		return "", ""
	}

	addressLines := strings.Split(string(addressBytes), "\n")

	fmt.Println(addressLines)

	line := 0
	if alt {
		line = 1
	}

	firstSpaceIdx := strings.Index(addressLines[line], " ")

	return addressLines[line][0:firstSpaceIdx], addressLines[line][firstSpaceIdx:]
}

func drawScreenC(alt bool) {
	display.ClearBuffer()
	display.ClearDisplay()

	if alt {
		tinyfont.WriteLineRotated(&display, &freemono.Bold24pt7b, 8, 30, "CLT", BLACK, tinyfont.NO_ROTATION)
	} else {
		tinyfont.WriteLineRotated(&display, &freemono.Bold24pt7b, 8, 30, "CCC", BLACK, tinyfont.NO_ROTATION)
	}

	display.Display()
}

func drawContactTxt(fn string, alt bool) {
	contactBytes, err := fsFiles.ReadFile(fn)
	if err != nil {
		fmt.Printf("ERROR: Unable to open file: %v\n", err)
		return
	}

	contactLines := strings.Split(string(contactBytes), "\n")

	tinyfont.WriteLine(&display, &freemono.Bold12pt7b, 0, 30, contactLines[0], BLACK)
	tinyfont.WriteLine(&display, &freemono.Bold9pt7b, 0, 55, contactLines[1], BLACK)

	if alt {
		tinyfont.WriteLine(&display, &freemono.Bold9pt7b, 0, 80, contactLines[2], BLACK)
	}
}

func drawBitmap(fn string, x, y int) {
	reader, err := fsFiles.Open(fn)
	if err != nil {
		fmt.Printf("ERROR: Unable to open file: %v\n", err)
		return
	}

	img, err := bmp.Decode(reader)
	if err != nil {
		fmt.Printf("ERROR: Unable to bmp.Decode: %v\n", err)
		return
	}

	bounds := img.Bounds()

	for ix := 0; ix < bounds.Dx(); ix++ {
		for iy := 0; iy < bounds.Dy(); iy++ {
			c := WHITE
			r, g, b, _ := img.At(ix, iy).RGBA()
			if r+g+b < (128 * 3) {
				c = BLACK
			}

			display.SetPixel(int16(x+ix), int16(y+iy), c)
		}
	}
}

// drawQRCode draw a QR code
// todo size not controlled
func drawQRCode(content string, x, y int) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		fmt.Printf("ERROR: Unable to qrcode.New: %v\n", err)
		return
	}

	bmp := qr.Bitmap()
	for bX := range bmp {
		for bY := range bmp[bX] {
			c := WHITE
			if bmp[bX][bY] {
				c = BLACK
			}

			const sF = 3
			sX := (x + bX) * sF
			sY := (y + bY) * sF
			for iX := 0; iX < sF; iX++ {
				for iY := 0; iY < sF; iY++ {
					display.SetPixel(int16(iX+sX), int16(iY+sY), c)
				}
			}
		}
	}
}
