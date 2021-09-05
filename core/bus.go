package core

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
)

// Bus represents a Read/Write 32 bit address bus
type Bus struct {
	handlers map[string]BusMap
	log      *logrus.Logger
}

func CreateBus(log *logrus.Logger) *Bus {
	return &Bus{
		log:      log,
		handlers: make(map[string]BusMap),
	}
}

// Read performs a read in the bus
func (b *Bus) Read(ctx context.Context, address uint32) (uint32, error) {
	handler, err := b.getReadHandler(address)
	if err != nil {
		return 0, err
	}

	return handler(ctx, address)
}

// Write performs a write in the bus
func (b *Bus) Write(ctx context.Context, address, value uint32, writeMask byte) error {
	handler, err := b.getWriteHandler(address)
	if err != nil {
		return err
	}

	return handler(ctx, address, value, writeMask)
}

// getReadHandler finds a bus read handler for the specified address and returns it
func (b *Bus) getReadHandler(address uint32) (handle BusReadHandle, err error) {
	for _, v := range b.handlers {
		if v.In(address) {
			handle = v.RHandler
			if handle == nil {
				err = fmt.Errorf("no read handler for 0x%08x", address)
			}
			return
		}
	}

	return handle, fmt.Errorf("unmmaped space at 0x%08x", address)
}

// getWriteHandler finds a bus read handler for the specified address and returns it
func (b *Bus) getWriteHandler(address uint32) (handle BusWriteHandle, err error) {
	for _, v := range b.handlers {
		if v.In(address) {
			handle = v.WHandler
			if handle == nil {
				err = fmt.Errorf("no write handler for 0x%08x", address)
			}
			return
		}
	}

	return handle, fmt.Errorf("unmmaped space at 0x%08x", address)
}
