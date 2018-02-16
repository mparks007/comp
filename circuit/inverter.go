package circuit

// Inverter
// 0 -> 1
// 1 -> 0

type Inverter struct {
	relay *Relay
	ch    chan bool
	pwrSource
}

func NewInverter(pin pwrEmitter) *Inverter {
	inv := &Inverter{}
	inv.ch = make(chan bool, 1)

	inv.relay = NewRelay(NewBattery(), pin)

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
			case <-inv.chDone:
				return
			default:
				transmit()
			}
		}
	}()

	return inv
}

func (inv *Inverter) Shutdown() {
	inv.relay.Shutdown()
	inv.chDone <- true
}
