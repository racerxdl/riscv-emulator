package spi

import (
	"context"
	"fmt"
	"github.com/racerxdl/riscv-emulator/core"
	"github.com/sirupsen/logrus"
)

// Dummy SPI Controller
type SPI struct {
	log *logrus.Logger
}

// NewDummySPI creates a dummy SPI controller that does nothing but print out the commands sent
func NewDummySPI(log *logrus.Logger) *SPI {
	if log == nil {
		log = logrus.New()
	}
	return &SPI{log: log}
}

// Read reads data from SPI Registers
func (spi *SPI) Read(address uint32) (uint32, error) {
	switch address {
	case 0: // SPI_CSR
		spi.log.Info("Read SPI_CSR")
	case 0xC: // SPI_RF
		spi.log.Info("Write SPI_RF")
	case 0x40:
		spi.log.Info("Read SPI Mode")
	case 0x74:
		spi.log.Info("Read QSPI Parameters")
	default:
		spi.log.Infof("Read %08x", address)
	}
	return 0, nil
}

// Write writes data to SPI registers
func (spi *SPI) Write(address uint32, value uint32, writeMask uint8) error {
	switch address {
	case 0: // SPI_CSR
		spi.log.Infof("Write SPI_CSR = %08x", value)
	case 0xC: // SPI_RF
		spi.log.Infof("Write SPI_RF = %08x", value)
	case 0x40:
		spi.log.Infof("Set SPI Mode %08x", value)
	case 0x74:
		spi.log.Infof("Set QSPI Parameters %08x", value)
	default:
		spi.log.Infof("Write %08x %08x %02x", address, value, writeMask)
	}
	return nil
}

// Map maps the SPI Controller into the specified bus with specified base address
func (spi *SPI) Map(baseAddress uint32, bus *core.Bus) error {
	rhandle := func(ctx context.Context, address uint32) (uint32, error) {
		return spi.Read(address - baseAddress)
	}
	whandle := func(ctx context.Context, address, value uint32, writeMask byte) error {
		return spi.Write(address-baseAddress, value, writeMask)
	}

	err := bus.Map("dummySPI", baseAddress, baseAddress+256, rhandle, whandle)
	if err != nil {
		return fmt.Errorf("(DummySPI) cannot map dummy spi: %s", err)
	}

	return nil
}
