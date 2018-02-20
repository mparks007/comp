package circuit

// Inverter is a standard circuit that inverts the power state of the input
//
// Truth Table
// in out
// 0   1
// 1   0
type Inverter struct {
	relay     *Relay // internal relay wired up in a way it reverses the power state sent to the inverter component
	pwrSource        // inverter gains all that is pwrSource too
}

// NewInverter will return an Inverter component whose output will be the opposite of the passed in pin's power state
func NewInverter(pin pwrEmitter) *Inverter {
	inv := &Inverter{}
	inv.Init()

	inv.relay = NewRelay(NewBattery(true), pin)

	chState := make(chan Electron, 1)
	go func() {
		for {
			select {
			case e := <-chState:
				inv.Transmit(e.powerState)
				e.wg.Done()
			case <-inv.chStop:
				return
			}
		}
	}()

	// in an Inverter, the Open Out is what gets the flipped state (the "answer")
	inv.relay.OpenOut.WireUp(chState)

	return inv
}

// Shutdown will allow the go funcs, which are handling listen/transmit on the inner relay and the inverter itself, to exit
func (inv *Inverter) Shutdown() {
	inv.relay.Shutdown()
	inv.chStop <- true
}
