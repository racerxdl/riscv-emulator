package ram

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/racerxdl/riscv-emulator/core"
)

type ROM struct {
	name string
	Data []byte
}

func NewROM(name string, size int) *ROM {
	if size%4 != 0 {
		// PAD to align
		size = size + (4 - (size % 4))
	}
	return &ROM{
		name: name,
		Data: make([]byte, size),
	}
}

// Read reads data from ram
func (rom *ROM) Read(address uint32) (uint32, error) {
	if address > uint32(len(rom.Data)) {
		return 0, fmt.Errorf("(%s) invalid read at %08x", rom.name, address)
	}

	return binary.LittleEndian.Uint32(rom.Data[address:]), nil
}

// Map maps the memory into the specified bus with specified base address
func (rom *ROM) Map(baseAddress uint32, bus *core.Bus) error {
	rhandle := func(ctx context.Context, address uint32) (uint32, error) {
		return rom.Read(address - baseAddress)
	}

	err := bus.Map(rom.name, baseAddress, baseAddress+uint32(len(rom.Data)), rhandle, nil)
	if err != nil {
		return fmt.Errorf("(%s) cannot map rom: %s", rom.name, err)
	}

	return nil
}
