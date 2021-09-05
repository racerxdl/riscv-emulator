package core

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"time"
)

type RISCV struct {
	log *logrus.Logger

	Registers *RegisterBank
	Bus       *Bus

	pc          uint32
	cycleNum    uint32
	running     bool
	step        bool
	started     bool
	breakpoints map[uint32]struct{}
}

func CreateEmulator(log *logrus.Logger) *RISCV {
	if log == nil {
		log = logrus.New()
	}
	return &RISCV{
		log:         log,
		Registers:   CreateRegisterBank(log),
		Bus:         CreateBus(log),
		breakpoints: make(map[uint32]struct{}),
	}
}

// Reset resets all registers and set the PC to 0
func (rv32 *RISCV) Reset() {
	rv32.log.Infof("CPU Reset")
	rv32.Registers.Reset()
	rv32.SetPC(0)
}

// AddBreak adds a breakpoint in the specified address
// A breakpoint will pause the CPU when is running by Start
func (rv32 *RISCV) AddBreak(addr uint32) {
	rv32.breakpoints[addr] = struct{}{}
}

// DelBreak deletes a breakpoint in the specified address
func (rv32 *RISCV) DelBreak(addr uint32) {
	delete(rv32.breakpoints, addr)
}

// SetPC sets the program counter
func (rv32 *RISCV) SetPC(pc uint32) {
	//rv32.log.Debugf("Entrypoint set to 0x%08x", pc)
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

// RunStep runs a single instruction
func (rv32 *RISCV) RunStep(ctx context.Context) error {
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
		err := rv32.RunStep(ctx)
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
		err := rv32.RunStep(ctx)
		if err != nil {
			return err
		}
		if _, ok := rv32.breakpoints[rv32.pc]; ok {
			return fmt.Errorf("breakpoint reached at %08x", rv32.pc)
		}
	}

	return nil
}

// Start starts a goroutine with the RISC-V emulation
// This starts the CPU as paused, so either Step or Continue should be run afterwards
func (rv32 *RISCV) Start() {
	if !rv32.started {
		rv32.started = true
		go rv32.Loop()
	}
}

// Stops stops the goroutine with the RISC-V emulation
func (rv32 *RISCV) Stop() {
	rv32.started = false
}

// Step makes the RISC-V Goroutine to step a single instruction
// This does nothing in standalone, and RunStep should be used when no gouroutine has been started
func (rv32 *RISCV) Step() {
	rv32.step = true
	rv32.running = true
}

// Paused returns if the core is currently paused
func (rv32 *RISCV) Paused() bool {
	return !rv32.running
}

// Loop is the loop used by Start
// Use manually with care
func (rv32 *RISCV) Loop() {
	ctx := context.Background()
	rv32.Reset()

	for rv32.started {
		if rv32.running {
			err := rv32.RunStep(ctx)
			//log.Infof("%s", Addr2Line(addr))
			if err != nil {
				rv32.log.Debugf("(RISCV) Error: %s", err)
				rv32.running = false
			}

			if rv32.step {
				rv32.log.Infof("Paused at %08x", rv32.pc)
				rv32.running = false
				rv32.step = false
			}

			if _, ok := rv32.breakpoints[rv32.pc]; ok {
				rv32.log.Infof("Breakpoint reached at %08x", rv32.pc)
				rv32.running = false
			}
		} else {
			time.Sleep(time.Millisecond)
		}
		runtime.Gosched()
	}
}

// Pause pauses the RISC-V emulation goroutine
func (rv32 *RISCV) Pause() {
	rv32.running = false
}

// Continue resumes the RISC-V emulation goroutine
func (rv32 *RISCV) Continue() {
	rv32.running = true
}
