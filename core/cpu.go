package core

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type RISCV struct {
	log *logrus.Logger

	Registers *RegisterBank
	Bus       *Bus

	pc       uint32
	cycleNum uint32
}

func CreateEmulator(log *logrus.Logger) *RISCV {
	if log == nil {
		log = logrus.New()
	}
	return &RISCV{
		log:       log,
		Registers: CreateRegisterBank(log),
		Bus:       CreateBus(log),
	}
}

func (rv32 *RISCV) Reset() {
	rv32.Registers.Reset()
}

// SetPC sets the program counter
func (rv32 *RISCV) SetPC(pc uint32) {
	rv32.log.Debugf("Entrypoint set to 0x%08x", pc)
	rv32.pc = pc
}

// GetPC gets the program counter
func (rv32 *RISCV) GetPC() uint32 {
	return rv32.pc
}

// AddPC adds the value offset to PC
func (rv32 *RISCV) AddPC(value int32) {
	rv32.pc = uint32(int32(rv32.pc) + value)
}

// Step runs a single instruction
func (rv32 *RISCV) Step(ctx context.Context) error {
	rv32.cycleNum++
	value, err := rv32.Bus.Read(ctx, rv32.pc)
	if err != nil {
		rv32.log.Errorf("error reading program at %08x: %s", rv32.pc, err)
		return err
	}
	rv32.pc += 4
	return rv32.runInstruction(ctx, value)
}

// RunUntil runs the emulation until the specified code address is reached or timeout
func (rv32 *RISCV) RunUntilWithTimeout(ctx context.Context, address uint32, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	deadline, _ := ctx.Deadline()

	instructions := 0
	for rv32.GetPC() != address {
		err := rv32.Step(ctx)
		if err != nil {
			return err
		}
		if instructions%16 == 0 && time.Now().After(deadline) {
			return fmt.Errorf("timeout at PC = %08x", rv32.GetPC())
		}
		instructions++
	}

	return rv32.RunUntil(ctx, address)
}

// RunUntil runs the emulation until the specified code address is reached
func (rv32 *RISCV) RunUntil(ctx context.Context, address uint32) error {
	for rv32.GetPC() != address {
		err := rv32.Step(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
