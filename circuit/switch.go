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
	ch          chan bool
	pwrSource
}

// NewSwitch creates a new Switch struct with its initial state based on the passed in initialization value
func NewSwitch(init bool) *Switch {
	sw := &Switch{}
	sw.ch = make(chan bool, 1)

	// setup the battery-based relay pin that will be used to toggle on/off of the switch
	sw.pin2Battery = NewBattery()
	if !init {
		sw.pin2Battery.Discharge()
	}

	sw.relay = NewRelay(NewBattery(), sw.pin2Battery)
	fmt.Println("Switch Wiring up to Relay Closedout")
	sw.relay.ClosedOut.WireUp(sw.ch)

	go func() {
		for {
			sw.Transmit(<-sw.ch)
		}
	}()

	return sw
}

// Set on a Switch will toggle the state of the underlying battery to activate/deactivate the internal relay
func (s *Switch) Set(newState bool) {
	if newState {
		s.pin2Battery.Charge()
	} else {
		s.pin2Battery.Discharge()
	}
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
