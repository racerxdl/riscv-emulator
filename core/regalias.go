package core

import "fmt"

var intRegAlias = map[int]string{
	0: "zero",
	1: "ra",
	2: "sp",
	3: "gp",
	4: "tp",
	5: "t0",
	6: "t1",
	7: "t2",
	8: "s0",
	9: "s1",
}

func init() {
	for i := 10; i < 18; i++ {
		intRegAlias[i] = fmt.Sprintf("a%d", i-10)
	}
	for i := 18; i < 28; i++ {
		intRegAlias[i] = fmt.Sprintf("s%d", i-16)
	}
	for i := 28; i < 32; i++ {
		intRegAlias[i] = fmt.Sprintf("t%d", i-25)
	}
}

// GetIntRegisterName returns the conventional name for the specified register (zero, ra, sp, etc...)
func GetIntRegisterName(reg int) string {
	return intRegAlias[reg]
}
