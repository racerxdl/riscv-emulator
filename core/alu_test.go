package core

import (
	"math/rand"
	"testing"
	"time"
)

const numRounds = 128

func TestALU(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	rv32 := RISCV{}
	//aluADD                        = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluADD, X, Y) != X+Y {
			t.Errorf("failed aluADD for X: %d and Y: %d", X, Y)
		}
	}
	//aluSUB                        = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluSUB, X, Y) != X-Y {
			t.Errorf("failed aluSUB for X: %d and Y: %d", X, Y)
		}
	}
	//aluOR                         = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluOR, X, Y) != X|Y {
			t.Errorf("failed aluOR for X: %d and Y: %d", X, Y)
		}
	}
	//aluXOR                        = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluXOR, X, Y) != X^Y {
			t.Errorf("failed aluXOR for X: %d and Y: %d", X, Y)
		}
	}
	//aluAND                        = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluAND, X, Y) != X&Y {
			t.Errorf("failed aluAND for X: %d and Y: %d", X, Y)
		}
	}
	//aluLesserThanUnsigned         = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		r := uint32(0)
		if X < Y {
			r = 1
		}

		if rv32.alu(aluLesserThanUnsigned, X, Y) != r {
			t.Errorf("failed aluLesserThanUnsigned for X: %d and Y: %d", X, Y)
		}
	}
	//aluLesserThanSigned           = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		r := uint32(0)
		if int32(X) < int32(Y) {
			r = 1
		}

		if rv32.alu(aluLesserThanSigned, X, Y) != r {
			t.Errorf("failed aluLesserThanSigned for X: %d and Y: %d", X, Y)
		}
	}
	//aluShiftRightUnsigned         = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluShiftRightUnsigned, X, Y) != X>>Y {
			t.Errorf("failed aluShiftRightUnsigned for X: %d and Y: %d", X, Y)
		}
	}
	//aluShiftRightSigned           = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluShiftRightSigned, X, Y) != uint32(int32(X)>>int32(Y)) {
			t.Errorf("failed aluShiftRightSigned for X: %d and Y: %d", X, Y)
		}
	}
	//aluShiftLeftUnsigned          = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluShiftLeftUnsigned, X, Y) != X<<Y {
			t.Errorf("failed aluShiftLeftUnsigned for X: %d and Y: %d", X, Y)
		}
	}
	//aluShiftLeftSigned            = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		if rv32.alu(aluShiftLeftSigned, X, Y) != uint32(int32(X)<<int32(Y)) {
			t.Errorf("failed aluShiftLeftSigned for X: %d and Y: %d", X, Y)
		}
	}
	//aluGreaterThanOrEqualUnsigned = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		r := uint32(0)
		if X >= Y {
			r = 1
		}

		if rv32.alu(aluGreaterThanOrEqualUnsigned, X, Y) != r {
			t.Errorf("failed aluGreaterThanOrEqualUnsigned for X: %d and Y: %d", X, Y)
		}
	}
	//aluGreaterThanOrEqualSigned   = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		r := uint32(0)
		if int32(X) >= int32(Y) {
			r = 1
		}

		if rv32.alu(aluGreaterThanOrEqualSigned, X, Y) != r {
			t.Errorf("failed aluGreaterThanOrEqualSigned for X: %d and Y: %d", X, Y)
		}
	}
	//aluEqual                      = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		r := uint32(0)
		if X == Y {
			r = 1
		}

		if rv32.alu(aluEqual, X, Y) != r {
			t.Errorf("failed aluEqual for X: %d and Y: %d", X, Y)
		}
	}
	//aluNotEqual                   = iota
	for i := 0; i < numRounds; i++ {
		X := uint32(rand.Int31())
		Y := uint32(rand.Int31())

		r := uint32(0)
		if X != Y {
			r = 1
		}

		if rv32.alu(aluNotEqual, X, Y) != r {
			t.Errorf("failed aluEqual for X: %d and Y: %d", X, Y)
		}
	}
}

func TestSignExtend(t *testing.T) {
	for i := 0; i < 256; i++ { // Test 8 bit range
		got := signExtend(uint32(i), 8)
		expected := int8(i)
		if got != int32(expected) {
			t.Errorf("expected %d got %d for 8 bit extension", expected, got)
		}
	}
	for i := 0; i < 65535; i++ { // Test 16 bit range
		got := signExtend(uint32(i), 16)
		expected := int16(i)
		if got != int32(expected) {
			t.Errorf("expected %d got %d for 8 bit extension", expected, got)
		}
	}
}
