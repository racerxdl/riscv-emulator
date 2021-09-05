package core

import "github.com/sirupsen/logrus"

type RegisterBank struct {
	integers [32]uint32
	float    [32]float32

	log *logrus.Logger
}

func CreateRegisterBank(log *logrus.Logger) *RegisterBank {
	if log == nil {
		log = logrus.New()
	}
	return &RegisterBank{
		log: log,
	}
}

func (rb *RegisterBank) Reset() {
	for i := 0; i < 32; i++ {
		rb.integers[i] = 0
		rb.float[i] = 0
	}
}

// SetInteger sets a integer register to the specified value
func (rb *RegisterBank) SetInteger(registerNum, value uint32) {
	if registerNum > 31 {
		rb.log.Errorf("registerNum == %d and it is > 31", registerNum)
		return
	}
	if registerNum > 0 {
		rb.integers[registerNum] = value
	}
}

// SetFloat sets a float register to the specified value
func (rb *RegisterBank) SetFloat(registerNum uint32, value float32) {
	if registerNum > 31 {
		rb.log.Errorf("registerNum == %d and it is > 31", registerNum)
		return
	}
	rb.float[registerNum] = value
}

// GetInteger gets the value of a integer register
func (rb *RegisterBank) GetInteger(registerNum uint32) uint32 {
	if registerNum > 31 {
		rb.log.Errorf("registerNum == %d and it is > 31", registerNum)
		return 0
	}
	if registerNum == 0 {
		return 0
	}
	return rb.integers[registerNum]
}

// GetFloat gets the value of a float register
func (rb *RegisterBank) GetFloat(registerNum uint32) float32 {
	if registerNum > 31 {
		rb.log.Errorf("registerNum == %d and it is > 31", registerNum)
		return 0
	}
	return rb.float[registerNum]
}
