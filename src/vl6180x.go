package main

import (
	"fmt"
	smbus "github.com/corrupt/go-smbus"
	"time"
)

const vlRegSystemFreshOutReset = 0x0016
const vlRegSysRangeStart       = 0x0018
const vlRegResultRangeVal      = 0x0062
const vlRegResultIntStatus     = 0x004F
const vlRegSystemIntClear      = 0x0015

const MUX_ADDR = 0x70
const VL6180X_ADDR  = 0x29

type VL6180X struct {
	bus      *smbus.SMBus
	muxBus   *smbus.SMBus
	address  uint8
	channels [2]uint8
}

// Write to 16bit register
func (v *VL6180X) writeReg(reg uint16, data uint8) error {
	_, err := v.bus.Write_i2c_block_data(uint8(reg>>8), []byte{uint8(reg & 0xFF), data})
	return err
}

// Read from 16bit register
func (v *VL6180X) readReg(reg uint16) (uint8, error) {
	if _, err := v.bus.Write_i2c_block_data(uint8(reg>>8), []byte{uint8(reg & 0xFF)}); err != nil {
		return 0, fmt.Errorf("reading: %v", err)
	}
	buf := make([]byte, 1)
	if _, err := v.bus.Read_i2c_block_data(uint8(reg>>8), buf); err != nil {
		return 0, fmt.Errorf("reading: %v", err)
	}
	return buf[0], nil
}

// Toggle to the specified channel (in case of using a multiplexer)
func (v *VL6180X) selectChannel(channel uint8) error {
	return v.muxBus.Write_byte(1 << channel)
}

// Initialize registers on startup
func (v *VL6180X) initRegisters() error {
	mandatory := [][2]uint16{
		{0x0207, 0x01}, {0x0208, 0x01}, {0x0096, 0x00}, {0x0097, 0xFD},
		{0x00E3, 0x01}, {0x00E4, 0x03}, {0x00E5, 0x02}, {0x00E6, 0x01},
		{0x00E7, 0x03}, {0x00F5, 0x02}, {0x00D9, 0x05}, {0x00DB, 0xCE},
		{0x00DC, 0x03}, {0x00DD, 0x01}, {0x009F, 0x00}, {0x00A3, 0x3C},
		{0x00B7, 0x00}, {0x00BB, 0x3C}, {0x00B2, 0x09}, {0x00CA, 0x09},
		{0x0198, 0x01}, {0x01B0, 0x17}, {0x01AD, 0x00}, {0x00FF, 0x05},
		{0x0100, 0x05}, {0x0199, 0x05}, {0x01A6, 0x1B}, {0x01AC, 0x3E},
		{0x01A7, 0x1F}, {0x0030, 0x00},
	}
	for _, rv := range mandatory {
		if err := v.writeReg(rv[0], uint8(rv[1])); err != nil {
			return fmt.Errorf("initRegisters reg: %v", err)
		}
	}

	return v.writeReg(vlRegSystemFreshOutReset, 0x00)
}

// Read in the distance for the given channel index (0 or 1)
func (v *VL6180X) ReadDistance(channel uint8) (int, error) {
	if v.muxBus != nil {
		if err := v.selectChannel(v.channels[channel]); err != nil {
			return 0, fmt.Errorf("selecting channel: %w", err)
		}
	}

	if err := v.writeReg(vlRegSysRangeStart, 0x01); err != nil {
		return 0, fmt.Errorf("writing: %w", err)
	}

	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		status, err := v.readReg(vlRegResultIntStatus)
		if err != nil {
			return 0, fmt.Errorf("status: %w", err)
		}
		if status&0x07 == 0x04 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}

	val, err := v.readReg(vlRegResultRangeVal)
	if err != nil {
		return 0, fmt.Errorf("read: %w", err)
	}

	_ = v.writeReg(vlRegSystemIntClear, 0x07)
	return int(val), nil
}

// Initialize the sensor, depending on whether we are using a multiplexer
func Initialize(bus uint, channels [2]uint8) (*VL6180X, error) {
	sensorBus, err := smbus.New(bus, VL6180X_ADDR)
	if err != nil {
		return nil, fmt.Errorf("open sensor bus: %w", err)
	}

	vl := &VL6180X{
		bus:      sensorBus,
		address:  VL6180X_ADDR,
		channels: channels,
		muxBus:   nil,
	}

	
	muxBus, err := smbus.New(bus, MUX_ADDR)
	if err != nil {
		return nil, fmt.Errorf("open mux bus: %w", err)
	}
	vl.muxBus = muxBus

	if err := vl.selectChannel(channels[0]); err != nil {
		return nil, fmt.Errorf("selectChannel during init: %w", err)
	}
	

	reset, err := vl.readReg(vlRegSystemFreshOutReset)
	if err != nil {
		return nil, fmt.Errorf("reset: %w", err)
	}
	if reset == 0x01 {
		if err := vl.initRegisters(); err != nil {
			return nil, fmt.Errorf("initRegisters: %w", err)
		}
	}

	return vl, nil
}