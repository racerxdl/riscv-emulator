package vga

import (
	"context"
	"fmt"
	"github.com/racerxdl/riscv-emulator/core"
	"image/color"
	"sync"
)

const (
	ControlAddressOffset = 0x00000
	PaletteAddressOffset = 0x10000
	ScreenAddressOffset  = 0x20000 // Palette Size
	ControlMapName       = "vga_ctrl"
	ScreenMapName        = "vga_screen"
	PaletteMapName       = "vga_palette"
)

// VGA is a device that simulates a video adapter with a palette color source
type VGA struct {
	sync.RWMutex
	palette    [256]color.RGBA
	screen     []uint8
	width      int
	height     int
	frameCount uint32
	vblank     uint32
}

// NewVGA creates and initialize a new VGA Device
func NewVGA(width, height int) *VGA {
	v := &VGA{
		width:  width,
		height: height,
		screen: make([]uint8, width*height),
	}

	for i := 0; i < 256; i++ {
		v.palette[i] = color.RGBA{R: uint8(i), G: uint8(i), B: uint8(i), A: 255}
	}

	return v
}

// ReadStatus return the status register
func (vga *VGA) ReadStatus(address uint32) (uint32, error) {
	v := (vga.frameCount & 0xFFFF) | (vga.vblank << 16)
	//fmt.Printf("STATUS: %032b - Frame Count: %d - VBLANK %d\n", v, vga.frameCount, vga.vblank)
	return v, nil
}

// WriteScreen writes to the screen buffer
// This call ignores mask, and always use the LSB byte of the uint32
// The address maps to each point in the screen, even though they're only one byte wide
// So for acessing point X of the screen, you should give address = X * 4
func (vga *VGA) WriteScreen(address uint32, value uint32, mask uint8) error {
	vga.Lock()
	defer vga.Unlock()

	//bval := uint8(value & 0xFF)
	if uint32(len(vga.screen)) < address {
		return fmt.Errorf("invalid write at screen address %08x", address)
	}

	vga.screen[address] = uint8(value & 0xFF)
	vga.screen[address+1] = uint8((value & 0xFF00) >> 8)
	vga.screen[address+2] = uint8((value & 0xFF0000) >> 16)
	vga.screen[address+3] = uint8((value & 0xFF000000) >> 24)
	//fmt.Printf("Wrote Screen at %d with %d (mask %d)\n", address, bval, mask)
	return nil
}

// WriteScreen writes to the palette buffer
// This call only accepts full word write and doesn't accept unaligned access
func (vga *VGA) WritePAL(address uint32, value uint32, mask uint8) error {
	vga.Lock()
	defer vga.Unlock()
	if address%4 != 0 {
		return fmt.Errorf("unaligned access at palette write address %08x", address)
	}
	if mask != 15 {
		return fmt.Errorf("unsupported mask %04b at palette write address %08x", mask, address)
	}
	address /= 4

	if address >= 256 {
		return fmt.Errorf("invalid write at palette address %08x", address*4)
	}

	vga.palette[address] = color.RGBA{
		B: uint8((value >> 0) & 0xFF),
		G: uint8((value >> 8) & 0xFF),
		R: uint8((value >> 16) & 0xFF),
		A: 255,
		//A: uint8((value >> 24) & 0xFF),
	}
	return nil
}

// ReadScreen reads from screen buffer
// It does not support unaligned access, and each screen pixel is mapped as a word
// So for acessing pixel X you should access X * 4
func (vga *VGA) ReadScreen(address uint32) (uint32, error) {
	//fmt.Printf("ReadScreen at %08x\n", address)
	vga.RLock()
	defer vga.RUnlock()

	if address%4 != 0 {
		return 0, fmt.Errorf("unaligned access at screen read address %08x", address)
	}
	address /= 4

	if uint32(len(vga.screen)) < address {
		return 0, fmt.Errorf("invalid write at screen address %08x", address)
	}

	return uint32(vga.screen[address]), nil
}

// ReadPAL reads from palette buffer
// It does not support unaligned access
func (vga *VGA) ReadPAL(address uint32) (uint32, error) {
	//fmt.Printf("ReadPAL at %08x\n", address)
	vga.RLock()
	defer vga.RUnlock()
	if address%4 != 0 {
		return 0, fmt.Errorf("unaligned access at palette read address %08x", address)
	}
	address /= 4

	if address >= 256 {
		return 0, fmt.Errorf("invalid read at palette address %08x", address*4)
	}

	v := uint32(vga.palette[address].B)
	v |= uint32(vga.palette[address].G) << 8
	v |= uint32(vga.palette[address].R) << 16
	v |= uint32(vga.palette[address].A) << 24

	return v, nil
}

// Map maps the palette and screen into the specified bus with specified base address
func (vga *VGA) Map(baseAddress uint32, bus *core.Bus) error {
	palRHandle := func(ctx context.Context, address uint32) (uint32, error) {
		return vga.ReadPAL(address - baseAddress - PaletteAddressOffset)
	}
	palWHandle := func(ctx context.Context, address, value uint32, writeMask byte) error {
		return vga.WritePAL(address-baseAddress-PaletteAddressOffset, value, writeMask)
	}

	err := bus.Map(PaletteMapName, baseAddress+PaletteAddressOffset, baseAddress+ScreenAddressOffset, palRHandle, palWHandle)
	if err != nil {
		return fmt.Errorf("cannot map vga palette: %s", err)
	}

	screenRHandle := func(ctx context.Context, address uint32) (uint32, error) {
		return vga.ReadScreen(address - baseAddress - ScreenAddressOffset)
	}
	screenWHandle := func(ctx context.Context, address, value uint32, writeMask byte) error {
		return vga.WriteScreen(address-baseAddress-ScreenAddressOffset, value, writeMask)
	}

	screenSize := uint32(len(vga.screen) * 4) // We align each byte inside a word

	err = bus.Map(ScreenMapName, baseAddress+ScreenAddressOffset, baseAddress+ScreenAddressOffset+screenSize, screenRHandle, screenWHandle)
	if err != nil {
		return fmt.Errorf("cannot map vga screen: %s", err)
	}

	controlRHandle := func(ctx context.Context, address uint32) (uint32, error) {
		return vga.ReadStatus(address)
	}

	err = bus.Map(ControlMapName, baseAddress+ControlAddressOffset, baseAddress+ControlAddressOffset+4, controlRHandle, nil)
	if err != nil {
		return fmt.Errorf("cannot map vga screen: %s", err)
	}

	return nil
}

// GetBuffer takes a color buffer as input and returns a color mapped buffer with the current screen contents
// If buffer argument is nil, or len(buffer) < screenPixels, it will return a new buffer
func (vga *VGA) GetBuffer(buffer []color.RGBA) []color.RGBA {
	vga.RLock()
	defer vga.RUnlock()

	pixels := vga.height * vga.width

	if buffer == nil || len(buffer) < pixels {
		buffer = make([]color.RGBA, pixels)
	}

	for i := 0; i < pixels; i++ {
		buffer[i] = vga.palette[vga.screen[i]]
	}

	return buffer
}
