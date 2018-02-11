package circuit

import (
	"fmt"
	"regexp"
)

// Switch is a basic On/Off
type Switch struct {
	relay       *Relay
	pin2Battery *Battery
	ch          chan bool
	pwrSource
}

// NewSwitch creates a new Switch struct with its initial state based on the passed in initialization value
func NewSwitch(init bool) *Switch {
	sw := &Switch{}
	sw.ch = make(chan bool, 1)

	// setup the battery-based relay pin that will be used to toggle on/off of the switch (see Set(bool) method)
	sw.pin2Battery = NewBattery()
	if !init {
		sw.pin2Battery.Discharge()
	}
	sw.relay = NewRelay(NewBattery(), sw.pin2Battery)

	// a switch acts like a relay, where Closed out is its power "answer"
	sw.relay.ClosedOut.WireUp(sw.ch)

	transmit := func() {
		sw.Transmit(<-sw.ch)
	}

	// calling transmit explicitly to ensure the 'answer' for the switch output, post WireUp above, has settled BEFORE returning and letting things wire up to it
	transmit()

	go func() {
		for {
			transmit()
		}
	}()

	return sw
}

// Set method on a Switch will toggle the state of the underlying battery to activate/deactivate the internal relay
func (s *Switch) Set(newState bool) {
	if newState {
		s.pin2Battery.Charge()
	} else {
		s.pin2Battery.Discharge()
	}
}

// NSwitchBank is a convenient way to get any number of power emitters from a string of 0/1s
type NSwitchBank struct {
	Switches []pwrEmitter
}

// NewNSwitchBank takes a string of 0/1s and creates a variable length list of Switch structs initialized based on their off/on-ness
func NewNSwitchBank(bits string) (*NSwitchBank, error) {

	match, err := regexp.MatchString("^[01]+$", bits)
	if err != nil {
		return nil, err
	}
	if !match {
		err = fmt.Errorf("Input not in binary format: \"%s\"", bits)
		return nil, err
	}

	sb := &NSwitchBank{}

	for _, bit := range bits {
		sb.Switches = append(sb.Switches, NewSwitch(bit == '1'))
	}

	return sb, nil
}
