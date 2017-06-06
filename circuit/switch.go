package circuit

import (
	"errors"
	"fmt"
	"regexp"
)

type Switch struct {
	relay       *Relay2
	pin2Battery *Battery
	bitPublication
}

func NewSwitch(init bool) *Switch {
	s := &Switch{}

	s.pin2Battery = NewBattery()
	if !init {
		s.pin2Battery.Discharge()
	}

	s.relay = NewRelay2(NewBattery(), s.pin2Battery)
	s.isPowered = init

	s.relay.ClosedOut.Register(s.Publish)

	return s
}

func (s *Switch) Set(newState bool) {
	if newState {
		s.pin2Battery.Charge()
	} else {
		s.pin2Battery.Discharge()
	}
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

	for i, sw := range s.Switches {
		bitPubs[i] = sw
	}

	return bitPubs
}


type SixteenSwitchBank struct {
	Switches [16]*Switch
}

func NewSixteenSwitchBank(bits string) (*SixteenSwitchBank, error) {

	match, err := regexp.MatchString("^[01]{16}$", bits)
	if err != nil {
		return nil, err
	}
	if !match {
		err = errors.New(fmt.Sprint("Input not in 16-bit binary format: " + bits))
		return nil, err
	}

	sb := &SixteenSwitchBank{}

	for i, bit := range bits {
		sb.Switches[i] = NewSwitch(bit == '1')
	}

	return sb, nil
}

func (s *SixteenSwitchBank) AsBitPublishers() [16]bitPublisher {
	bitPubs := [16]bitPublisher{}

	for i, sw := range s.Switches {
		bitPubs[i] = sw
	}

	return bitPubs
}
