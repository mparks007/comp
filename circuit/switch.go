package circuit

import (
	"fmt"
	"regexp"
)

// Switch is a basic On/Off component typically used to be the initial input into circuits
type Switch struct {
	relay       *Relay
	pin2Battery *Battery
	ch          chan bool
	pwrSource
}

// NewSwitch returns a new switch whose initial state is based on the passed in initialization value
func NewSwitch(startState bool) *Switch {
	sw := &Switch{}
	sw.Init()

	sw.ch = make(chan bool, 1)

	sw.pin2Battery = NewBattery()
	if !startState {
		sw.pin2Battery.Discharge()
	}
	// setup the battery-based relay pins which will be used to toggle on/off of the switch (see Set(bool) method)
	sw.relay = NewRelay(NewBattery(), sw.pin2Battery)

	// a switch acts like a relay, where Closed Out on the relay is the switch's power "answer"
	sw.relay.ClosedOut.WireUp(sw.ch)

	transmit := func() {
		sw.Transmit(<-sw.ch)
	}

	// calling transmit explicitly to ensure the 'answer' for the output, post WireUp above, has settled BEFORE returning and letting things wire up to it
	transmit()

	go func() {
		for {
			select {
			case <-sw.chDone:
				return
			default:
				transmit()
			}
		}
	}()

	return sw
}

// Shutdown will allow the go func, which is handling listen/transmit, to exit, and propogates the Shuthdown action to the internal relay
func (s *Switch) Shutdown() {
	s.relay.Shutdown()
	s.chDone <- true
}

// Set on a switch will toggle the power state of the underlying battery to activate/deactivate the internal relay, and therefore the switch's output power state
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

// NewNSwitchBank takes a string of 0/1s and creates a slice of Switch objects, where each one is independently initialized based on "0" or "1" in the string
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

// Shutdown will allow the go funcs, which are handling listen/transmit on each switch, to exit
func (sb *NSwitchBank) Shutdown() {
	for i, _ := range sb.Switches {
		sb.Switches[i].(*Switch).Shutdown()
	}
}
