package circuit

import (
	"fmt"
	"regexp"
)

// Switch is a basic On/Off component typically used to be the initial input into circuits
type Switch struct {
	relay       *Relay   // innards of the switch are using a relay to control on/off
	pin2Battery *Battery // switch on/off is controlled by charging/discharging this battery
	pwrSource            // switch gains all that is pwrSource too
}

// NewSwitch returns a new switch whose initial state is based on the passed in initialization value
func NewSwitch(name string, startState bool) *Switch {
	sw := &Switch{}
	sw.Init()
	sw.Name = name

	// setup the battery-based relay pins, where pin2's battery will be used to toggle on/off of the switch (see Set(bool) method)
	sw.pin2Battery = NewBattery(fmt.Sprintf("%s-Relay-pin2Battery", name), startState)
	sw.relay = NewRelay(fmt.Sprintf("%s-Relay", name), NewBattery(fmt.Sprintf("%s-Relay-pin1Battery", name), true), sw.pin2Battery)

	chState := make(chan Electron, 1)
	go func() {
		for {
			select {
			case e := <-chState:
				Debug(name, fmt.Sprintf("Received on Channel (%v), Electron {%s}", chState, e.String()))
				sw.Transmit(e)
				e.Done()
			case <-sw.chStop:
				Debug(name, "Stopped")
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
	switches []pwrEmitter
}

// NewNSwitchBank takes a string of 0/1s and creates a slice of Switch objects, where each one is independently initialized based on "0" or "1" in the string
func NewNSwitchBank(name string, bits string) (*NSwitchBank, error) {

	match, err := regexp.MatchString("^[01]+$", bits)
	if err != nil {
		return nil, err
	}
	if !match {
		err = fmt.Errorf("Input not in binary format: \"%s\"", bits)
		return nil, err
	}

	sb := &NSwitchBank{}

	for i, bit := range bits {
		sb.switches = append(sb.switches, NewSwitch(fmt.Sprintf("%s-Switches[%d]", name, i), bit == '1'))
	}

	return sb, nil
}

// Switches returns the internal switch slice for the switch bank (still allows external altering of the inner variable?  fix if so...)
func (sb *NSwitchBank) Switches() []pwrEmitter {
	return sb.switches
}

// SetSwitches will flip the switches of the switch bank to match a passed in bits string
func (sb *NSwitchBank) SetSwitches(bits string) {
	for i, b := range bits {
		sb.switches[i].(*Switch).Set(b == '1')
	}
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each switch, to exit
func (sb *NSwitchBank) Shutdown() {
	for _, sw := range sb.switches {
		sw.(*Switch).Shutdown()
	}
}
