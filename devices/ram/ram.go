package ram

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/racerxdl/riscv-emulator/core"
)

type RAM struct {
	ROM
}

func NewRAM(name string, size int) *RAM {
	if size%4 != 0 {
		// PAD to align
		size = size + (4 - (size % 4))
	}
	return &RAM{
		ROM: ROM{
			name: name,
			Data: make([]byte, size),
		},
	}
}

// Write writes data to ram
func (ram *RAM) Write(address uint32, value uint32, writeMask uint8) error {
	slice := ram.Data[address:]
	if len(slice) < 4 {
		return fmt.Errorf("(%s) not enough bytes to write at %08x", ram.name, address)
	}
	current, _ := ram.Read(address)
	switch writeMask {
	case 1: // Single byte
		current &= 0xFFFFFF00
		current |= value & 0xFF
	case 3: // Single Short
		current &= 0xFFFF0000
		current |= value & 0xFFFF
	case 15: // Full Word
		current = value
	default:
		return fmt.Errorf("(%s) invalid mask %04b on write at %08x", ram.name, writeMask, address)
	}
	binary.LittleEndian.PutUint32(slice, current)
	//fmt.Printf("Wrote %08x to %08x\n", current, address)
	return nil
}

// Map maps the memory into the specified bus with specified base address
func (ram *RAM) Map(baseAddress uint32, bus *core.Bus) error {
	rhandle := func(ctx context.Context, address uint32) (uint32, error) {
		return ram.Read(address - baseAddress)
	}
	whandle := func(ctx context.Context, address, value uint32, writeMask byte) error {
		return ram.Write(address-baseAddress, value, writeMask)
	}

	err := bus.Map(ram.name, baseAddress, baseAddress+uint32(len(ram.Data)), rhandle, whandle)
	if err != nil {
		return fmt.Errorf("(%s) cannot map ram: %s", ram.name, err)
	}

	return nil
}
