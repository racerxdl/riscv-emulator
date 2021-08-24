package core

const (
	aluINVALID                    = -1
	aluADD                        = iota
	aluSUB                        = iota
	aluOR                         = iota
	aluXOR                        = iota
	aluAND                        = iota
	aluLesserThanUnsigned         = iota
	aluLesserThanSigned           = iota
	aluShiftRightUnsigned         = iota
	aluShiftRightSigned           = iota
	aluShiftLeftUnsigned          = iota
	aluShiftLeftSigned            = iota
	aluGreaterThanOrEqualUnsigned = iota
	aluGreaterThanOrEqualSigned   = iota
	aluEqual                      = iota
	aluNotEqual                   = iota
)

// alu mimics the hardware ALU operations
func (rv32 *RISCV) alu(aluOp int, X, Y uint32) uint32 {
	switch aluOp {
	case aluADD:
		return X + Y
	case aluSUB:
		return X - Y
	case aluOR:
		return X | Y
	case aluXOR:
		return X ^ Y
	case aluAND:
		return X & Y
	case aluLesserThanUnsigned:
		if X < Y {
			return 1
		}
		return 0
	case aluLesserThanSigned:
		if int32(X) < int32(Y) {
			return 1
		}
		return 0
	case aluShiftRightUnsigned:
		return X >> Y
	case aluShiftRightSigned:
		return uint32(int32(X) >> Y)
	case aluShiftLeftUnsigned:
		return X << Y
	case aluShiftLeftSigned:
		return uint32(int32(X) << Y)
	case aluGreaterThanOrEqualUnsigned:
		if X >= Y {
			return 1
		}
		return 0
	case aluGreaterThanOrEqualSigned:
		if int32(X) >= int32(Y) {
			return 1
		}
		return 0
	case aluEqual:
		if X == Y {
			return 1
		}
		return 0
	case aluNotEqual:
		if X != Y {
			return 1
		}
		return 0
	}

	rv32.log.Errorf("invalid ALU operation %d", aluOp)
	return 0
}

// signExtend assumes value to be bits length and sign extends to 32 bit
func signExtend(value, bits uint32) int32 {
	bits -= 1
	max := uint32(1) << bits
	sign := (value & max) > 0 // Get the sign
	if sign {
		comp := 0xFF_FF_FF_FF &^ (max - 1) // generate the complement
		return int32(value | comp)
	}
	return int32(value)
}
