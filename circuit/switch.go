package circuit

import (
	"fmt"
	"regexp"
)

// Switch is a basic On/Off component typically used to be the initial input into circuits
type Switch struct {
	relay              *Relay          // innards of the switch are using a relay to control on/off
	pin2ChargeProvider *ChargeProvider // switch on/off is controlled by charging/discharging this ChargeProvider (which controls how the inner relay outputs)
	chargeSource
}

// NewSwitch returns a new switch whose initial state is based on the passed in initialization value
func NewSwitch(name string, startState bool) *Switch {
	sw := &Switch{}
	sw.Init()
	sw.Name = name

	// setup the ChargeProvider-based relay pins, where pin2's ChargeProvider will be used to toggle on/off of the switch (see Set(bool) method)
	sw.pin2ChargeProvider = NewChargeProvider(fmt.Sprintf("%s-Relay-pin2ChargeProvider", name), startState)
	sw.relay = NewRelay(fmt.Sprintf("%s-Relay", name), NewChargeProvider(fmt.Sprintf("%s-Relay-pin1ChargeProvider", name), true), sw.pin2ChargeProvider)

	chState := make(chan Charge, 1)
	go func() {
		for {
			select {
			case c := <-chState:
				Debug(name, fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
				sw.Transmit(c)
				c.Done()
			case <-sw.chStop:
				Debug(name, "Stopped")
				return
			}
		}
	}()

	// a switch acts like a relay, where Closed Out on the relay is the switch's charge "answer"
	sw.relay.ClosedOut.WireUp(chState)

	return sw
}

// Shutdown will allow the go func, which is handling listen/transmit, to exit, and propogates the Shuthdown action to the internal relay
func (sw *Switch) Shutdown() {
	sw.relay.Shutdown()
	sw.chStop <- true
}

// Set will toggle the charge state of the underlying ChargeProvider to activate/deactivate the internal relay, and therefore the switch's output charge state
func (sw *Switch) Set(newState bool) {
	if newState {
		sw.pin2ChargeProvider.Charge()
	} else {
		sw.pin2ChargeProvider.Discharge()
	}
}

// NSwitchBank is a convenient way to get any number of charge emitters from a string of 0/1s
type NSwitchBank struct {
	switches []chargeEmitter
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

	// build out the switch slices for each bit sent in
	for i, bit := range bits {
		sb.switches = append(sb.switches, NewSwitch(fmt.Sprintf("%s-Switches[%d]", name, i), bit == '1'))
	}

	return sb, nil
}

// Switches returns the internal switch slice for the switch bank (still allows external altering of the inner variable?  fix if so...)
func (sb *NSwitchBank) Switches() []chargeEmitter {
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
