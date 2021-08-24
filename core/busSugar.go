package core

import (
	"context"
	"fmt"
)

// ReadByte performs a byte read in the bus
func (b *Bus) ReadByte(ctx context.Context, address uint32) (byte, error) {
	v, err := b.Read(ctx, address)
	return uint8(v & 0xFF), err
}

// ReadShort performs a uint16 read in the bus
func (b *Bus) ReadShort(ctx context.Context, address uint32) (uint16, error) {
	v, err := b.Read(ctx, address)
	return uint16(v & 0xFFFF), err
}

// ReadWord performs a uint32 read in the bus
func (b *Bus) ReadWord(ctx context.Context, address uint32) (uint32, error) {
	return b.Read(ctx, address)
}

// WriteByte performs a byte write in the bus
func (b *Bus) WriteByte(ctx context.Context, address uint32, value byte) error {
	return b.Write(ctx, address, uint32(value), 1)
}

// WriteShort performs a byte write in the bus
func (b *Bus) WriteShort(ctx context.Context, address uint32, value uint16) error {
	return b.Write(ctx, address, uint32(value), 3)
}

// WriteWord performs a uint32 write in the bus
func (b *Bus) WriteWord(ctx context.Context, address uint32, value uint32) error {
	return b.Write(ctx, address, value, 15)
}

const busMapHeadFormat = "%20s %08s %08s %02s\n"

// String returns all current maps in human readable format
func (b *Bus) String() string {
	result := fmt.Sprintf(busMapHeadFormat, "Name", "Start", "End", "RW")
	for _, m := range b.handlers {
		result += "\t" + m.String() + "\n"
	}
	return result
}
