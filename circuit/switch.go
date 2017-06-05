package circuit

import (
	"errors"
	"fmt"
	"regexp"
)

type Switch struct {
	bitPublication
}

func NewSwitch(init bool) *Switch {
	s := &Switch{}

	s.isPowered = init

	return s
}

type EightSwitchBank struct {
	Switches [8]*Switch
}

func NewEightSwitchBank(bits string) (*EightSwitchBank, error) {

	match, err := regexp.MatchString("^[01]{8}$", bits)
	if err != nil {
		return nil, err
	}
	if !match {
		err = errors.New(fmt.Sprint("Input not in 8-bit binary format: " + bits))
		return nil, err
	}

	sb := &EightSwitchBank{}

	for i, bit := range bits {
		sb.Switches[i] = NewSwitch(bit == '1')
	}

	return sb, nil
}

func (s *EightSwitchBank) AsBitPublishers() [8]bitPublisher {
	bitPubs := [8]bitPublisher{}

	for _, sw := range s.Switches {
		bitPubs[0] = sw
	}

	return bitPubs
}
