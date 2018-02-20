package circuit

import (
	"fmt"
	"regexp"
)

// Switch is a basic On/Off component typically used to be the initial input into circuits
type Switch struct {
	relay       *Relay        // innards of the switch are using a relay to control on/off
	pin2Battery *Battery      // switch on/off is controlled by charging/discharging this battery
	pwrSource                 // switch gains all that is pwrSource too
}

// NewSwitch returns a new switch whose initial state is based on the passed in initialization value
func NewSwitch(startState bool) *Switch {
	sw := &Switch{}
	sw.Init()

	// setup the battery-based relay pins, where pin2's battery will be used to toggle on/off of the switch (see Set(bool) method)
	sw.pin2Battery = NewBattery(startState)
	sw.relay = NewRelay(NewBattery(true), sw.pin2Battery)

	chState := make(chan Electron, 1)
	go func() {
		for {
			select {
			case e := <-chState:
				sw.Transmit(e.powerState)
				e.wg.Done()
			case <-sw.chStop:
				return
			}
		}
	}()

	// a switch acts like a relay, where Closed Out on the relay is the switch's power "answer"
	sw.relay.ClosedOut.WireUp(chState)

	return sw
}

// Shutdown will allow the go func, which is handling listen/transmit, to exit, and propogates the Shuthdown action to the internal relay
func (s *Switch) Shutdown() {
	s.relay.Shutdown()
	s.chStop <- true
}

// Set will toggle the power state of the underlying battery to activate/deactivate the internal relay, and therefore the switch's output power state
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
	for i := range sb.Switches {
		sb.Switches[i].(*Switch).Shutdown()
	}
}
