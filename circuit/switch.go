package circuit

import (
	"errors"
	"fmt"
	"regexp"
)

// Switch is a basic On/Off
type Switch struct {
	relay       *Relay
	pin2Battery *Battery
	pwrSource
}

// NewSwitch creates a new Switch struct with its initial state based on the passed in initialization value
func NewSwitch(init bool) *Switch {
	s := &Switch{}

	// setup the battery-based relay pin that will be used to toggle on/off of the switch
	s.pin2Battery = NewBattery()
	if !init {
		s.pin2Battery.Discharge()
	}

	s.relay = NewRelay(NewBattery(), s.pin2Battery)
	s.relay.ClosedOut.WireUp(s.Transmit)

	return s
}

// Set on a Switch will toggle the state of the underlying battery to activate/deactivate the internal relay
func (s *Switch) Set(newState bool) {
	if newState {
		s.pin2Battery.Charge()
	} else {
		s.pin2Battery.Discharge()
	}
}

// EightSwitchBank is a convenient way to get 8 bit providers from a string of 8 0/1s
type EightSwitchBank struct {
	Switches [8]*Switch
}

// NewEightSwitchBank takes a string of 0/1s and creates 8 Switch structs initialized based on their off/on-ness
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

// AsPwrEmitters will return pwrEmitter versions of the internal Switches
func (s *EightSwitchBank) AsPwrEmitters() [8]pwrEmitter {
	pwrEmits := [8]pwrEmitter{}

	for i, sw := range s.Switches {
		pwrEmits[i] = sw
	}

	return pwrEmits
}

// SixteenSwitchBank is a convenient way to get 16 bit providers from a string of 16 0/1s
type SixteenSwitchBank struct {
	Switches [16]*Switch
}

// NewEightSwitchBank takes a string of 0/1s and creates 16 Switch structs initialized based on their off/on-ness
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

// AsPwrEmitters will return pwrEmitter versions of the internal Switches
func (s *SixteenSwitchBank) AsPwrEmitters() [16]pwrEmitter {
	pwrEmits := [16]pwrEmitter{}

	for i, sw := range s.Switches {
		pwrEmits[i] = sw
	}

	return pwrEmits
}

// NSwitchBank is a convenient way to get any number of providers from a string of 0/1s
type NSwitchBank struct {
	Switches []*Switch
}

// NewNSwitchBank takes a string of 0/1s and creates a variable length list of Switch structs initialized based on their off/on-ness
func NewNSwitchBank(bits string) (*NSwitchBank, error) {

	match, err := regexp.MatchString("^[01]+$", bits)
	if err != nil {
		return nil, err
	}
	if !match {
		err = errors.New(fmt.Sprint("Input not in binary format: " + bits))
		return nil, err
	}

	sb := &NSwitchBank{}

	for _, bit := range bits {
		sb.Switches = append(sb.Switches, NewSwitch(bit == '1'))
	}

	return sb, nil
}

// AsPwrEmitters will return pwrEmitter versions of the internal Switches
func (s *NSwitchBank) AsPwrEmitters() []pwrEmitter {
	pwrEmits := []pwrEmitter{}

	for _, sw := range s.Switches {
		pwrEmits = append(pwrEmits, sw)
	}

	return pwrEmits
}
