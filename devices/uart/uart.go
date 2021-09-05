package uart

import (
	"context"
	"fmt"
	"github.com/racerxdl/riscv-emulator/core"
	"sync"
)

type UART struct {
	sync.RWMutex
	inputBuffer  []byte
	outputBuffer []byte
}

func NewUART() *UART {
	return &UART{}
}

// PutC puts a character in UART input buffer
func (uart *UART) PutC(c byte) {
	uart.Lock()
	defer uart.Unlock()

	uart.inputBuffer = append(uart.inputBuffer, c)
}

func (uart *UART) ReadOutputBuffer() []byte {
	uart.Lock()
	defer uart.Unlock()
	c := uart.outputBuffer
	uart.outputBuffer = nil
	return c
}

// Write writes data to UART output buffer
func (uart *UART) Write(address uint32, value uint32, writeMask uint8) error {
	uart.Lock()
	defer uart.Unlock()
	if address == 0 {
		uart.outputBuffer = append(uart.outputBuffer, byte(value&0xFF))
	}
	return nil
}

// Read data from UART input buffer
func (uart *UART) Read(address uint32) (uint32, error) {
	uart.Lock()
	defer uart.Unlock()
	if len(uart.inputBuffer) > 0 && address == 0 {
		v := uart.inputBuffer[0]
		uart.inputBuffer = uart.inputBuffer[1:]
		return uint32(v), nil
	}
	return 0xFFFFFFFF, nil
}

// Map maps the memory into the specified bus with specified base address
func (uart *UART) Map(baseAddress uint32, bus *core.Bus) error {
	rhandle := func(ctx context.Context, address uint32) (uint32, error) {
		return uart.Read(address - baseAddress)
	}
	whandle := func(ctx context.Context, address, value uint32, writeMask byte) error {
		return uart.Write(address-baseAddress, value, writeMask)
	}

	err := bus.Map("uart", baseAddress, baseAddress+8, rhandle, whandle)
	if err != nil {
		return fmt.Errorf("(UART) cannot map uart: %s", err)
	}

	return nil
}
