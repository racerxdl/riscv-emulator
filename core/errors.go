package core

// RuntimeError represents an runtime exception at RISC-V Emulation
type RuntimeError struct {
	Message       string
	RegisterState RegisterBank
	PC            uint32
}

func (re RuntimeError) String() string {
	return re.Message
}

func (re RuntimeError) Error() string {
	return re.String()
}
