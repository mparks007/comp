package circuit

import "fmt"

// Inverter is a standard circuit that inverts the charge state of the input as its output
//
// Truth Table
// in out
// 0   1
// 1   0
type Inverter struct {
	relay *Relay // internal relay wired up in a way it reverses the charge state sent to the inverter component
	chargeSource
}

// NewInverter will return an Inverter component whose output will be the opposite of the passed in pin's charge state
func NewInverter(name string, pin chargeEmitter) *Inverter {
	inv := &Inverter{}
	inv.Init()
	inv.Name = name

	cp := NewChargeProvider(fmt.Sprintf("%s-Relay-pin1ChargeProvider", name), true)
	inv.relay = NewRelay(fmt.Sprintf("%s-Relay", name), cp, pin)

	chState := make(chan Charge, 1)
	go func() {
		for {
			select {
			case c := <-chState:
				Debug(name, fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
				inv.Transmit(c)
				c.Done()
			case <-inv.chStop:
				Debug(name, "Stopped")
				return
			}
		}
	}()

	// in an Inverter, the Open Out is what gets the flipped charge state as the output
	inv.relay.OpenOut.WireUp(chState)

	return inv
}

// Shutdown will allow the go funcs, which are handling listen/transmit on the inner relay and the inverter itself, to exit
func (inv *Inverter) Shutdown() {
	inv.relay.Shutdown()
	inv.chStop <- true
}
