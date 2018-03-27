// Binary blaster reads/writes the EEPROM of the kinX’s CY7C65632 USB 2.0 hub.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/gousb"
)

var (
	write = flag.Bool("write",
		false,
		"Write the default config instead of reading and displaying the current config")
	raw = flag.Bool("raw",
		false,
		"Print the raw EEPROM bytes to stdout instead of parsing")
)

const eepromRequest = 14

var defaultConfig = &config{
	VID:       0x04b4,
	PID:       0x6570,
	Removable: 0x18,
	Ports:     0x4,
	MaxPower:  0xfa, // 500mA
	Vendor:    "stapelberg",
	Product:   "kinX hub v2018-02-11",
	Serial:    "00050034031B",
}

func logic() error {
	usb := gousb.NewContext()
	defer usb.Close()
	dev, err := usb.OpenDeviceWithVIDPID(0x04b4, 0x6570)
	if err != nil {
		return err
	}
	log.Printf("device = %+v", dev)
	if *write {
		b, err := defaultConfig.Marshal()
		if err != nil {
			return err
		}
		log.Printf("writing EEPROM (takes about 3s)")
		for wIndex := uint16(0); wIndex < 64; wIndex++ {
			n, err := dev.Control(gousb.RequestTypeVendor, eepromRequest, 0, wIndex, b[wIndex*2:wIndex*2+2])
			if err != nil {
				return err
			}
			if got, want := n, 2; got != want {
				return fmt.Errorf("protocol error: unexpected response length: got %d, want %d", got, want)
			}
			// Must not overwhelm the device by sending too quickly, otherwise
			// writes will silently fail:
			time.Sleep(10 * time.Millisecond)
		}
	} else {
		eeprom := make([]byte, 128)
		for wIndex := uint16(0); wIndex < 64; wIndex++ {
			data := make([]byte, 2)
			n, err := dev.Control(gousb.RequestTypeVendor|0x80, eepromRequest, 0, wIndex, data)
			if err != nil {
				return err
			}
			if got, want := n, 2; got != want {
				return fmt.Errorf("protocol error: unexpected response length: got %d, want %d", got, want)
			}
			copy(eeprom[wIndex*2:], data)
		}
		if *raw {
			io.Copy(os.Stdout, bytes.NewReader(eeprom))
			return nil
		}
		cfg, err := parse(eeprom)
		if err != nil {
			return err
		}
		fmt.Println(cfg)
	}
	return nil
}

func main() {
	flag.Parse()
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
