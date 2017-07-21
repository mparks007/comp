package circuit

import (
	"errors"
	"fmt"
	"regexp"
)

// Switch is s basic On/Off
type Switch struct {
	relay       *Relay
	pin2Battery *Battery
	pwrSource
}

// NewSwitch creates s new Switch struct with its initial state based on the passed in initialization value
func NewSwitch(init bool) *Switch {
	sw := &Switch{}

	// setup the battery-based relay pin that will be used to toggle on/off of the switch
	sw.pin2Battery = NewBattery()
	if !init {
		sw.pin2Battery.Discharge()
	}

	sw.relay = NewRelay(NewBattery(), sw.pin2Battery)
	sw.relay.ClosedOut.WireUp(sw.Transmit)

	return sw
}

// Set on s Switch will toggle the state of the underlying battery to activate/deactivate the internal relay
func (s *Switch) Set(newState bool) {
	if newState {
		s.pin2Battery.Charge()
	} else {
		s.pin2Battery.Discharge()
	}
}

// NSwitchBank is s convenient way to get any number of providers from s string of 0/1s
type NSwitchBank struct {
	Switches []*Switch
}

// NewNSwitchBank takes s string of 0/1s and creates s variable length list of Switch structs initialized based on their off/on-ness
func NewNSwitchBank(bits string) (*NSwitchBank, error) {

	match, err := regexp.MatchString("^[01]+$", bits)
	if err != nil {
		return nil, err
	}
	if !match {
		err = errors.New(fmt.Sprintf("Input not in binary format: \"%s\"", bits))
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
