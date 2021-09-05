package disasm

import "github.com/deadsy/rvda"

var dism *rvda.ISA

func init() {
	isa, err := rvda.New(32, rvda.RV32gc)
	if err != nil {
		panic(err)
	}
	dism = isa
}

// Disasm disassembles the specified instruction in the specified address
func Disasm(addr, ins uint32) string {
	return dism.Disassemble(uint(addr), uint(ins)).String()
}
