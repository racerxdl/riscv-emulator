package core

import (
	"context"
	"fmt"
)

const insOpcodeMask = 0x00_00_00_7f // Bits 0 to 6
const insRdMask = 0x00_00_0F_80     // Bits 7 to 11
const insFunct3Mask = 0x00_00_70_00 // Bits 12 to 14
const insRs1Mask = 0x00_0F_80_00    // Bits 15 to 19
const insRs2Mask = 0x01_F0_00_00    // Bits 20 to 24
const insFunct7Mask = 0xFE_00_00_00 // Bits 25 to 31

const insImmTypeI = 0xFF_F0_00_00
const insImmTypeS0 = insRdMask
const insImmTypeS1 = insFunct7Mask
const insImmTypeB0 = 0x00_00_0F_00
const insImmTypeB1 = 0x7E_00_00_00
const insImmTypeB2 = 0x00_00_00_80
const insImmTypeB3 = 0x80_00_00_00
const insImmTypeU = 0xFF_FF_F0_00
const insImmTypeJ0 = 0x7F_E0_00_00
const insImmTypeJ1 = 0x00_10_00_00
const insImmTypeJ2 = 0x00_0F_F8_00
const insImmTypeJ3 = 0x80_00_00_00

func (rv32 *RISCV) runInstruction(ctx context.Context, ins uint32) error {
	// Splice the instruction
	opcode := ins & insOpcodeMask
	rd := (ins & insRdMask) >> 7
	funct3 := (ins & insFunct3Mask) >> 12
	rs1 := (ins & insRs1Mask) >> 15
	rs2 := (ins & insRs2Mask) >> 20
	funct7 := (ins & insFunct7Mask) >> 25

	immTypeI := (ins & insImmTypeI) >> 20
	immTypeS := ((ins & insImmTypeS0) >> 7) + ((ins & insImmTypeS1) >> 25)
	immTypeB := ((ins & insImmTypeB0) >> 7) + ((ins & insImmTypeB1) >> 20) + ((ins & insImmTypeB2) << 4) + ((ins & insImmTypeB3) >> 19)
	immTypeU := ins & insImmTypeU
	immTypeJ := ((ins & insImmTypeJ0) >> 20) + ((ins & insImmTypeJ1) >> 9) + (ins & insImmTypeJ2) + ((ins & insImmTypeJ3) >> 11)

	rs1Val := rv32.Registers.GetInteger(rs1)
	rs2Val := rv32.Registers.GetInteger(rs2)
	rdVal := rv32.Registers.GetInteger(rd)

	imm := uint32(0)

	// Normalize IMM Value to 32 bit
	switch opcode {
	case 0b0010011, 0b1100111, 0b0000011: // Type I
		if funct3 == 0b001 || funct3 == 0b101 {
			imm = immTypeI
		} else { // Sign Extend
			imm = uint32(signExtend(immTypeI, 12))
		}
	case 0b0100011: // Type S instructions
		imm = uint32(signExtend(immTypeS, 12))
	case 0b1100011: // Type B instructions
		imm = uint32(signExtend(immTypeB, 13))
	case 0b0010111, 0b0110111: // Type U instructions
		imm = immTypeU
	case 0b1101111: // Type J instructions
		imm = uint32(signExtend(immTypeJ, 20))
	}

	if opcode == 0b0010011 { // addi, slti, sltiu, xori, ori, andi, slli, srli, srai
		// imm[11:0]     rs1 000 rd 0010011 I addi
		// imm[11:0]     rs1 010 rd 0010011 I slti
		// imm[11:0]     rs1 011 rd 0010011 I sltiu
		// imm[11:0]     rs1 100 rd 0010011 I xori
		// imm[11:0]     rs1 110 rd 0010011 I ori
		// imm[11:0]     rs1 111 rd 0010011 I andi
		// 0000000 shamt rs1 001 rd 0010011 I slli
		// 0000000 shamt rs1 101 rd 0010011 I srli
		// 0100000 shamt rs1 101 rd 0010011 I srai
		aluOp := aluINVALID

		switch funct3 {
		case 0: // addi
			aluOp = aluADD
		case 1: // Shift Left Unsigned
			aluOp = aluShiftLeftUnsigned
		case 2: // LesserThanSigned;
			aluOp = aluLesserThanSigned
		case 3: // LesserThanUnsigned
			aluOp = aluLesserThanUnsigned
		case 4: // XOR;
			aluOp = aluXOR
		case 5: // funct7[5] ? ShiftRightSigned : ShiftRightUnsigned;
			aluOp = aluShiftRightUnsigned
			if funct7&0x20 > 0 {
				aluOp = aluShiftRightSigned
			}
			imm &= 0x1F
		case 6: // OR;
			aluOp = aluOR
		case 7: // AND;
			aluOp = aluAND
		}

		rdVal = rv32.alu(aluOp, rs1Val, imm)
		rv32.Registers.SetInteger(rd, rdVal)
		return nil
	}

	if opcode == 0b0110011 { // add, sub, sll, slt, sltu, xor, srl, sra, or, and
		//0000000 rs2 rs1 000 rd 0110011 R add
		//0100000 rs2 rs1 000 rd 0110011 R sub
		//0000000 rs2 rs1 001 rd 0110011 R sll
		//0000000 rs2 rs1 010 rd 0110011 R slt
		//0000000 rs2 rs1 011 rd 0110011 R sltu
		//0000000 rs2 rs1 100 rd 0110011 R xor
		//0000000 rs2 rs1 101 rd 0110011 R srl
		//0100000 rs2 rs1 101 rd 0110011 R sra
		//0000000 rs2 rs1 110 rd 0110011 R or
		//0000000 rs2 rs1 111 rd 0110011 R and

		aluOp := aluINVALID
		switch funct3 {
		case 0: // add / sub
			aluOp = aluADD
			if funct7&0x20 > 0 { // SUB
				aluOp = aluSUB
			}
		case 1: // Shift Left Unsigned
			aluOp = aluShiftLeftUnsigned
		case 2: // LesserThanSigned;
			aluOp = aluLesserThanSigned
		case 3: // LesserThanUnsigned
			aluOp = aluLesserThanUnsigned
		case 4: // XOR;
			aluOp = aluXOR
		case 5: // funct7[5] ? ShiftRightSigned : ShiftRightUnsigned;
			aluOp = aluShiftRightUnsigned
			if funct7&0x10 > 0 {
				aluOp = aluShiftRightSigned
			}
		case 6: // OR;
			aluOp = aluOR
		case 7: // AND;
			aluOp = aluAND
		}

		rdVal = rv32.alu(aluOp, rs1Val, rs2Val)
		rv32.Registers.SetInteger(rd, rdVal)
		return nil
	}

	if opcode == 0b1100011 { // beq, bne, blt, bge, bltu, bgeu
		aluOp := aluINVALID
		switch funct3 {
		case 0:
			aluOp = aluEqual
		case 1:
			aluOp = aluNotEqual
		case 2, 3:
			return fmt.Errorf("invalid instruction %08x", ins)
		case 4:
			aluOp = aluLesserThanSigned
		case 5:
			aluOp = aluGreaterThanOrEqualSigned
		case 6:
			aluOp = aluLesserThanUnsigned
		case 7:
			aluOp = aluGreaterThanOrEqualUnsigned
		}

		res := rv32.alu(aluOp, rs1Val, rs2Val)
		if res == 1 { // Branch
			rv32.AddPC(int32(imm) - 4)
		}
		return nil
	}

	if opcode == 0b0010111 { // auipc
		rdVal = rv32.alu(aluADD, rv32.GetPC()-4, imm)
		rv32.Registers.SetInteger(rd, rdVal)
		return nil
	}

	if opcode == 0b0110111 { // lui
		rv32.Registers.SetInteger(rd, imm)
		return nil
	}

	if opcode == 0b1101111 { // jal
		rv32.Registers.SetInteger(rd, rv32.GetPC())
		rv32.AddPC(int32(imm) - 4)
		return nil
	}

	if opcode == 0b1100111 { // jalr
		rv32.Registers.SetInteger(rd, rv32.GetPC()+4)
		newPC := rv32.alu(aluADD, rs1Val, imm) &^ 1
		rv32.SetPC(newPC)
		return nil
	}

	if opcode == 0b0000011 { // lb, lh, lw, lbu, lhu
		numBytes := funct3 & 3
		data := uint32(0)

		addr := rv32.alu(aluADD, rs1Val, imm)

		switch numBytes {
		case 0:
			b, err := rv32.Bus.ReadByte(ctx, addr)
			if err != nil {
				return err
			}
			data = uint32(b)
			if funct3&4 == 0 { // Sign Extend
				data = uint32(signExtend(data, 8))
			}
		case 1:
			b, err := rv32.Bus.ReadShort(ctx, addr)
			if err != nil {
				return err
			}
			data = uint32(b)
			if funct3&4 == 0 { // Sign Extend
				data = uint32(signExtend(data, 16))
			}
		case 2:
			b, err := rv32.Bus.ReadWord(ctx, addr)
			if err != nil {
				return err
			}
			data = b
		}

		rv32.Registers.SetInteger(rd, data)
		return nil
	}

	if opcode == 0b0100011 { // sw, sh, sb
		numBytes := funct3 & 3
		addr := rv32.alu(aluADD, rs1Val, imm)
		switch numBytes {
		case 0:
			return rv32.Bus.WriteByte(ctx, addr, byte(rs2Val&0xFF))
		case 1:
			return rv32.Bus.WriteShort(ctx, addr, uint16(rs2Val&0xFFFF))
		case 2:
			return rv32.Bus.WriteWord(ctx, addr, rs2Val)
		}
	}

	if opcode == 0b1110011 {
		if funct3 == 0 { // ecall / ebreak
			rv32.log.Info("ecall/ebreak not implemented")
			// TODO
			return nil
		}
		// CSR
		rv32.log.Info("CSR not implemented")
		return nil
	}

	return fmt.Errorf("invalid instruction %08x", ins)
}
