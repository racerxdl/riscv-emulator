package core

import (
	"context"
	"fmt"
)

// BusWriteHandle is a handler for bus writes
// writeMask lower nibble specifies which bytes will be changed
// writeMask == 1 (0x000000FF)
// writeMask == 2 (0x0000FF00)
// writeMask == 3 (0x0000FFFF)
// writeMask == 4 (0x00FF0000)
// writeMask == 5 (0x00FF00FF)
// (...)
type BusWriteHandle func(ctx context.Context, address, value uint32, writeMask byte) error

// BusReadHandle is a handler for bus reads
type BusReadHandle func(ctx context.Context, address uint32) (uint32, error)

// BusMap represents a mapping range of the bus
type BusMap struct {
	// Name is the name of the mapping
	Name string
	// Start of the bus map (inclusive)
	Start uint32
	// End of the bus map (exclusive)
	End uint32
	// RHandler is the read handler (can be nil if no read permission)
	RHandler BusReadHandle
	// WHandler is the write handler (can be nil if no write permission)
	WHandler BusWriteHandle
}

const busMapLineFormat = "%20s %08x %08x %2s"

// String returns the map specification
func (b BusMap) String() string {
	rw := ""
	if b.RHandler != nil {
		rw += "R"
	} else {
		rw += "-"
	}
	if b.WHandler != nil {
		rw += "W"
	} else {
		rw += "-"
	}
	return fmt.Sprintf(busMapLineFormat, b.Name, b.Start, b.End-1, rw)
}

// In returns true in case of the specified address to be inside that map
func (b BusMap) In(address uint32) bool {
	return address >= b.Start && address < b.End
}

// OverlapsWith returns true in case the specified range overlaps with the current map
func (b BusMap) OverlapsWith(startAddress, endAddress uint32) bool {
	return startAddress < b.End && endAddress > b.Start
}

// Map tries to map a space handler
func (b *Bus) Map(name string, startAddress, endAddress uint32, rhandler BusReadHandle, whandler BusWriteHandle) error {
	for _, m := range b.handlers {
		if m.OverlapsWith(startAddress, endAddress) {
			return fmt.Errorf("read range %08x-%08x is already mapped to %q", startAddress, endAddress, m.Name)
		}
	}
	b.handlers[name] = BusMap{
		Name:     name,
		Start:    startAddress,
		End:      endAddress,
		RHandler: rhandler,
		WHandler: whandler,
	}

	return nil
}

// UnmapRead removes a bus read mapping with the specified name
func (b *Bus) Unmap(name string) {
	delete(b.handlers, name)
}
