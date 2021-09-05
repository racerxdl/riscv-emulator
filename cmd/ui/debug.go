package main

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
)

// ok thats WAYYYY of a workarround
func Addr2Line(addr uint32) string {
	cmd := exec.Command("/home/lucas/.local/xPacks/@xpack-dev-tools/riscv-none-embed-gcc/10.1.0-1.1.1/.content/bin//riscv-none-embed-addr2line", "-f", "-e", "/media/lucas/ELTNEXT/Works2/doom_riscv/src/riscv/doom-riscv.elf", fmt.Sprintf("0x%08x", addr))

	out, _ := cmd.Output()
	lines := strings.Split(string(out), "\n")
	filename_line := strings.Split(lines[1], ":")
	funcname := lines[0]
	filename := path.Base(filename_line[0])
	return fmt.Sprintf("%s:%s (%s)", filename, filename_line[1], funcname)
}
