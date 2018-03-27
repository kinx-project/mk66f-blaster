package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type config struct {
	VID       uint16
	PID       uint16
	Removable byte
	Ports     uint8 // in [1, 4]
	MaxPower  byte  // current in 2mA increments, in [0x00, 0xfa]
	Vendor    string
	Product   string
	Serial    string
}

func (c *config) String() string {
	return fmt.Sprintf(`Vendor: 0x%x,
Product: 0x%x,
Removable: %v,
Ports: %v,
MaxPower: %dmA,
Vendor: %q,
Product: %q,
Serial: %q
`,
		c.VID,
		c.PID,
		c.Removable,
		c.Ports,
		int(c.MaxPower)*2,
		c.Vendor,
		c.Product,
		c.Serial)
}

var pad0xFF = bytes.Repeat([]byte{0xFF}, 64)

func (c *config) Marshal() ([]byte, error) {
	// See the CY7C65632 datasheet at
	// http://www.cypress.com/file/114101/download,
	// page 16 (EEPROM Configuration Options).

	output := make([]byte, 128)
	binary.LittleEndian.PutUint16(output[0:2], c.VID)
	binary.LittleEndian.PutUint16(output[2:4], c.PID)
	output[0x04] = byte(c.VID&0xFF) + byte((c.VID&0xFF00)>>8) +
		byte(c.PID&0xFF) + byte((c.PID&0xFF00)>>8) +
		1
	output[0x05] = 0xFE // reserved
	output[0x06] = c.Removable
	output[0x07] = c.Ports
	output[0x08] = c.MaxPower
	copy(output[0x09:0x0F+1], pad0xFF)

	output[0x10] = byte(len(c.Vendor))
	copy(output[0x11:0x3F+1], pad0xFF)
	copy(output[0x11:0x3F+1], []byte(c.Vendor))

	output[0x40] = byte(len(c.Product))
	copy(output[0x41:0x6F+1], pad0xFF)
	copy(output[0x41:0x6F+1], []byte(c.Product))

	output[0x70] = byte(len(c.Serial))
	copy(output[0x71:0x80], pad0xFF)
	copy(output[0x71:0x80], []byte(c.Serial))

	return output, nil
}

func parse(b []byte) (*config, error) {
	var cfg config

	cfg.VID = binary.LittleEndian.Uint16(b[0:2])
	cfg.PID = binary.LittleEndian.Uint16(b[2:4])
	cfg.Removable = b[0x06]
	cfg.Ports = b[0x07]
	cfg.MaxPower = b[0x08]
	vendorLen := b[0x10]
	if vendorLen == 0 || vendorLen > 47 {
		return nil, fmt.Errorf("invalid vendor length: got %d, want (0, 46]", vendorLen)
	}
	cfg.Vendor = string(b[17 : 17+vendorLen])
	productLen := b[0x40]
	if productLen == 0 || productLen > 47 {
		return nil, fmt.Errorf("invalid product length: got %d, want (0, 46]", productLen)
	}
	cfg.Product = string(b[65 : 65+productLen])
	serialLen := b[0x70]
	if serialLen == 0 || serialLen > 15 {
		return nil, fmt.Errorf("invalid serial length: got %d, want (0, 15]", serialLen)
	}
	cfg.Serial = string(b[0x71 : 0x71+serialLen])

	return &cfg, nil
}
