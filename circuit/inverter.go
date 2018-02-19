package circuit
/*
// Inverter is a standard circuit that inverts the power state of the input
//
// Truth Table
// in out
// 0   1
// 1   0
type Inverter struct {
	relay *Relay
	ch    chan bool
	pwrSource
}

// NewInverter will return an Inverter component whose output will be the opposite of the passed in pin's power state
func NewInverter(pin pwrEmitter) *Inverter {
	inv := &Inverter{}
	inv.Init()

	inv.ch = make(chan bool, 1)

	inv.relay = NewRelay(NewBattery(true), pin)

	// in an Inverter, the Open Out is what gets the flipped state (the "answer")
	inv.relay.OpenOut.WireUp(inv.ch)

	transmit := func() {
		inv.Transmit(<-inv.ch)
	}

	// calling transmit explicitly to ensure the 'answer' for the output, post WireUp above, has settled BEFORE returning and letting things wire up to it
	transmit()

	go func() {
		for {
			select {
			case <-inv.chStop:
				return
			default:
				transmit()
			}
		}
	}()

	return inv
}

// Shutdown will allow the go funcs, which are handling listen/transmit on the inner relay and the inverter itself, to exit
func (inv *Inverter) Shutdown() {
	inv.relay.Shutdown()
	inv.chStop <- true
}
*/