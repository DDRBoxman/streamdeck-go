package streamdeck

import (
	"image"
	"log"
	"time"

	"github.com/karalabe/hid"
)

var NUM_KEYS = 15
var ICON_SIZE = 72

var NUM_FIRST_PAGE_PIXELS = 2583
var NUM_SECOND_PAGE_PIXELS = 2601

var reset = []byte{0x0B, 0x63, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

var brightnessBytes = []byte{0x05, 0x55, 0xAA, 0xD1, 0x01, 0, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

var streamDeck *hid.Device

type StreamDeck struct {
	DeviceInfo hid.DeviceInfo
	Device *hid.Device
}

func FindDecks() []StreamDeck {
	devices := hid.Enumerate(4057, 96)

	decks := []StreamDeck{}
	for _, deviceInfo := range devices {
		decks = append(decks, StreamDeck{
			DeviceInfo: deviceInfo,
		})
	}

	return decks
}

func (deck *StreamDeck) Open() error {
	var err error
	deck.Device, err = deck.DeviceInfo.Open()
	return err
}

func (deck *StreamDeck) Close() {
	deck.Device.Close()
}

func (deck *StreamDeck) Reset() error {
	_, err := deck.Device.SendFeatureReport(reset)
	return err
}

func (deck *StreamDeck) SetBrightness(brightness int) error {
	brightnessBytes[5] = byte(brightness)

	_, err := deck.Device.SendFeatureReport(brightnessBytes)
	return err
}

func (deck *StreamDeck) WriteImageToKey(image *image.RGBA, key int) {
	pixels := make([]byte, ICON_SIZE*ICON_SIZE*3)

	for r := 0; r < ICON_SIZE; r++ {
		rowStartImage := r * ICON_SIZE * 4
		rowStartPixels := r * ICON_SIZE * 3
		for c := 0; c < ICON_SIZE; c++ {
			colPosImage := (c * 4) + rowStartImage
			colPosPixels := (ICON_SIZE * 3) + rowStartPixels - (c * 3) - 1

			pixels[colPosPixels-2] = image.Pix[colPosImage+2]
			pixels[colPosPixels-1] = image.Pix[colPosImage+1]
			pixels[colPosPixels] = image.Pix[colPosImage]
		}
	}

	writePage1(deck.Device, key, pixels[0:NUM_FIRST_PAGE_PIXELS*3])
	writePage2(deck.Device, key, pixels[NUM_FIRST_PAGE_PIXELS*3:])
}

func writePage1(sd *hid.Device, key int, pixels []byte) error {
	header := []byte{
		0x02, 0x01, 0x01, 0x00, 0x00, (byte)(key + 1), 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x42, 0x4d, 0xf6, 0x3c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x36, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x48, 0x00, 0x00, 0x00, 0x48, 0x00,
		0x00, 0x00, 0x01, 0x00, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xc0, 0x3c, 0x00, 0x00, 0xc4, 0x0e,
		0x00, 0x00, 0xc4, 0x0e, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	header = append(header, pixels...)

	data := make([]byte, 8191)

	copy(data, header)

	_, err := sd.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func writePage2(sd *hid.Device, key int, pixels []byte) error {
	header := []byte{
		0x02, 0x01, 0x02, 0x00, 0x01, (byte)(key + 1),
	}

	padding := make([]byte, 10)

	header = append(header, padding...)
	header = append(header, pixels...)

	data := make([]byte, 8191)

	copy(data, header)

	_, err := sd.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func readLoop(sd *hid.Device) {
	data := make([]byte, 255)
	for {
		time.Sleep(1000)
		size, err := sd.Read(data)
		if err != nil {
			log.Println("Failed to read from streamdeck: ", err)
			continue
		}

		log.Println(size)
		log.Println(data)
	}
}
