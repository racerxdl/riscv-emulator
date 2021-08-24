package core

import (
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"
)

func loadmem(file string) []byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var program []byte
	tmp := make([]byte, 4)

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		line = strings.Trim(line, " \r\n")
		if len(line) != 0 {
			v, err := strconv.ParseInt(line, 16, 64)
			if err != nil {
				panic(fmt.Errorf("cannot parse line %d: %q", i, line))
			}
			binary.LittleEndian.PutUint32(tmp, uint32(v))
			program = append(program, tmp...)
		}
	}

	if len(program)%4 != 0 {
		// Pad to 32 bit
		pad := make([]byte, 4-(len(program)%4))
		fmt.Printf("warning: padding input program by %d bytes\n", pad)
		program = append(program, pad...)
	}

	return program
}

func TestCPU_LoadStore(t *testing.T) {
	cpu := CreateEmulator(nil)

	program := loadmem("../testdata/test_loadstore.mem")
	padding := make([]byte, 64)
	program = append(program, padding...) // Allow unaligned access to go beyond read
	memory := make([]byte, 1024)

	readProgram := func(ctx context.Context, address uint32) (uint32, error) {
		var slice []byte
		if address >= 0x10000 {
			slice = memory[address-0x10000:]
		} else {
			slice = program[address:]
		}
		if len(slice) < 4 {
			return 0, fmt.Errorf("not enough bytes at %08x", address)
		}
		return binary.LittleEndian.Uint32(slice), nil
	}

	writeData := func(ctx context.Context, address, value uint32, writeMask byte) error {
		slice := memory[address-0x10000:]
		if len(slice) < 4 {
			return fmt.Errorf("not enough bytes to write at %08x", address)
		}
		current, _ := readProgram(ctx, address)
		switch writeMask {
		case 1: // Single byte
			current &= 0xFFFFFF00
			current |= value & 0xFF
		case 3: // Single Short
			current &= 0xFFFF0000
			current |= value & 0xFFFF
		case 15: // Full Word
			current = value
		}
		binary.LittleEndian.PutUint32(slice, current)

		return nil
	}

	ctx := context.Background()

	if err := cpu.Bus.Map("program", 0, uint32(len(program)), readProgram, writeData); err != nil {
		t.Fatal(err)
	}
	if err := cpu.Bus.Map("memory", 0x10000, 0x10000+1024, readProgram, writeData); err != nil {
		t.Fatal(err)
	}

	if err := cpu.RunUntil(ctx, 0x2C); err != nil { //  End of LOAD
		t.Fatalf("LOAD: %s", err)
	}

	if cpu.Registers.integers[8] != 0x000000DE {
		t.Fatalf("LOAD: Expected X%02d to be %08x but got %08x", 8, 0x000000DE, cpu.Registers.integers[8])
	}
	if cpu.Registers.integers[9] != 0x000000AD {
		t.Fatalf("LOAD: Expected X%02d to be %08x but got %08x", 9, 0x000000AD, cpu.Registers.integers[9])
	}
	if cpu.Registers.integers[10] != 0x000000BE {
		t.Fatalf("LOAD: Expected X%02d to be %08x but got %08x", 10, 0x000000BE, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0x000000EF {
		t.Fatalf("LOAD: Expected X%02d to be %08x but got %08x", 11, 0x000000EF, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0x0000DEAD {
		t.Fatalf("LOAD: Expected X%02d to be %08x but got %08x", 12, 0x0000DEAD, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0x0000ADBE {
		t.Fatalf("LOAD: Expected X%02d to be %08x but got %08x", 13, 0x0000ADBE, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0x0000BEEF {
		t.Fatalf("LOAD: Expected X%02d to be %08x but got %08x", 14, 0x0000BEEF, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0xDEADBEEF {
		t.Fatalf("LOAD: Expected X%02d to be %08x but got %08x", 15, 0xDEADBEEF, cpu.Registers.integers[15])
	}

	if err := cpu.RunUntil(ctx, 0x54); err != nil { //  End of LOAD with Sign Extension
		t.Fatalf("LOAD Sign Extended: %s", err)
	}

	if cpu.Registers.integers[9] != 0xFFFFFF84 {
		t.Fatalf("LOAD Sign Extended: Expected X%02d to be %08x but got %08x", 9, 0xFFFFFF84, cpu.Registers.integers[9])
	}
	if cpu.Registers.integers[10] != 0xFFFFFF83 {
		t.Fatalf("LOAD Sign Extended: Expected X%02d to be %08x but got %08x", 10, 0xFFFFFF83, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0xFFFFFF82 {
		t.Fatalf("LOAD Sign Extended: Expected X%02d to be %08x but got %08x", 11, 0xFFFFFF82, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0xFFFFFF81 {
		t.Fatalf("LOAD Sign Extended: Expected X%02d to be %08x but got %08x", 12, 0xFFFFFF81, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0xFFFF8483 {
		t.Fatalf("LOAD Sign Extended: Expected X%02d to be %08x but got %08x", 13, 0xFFFF8483, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0xFFFF8382 {
		t.Fatalf("LOAD Sign Extended: Expected X%02d to be %08x but got %08x", 14, 0xFFFF8382, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0xFFFF8281 {
		t.Fatalf("LOAD Sign Extended: Expected X%02d to be %08x but got %08x", 15, 0xFFFF8281, cpu.Registers.integers[15])
	}

	if err := cpu.RunUntil(ctx, 0x74); err != nil { //  End of Aligned Store
		t.Fatalf("Aligned Store: %s", err)
	}

	if binary.LittleEndian.Uint32(memory[0:]) != 0x81 {
		t.Fatalf("Aligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x10000, 0x81, binary.LittleEndian.Uint32(memory[0:]))
	}
	if binary.LittleEndian.Uint32(memory[4:]) != 0x8281 {
		t.Fatalf("Aligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x10004, 0x8281, binary.LittleEndian.Uint32(memory[4:]))
	}
	if binary.LittleEndian.Uint32(memory[8:]) != 0x84838281 {
		t.Fatalf("Aligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x10008, 0x84838281, binary.LittleEndian.Uint32(memory[8:]))
	}

	if err := cpu.RunUntil(ctx, 0xB8); err != nil { //  End of Unaligned Store
		t.Fatalf("Unaligned Store: %s", err)
	}
	if binary.LittleEndian.Uint32(memory[0x0:]) != 0x00000081 {
		t.Fatalf("Unaligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x10000, 0x00000081, binary.LittleEndian.Uint32(memory[0x0:]))
	}
	if binary.LittleEndian.Uint32(memory[0x4:]) != 0x00008100 {
		t.Fatalf("Unaligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x10004, 0x00008100, binary.LittleEndian.Uint32(memory[0x4:]))
	}
	if binary.LittleEndian.Uint32(memory[0x8:]) != 0x00810000 {
		t.Fatalf("Unaligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x10008, 0x00810000, binary.LittleEndian.Uint32(memory[0x8:]))
	}
	if binary.LittleEndian.Uint32(memory[0xC:]) != 0x81000000 {
		t.Fatalf("Unaligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x1000C, 0x81000000, binary.LittleEndian.Uint32(memory[0xC:]))
	}

	if binary.LittleEndian.Uint32(memory[0x10:]) != 0x00008281 {
		t.Fatalf("Unaligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x10004, 0x00008281, binary.LittleEndian.Uint32(memory[0x10:]))
	}
	if binary.LittleEndian.Uint32(memory[0x14:]) != 0x00828100 {
		t.Fatalf("Unaligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x10008, 0x00828100, binary.LittleEndian.Uint32(memory[0x14:]))
	}
	if binary.LittleEndian.Uint32(memory[0x18:]) != 0x82810000 {
		t.Fatalf("Unaligned Store: Expected Memory 0x%08x to be %08x but got %08x", 0x1000C, 0x82810000, binary.LittleEndian.Uint32(memory[0x18:]))
	}

}

func TestCPU_JALJALR(t *testing.T) {
	cpu := CreateEmulator(nil)

	program := loadmem("../testdata/test_jaljalr.mem")

	readProgram := func(ctx context.Context, address uint32) (uint32, error) {
		return binary.LittleEndian.Uint32(program[address:]), nil
	}

	ctx := context.Background()

	if err := cpu.Bus.Map("program", 0, uint32(len(program)), readProgram, nil); err != nil {
		t.Fatal(err)
	}

	if err := cpu.RunUntilWithTimeout(ctx, 0x1C, time.Second*2); err != nil { //  End of JAL/JALR
		t.Fatalf("JAL/JALR: %s", err)
	}

	if cpu.Registers.integers[1] != 0x18 {
		t.Errorf("JAL/JALR: Expected %02d to be %08x but got %08x", 1, 0x18, cpu.Registers.integers[1])
	}
	if cpu.Registers.integers[2] != 0x30 {
		t.Errorf("JAL/JALR: Expected %02d to be %08x but got %08x", 2, 0x30, cpu.Registers.integers[2])
	}
}

func TestCPU_LUIAUIPC(t *testing.T) {
	cpu := CreateEmulator(nil)

	program := loadmem("../testdata/test_luiauipc.mem")

	readProgram := func(ctx context.Context, address uint32) (uint32, error) {
		return binary.LittleEndian.Uint32(program[address:]), nil
	}

	ctx := context.Background()

	if err := cpu.Bus.Map("program", 0, uint32(len(program)), readProgram, nil); err != nil {
		t.Fatal(err)
	}

	if err := cpu.RunUntilWithTimeout(ctx, 0x24, time.Second*2); err != nil { // End of LUI/AUIPC
		t.Fatalf("LUI/AUIPC: %s", err)
	}

	if cpu.Registers.integers[1] != 0xFFFFF000 {
		t.Errorf("LUI/AUIPC: Expected %02d to be %08x but got %08x", 1, 0xFFFFF000, cpu.Registers.integers[1])
	}
	if cpu.Registers.integers[2] != 0xFFFFF018 {
		t.Errorf("LUI/AUIPC: Expected %02d to be %08x but got %08x", 2, 0xFFFFF018, cpu.Registers.integers[2])
	}
}

func TestCPU_JMPS(t *testing.T) {
	cpu := CreateEmulator(nil)

	program := loadmem("../testdata/test_jmps.mem")

	readProgram := func(ctx context.Context, address uint32) (uint32, error) {
		return binary.LittleEndian.Uint32(program[address:]), nil
	}

	ctx := context.Background()

	if err := cpu.Bus.Map("program", 0, uint32(len(program)), readProgram, nil); err != nil {
		t.Fatal(err)
	}

	// If jmps are wrong, the test should timeout
	if err := cpu.RunUntilWithTimeout(ctx, 0x34, time.Second*2); err != nil { // End of BEQ
		t.Fatalf("BEQ: %s", err)
	}

	if err := cpu.RunUntilWithTimeout(ctx, 0x4C, time.Second*2); err != nil { // End of BNE
		t.Fatalf("BNE: %s", err)
	}

	if err := cpu.RunUntilWithTimeout(ctx, 0x7C, time.Second*2); err != nil { // End of BLT
		t.Fatalf("BLT: %s", err)
	}

	if err := cpu.RunUntilWithTimeout(ctx, 0xB8, time.Second*2); err != nil { // End of BGE
		t.Fatalf("BGE: %s", err)
	}

	if err := cpu.RunUntilWithTimeout(ctx, 0xD0, time.Second*2); err != nil { // End of BLTU
		t.Fatalf("BLTU: %s", err)
	}

	if err := cpu.RunUntilWithTimeout(ctx, 0xE4, time.Second*2); err != nil { // End of BGEU
		t.Fatalf("BGEU: %s", err)
	}
}

func TestCPU_ALU(t *testing.T) {
	cpu := CreateEmulator(nil)

	program := loadmem("../testdata/test_alu.mem")

	readProgram := func(ctx context.Context, address uint32) (uint32, error) {
		return binary.LittleEndian.Uint32(program[address:]), nil
	}

	ctx := context.Background()

	if err := cpu.Bus.Map("program", 0, uint32(len(program)), readProgram, nil); err != nil {
		t.Fatal(err)
	}

	// Test ADDI
	if err := cpu.RunUntil(ctx, 0x3C); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[0] != 0x000 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 0, 0x000, cpu.Registers.integers[0])
	}
	if cpu.Registers.integers[1] != 0x3E8 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 1, 0x3E8, cpu.Registers.integers[1])
	}
	if cpu.Registers.integers[2] != 0xBB8 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 2, 0xBB8, cpu.Registers.integers[2])
	}
	if cpu.Registers.integers[3] != 0x7D0 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 3, 0x7D0, cpu.Registers.integers[3])
	}
	if cpu.Registers.integers[4] != 0x000 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 4, 0x000, cpu.Registers.integers[4])
	}
	if cpu.Registers.integers[5] != 0x3E8 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 5, 0x3E8, cpu.Registers.integers[5])
	}
	if cpu.Registers.integers[6] != 0xBB8 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 6, 0xBB8, cpu.Registers.integers[6])
	}
	if cpu.Registers.integers[7] != 0x7D0 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 7, 0x7D0, cpu.Registers.integers[7])
	}
	if cpu.Registers.integers[8] != 0x000 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 8, 0x000, cpu.Registers.integers[8])
	}
	if cpu.Registers.integers[9] != 0x3E8 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 9, 0x3E8, cpu.Registers.integers[9])
	}
	if cpu.Registers.integers[10] != 0xBB8 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 10, 0xBB8, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0x7D0 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 11, 0x7D0, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0x000 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 12, 0x000, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0x3E8 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 13, 0x3E8, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0xBB8 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 14, 0xBB8, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0x7D0 {
		t.Errorf("ADDI: Expected %02d to be %08x but got %08x", 15, 0x7D0, cpu.Registers.integers[15])
	}

	// Test SLTI/SLTIU
	if err := cpu.RunUntil(ctx, 0x58); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[10] != 0x0 {
		t.Errorf("SLTI/SLTIU: Expected X%02d to be %08x but got %08x", 10, 0x0, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0x1 {
		t.Errorf("SLTI/SLTIU: Expected X%02d to be %08x but got %08x", 11, 0x1, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0x1 {
		t.Errorf("SLTI/SLTIU: Expected X%02d to be %08x but got %08x", 12, 0x1, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0x0 {
		t.Errorf("SLTI/SLTIU: Expected X%02d to be %08x but got %08x", 13, 0x0, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0x0 {
		t.Errorf("SLTI/SLTIU: Expected X%02d to be %08x but got %08x", 14, 0x0, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0x1 {
		t.Errorf("SLTI/SLTIU: Expected X%02d to be %08x but got %08x", 15, 0x1, cpu.Registers.integers[15])
	}

	// Test XORI
	if err := cpu.RunUntil(ctx, 0x70); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[13] != 0xFFFFFF0F {
		t.Errorf("XORI: Expected X%02d to be %08x but got %08x", 13, 0xFFFFFF0F, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0xFFFFF8F0 {
		t.Errorf("XORI: Expected X%02d to be %08x but got %08x", 14, 0xFFFFF8F0, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0xFFFFF800 {
		t.Errorf("XORI: Expected X%02d to be %08x but got %08x", 15, 0xFFFFF800, cpu.Registers.integers[15])
	}

	// Test ORI
	if err := cpu.RunUntil(ctx, 0x94); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[10] != 0x000000F0 {
		t.Errorf("ORI: Expected X%02d to be %08x but got %08x", 10, 0x000000F0, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0x0000070F {
		t.Errorf("ORI: Expected X%02d to be %08x but got %08x", 11, 0x0000070F, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0x000007FF {
		t.Errorf("ORI: Expected X%02d to be %08x but got %08x", 12, 0x000007FF, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0xFFFFFFFF {
		t.Errorf("ORI: Expected X%02d to be %08x but got %08x", 13, 0xFFFFFFFF, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0xFFFFFFFF {
		t.Errorf("ORI: Expected X%02d to be %08x but got %08x", 14, 0xFFFFFFFF, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0xFFFFFFFF {
		t.Errorf("ORI: Expected X%02d to be %08x but got %08x", 15, 0xFFFFFFFF, cpu.Registers.integers[15])
	}

	// Test ANDI
	if err := cpu.RunUntil(ctx, 0xB8); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[10] != 0x00000050 {
		t.Errorf("ANDI: Expected X%02d to be %08x but got %08x", 10, 0x00000050, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0x00000105 {
		t.Errorf("ANDI: Expected X%02d to be %08x but got %08x", 11, 0x00000105, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0x00000155 {
		t.Errorf("ANDI: Expected X%02d to be %08x but got %08x", 12, 0x00000155, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0x000000F0 {
		t.Errorf("ANDI: Expected X%02d to be %08x but got %08x", 13, 0x000000F0, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0x0000070F {
		t.Errorf("ANDI: Expected X%02d to be %08x but got %08x", 14, 0x0000070F, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0x000007FF {
		t.Errorf("ANDI: Expected X%02d to be %08x but got %08x", 15, 0x000007FF, cpu.Registers.integers[15])
	}

	// Testing SLLI
	if err := cpu.RunUntil(ctx, 0xD8); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[9] != 0xFF000000 {
		t.Errorf("SLLI: Expected X%02d to be %08x but got %08x", 9, 0x00000050, cpu.Registers.integers[9])
	}
	if cpu.Registers.integers[10] != 0x07FF0000 {
		t.Errorf("SLLI: Expected X%02d to be %08x but got %08x", 10, 0x00000050, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0x0007FF00 {
		t.Errorf("SLLI: Expected X%02d to be %08x but got %08x", 11, 0x00000105, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0x00003FF8 {
		t.Errorf("SLLI: Expected X%02d to be %08x but got %08x", 12, 0x00000155, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0x00001FFC {
		t.Errorf("SLLI: Expected X%02d to be %08x but got %08x", 13, 0x000000F0, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0x00000FFE {
		t.Errorf("SLLI: Expected X%02d to be %08x but got %08x", 14, 0x0000070F, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0x000007FF {
		t.Errorf("SLLI: Expected X%02d to be %08x but got %08x", 15, 0x000007FF, cpu.Registers.integers[15])
	}

	// Testing SRLI
	if err := cpu.RunUntil(ctx, 0xFC); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[9] != 0x000000FF {
		t.Errorf("SRLI: Expected X%02d to be %08x but got %08x", 9, 0x00000050, cpu.Registers.integers[9])
	}
	if cpu.Registers.integers[10] != 0x0000FF00 {
		t.Errorf("SRLI: Expected X%02d to be %08x but got %08x", 10, 0x00000050, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0x00FF0000 {
		t.Errorf("SRLI: Expected X%02d to be %08x but got %08x", 11, 0x00000105, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0x1FE00000 {
		t.Errorf("SRLI: Expected X%02d to be %08x but got %08x", 12, 0x00000155, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0x3FC00000 {
		t.Errorf("SRLI: Expected X%02d to be %08x but got %08x", 13, 0x000000F0, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0x7F800000 {
		t.Errorf("SRLI: Expected X%02d to be %08x but got %08x", 14, 0x0000070F, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0xFF000000 {
		t.Errorf("SRLI: Expected X%02d to be %08x but got %08x", 15, 0x000007FF, cpu.Registers.integers[15])
	}

	// Testing SRAI
	if err := cpu.RunUntil(ctx, 0x11C); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[9] != 0xFFFFFFFF {
		t.Errorf("SRAI: Expected X%02d to be %08x but got %08x", 9, 0xFFFFFFFF, cpu.Registers.integers[9])
	}
	if cpu.Registers.integers[10] != 0xFFFFFF00 {
		t.Errorf("SRAI: Expected X%02d to be %08x but got %08x", 10, 0xFFFFFF00, cpu.Registers.integers[10])
	}
	if cpu.Registers.integers[11] != 0xFFFF0000 {
		t.Errorf("SRAI: Expected X%02d to be %08x but got %08x", 11, 0xFFFF0000, cpu.Registers.integers[11])
	}
	if cpu.Registers.integers[12] != 0xFFE00000 {
		t.Errorf("SRAI: Expected X%02d to be %08x but got %08x", 12, 0xFFE00000, cpu.Registers.integers[12])
	}
	if cpu.Registers.integers[13] != 0xFFC00000 {
		t.Errorf("SRAI: Expected X%02d to be %08x but got %08x", 13, 0xFFC00000, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0xFF800000 {
		t.Errorf("SRAI: Expected X%02d to be %08x but got %08x", 14, 0xFF800000, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0xFF000000 {
		t.Errorf("SRAI: Expected X%02d to be %08x but got %08x", 15, 0xFF000000, cpu.Registers.integers[15])
	}

	// Testing ADD
	if err := cpu.RunUntil(ctx, 0x12C); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[13] != 0xFF000000 {
		t.Errorf("ADD: Expected X%02d to be %08x but got %08x", 13, 0xFF000000, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0xFEFFFFFF {
		t.Errorf("ADD: Expected X%02d to be %08x but got %08x", 14, 0xFEFFFFFF, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0xFE000000 {
		t.Errorf("ADD: Expected X%02d to be %08x but got %08x", 15, 0xFE000000, cpu.Registers.integers[15])
	}

	// Testing SUB
	if err := cpu.RunUntil(ctx, 0x13C); err != nil {
		t.Fatal(err)
	}

	if cpu.Registers.integers[13] != 0x01000000 {
		t.Errorf("SUB: Expected X%02d to be %08x but got %08x", 13, 0x01000000, cpu.Registers.integers[13])
	}
	if cpu.Registers.integers[14] != 0xFF000001 {
		t.Errorf("SUB: Expected X%02d to be %08x but got %08x", 14, 0xFF000001, cpu.Registers.integers[14])
	}
	if cpu.Registers.integers[15] != 0x00000000 {
		t.Errorf("SUB: Expected X%02d to be %08x but got %08x", 15, 0x00000000, cpu.Registers.integers[15])
	}

}
