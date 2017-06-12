package circuit

// Inverter
// 0 -> 1
// 1 -> 0

type Inverter struct {
	relay *Relay
	pwrSource
}

func NewInverter(pin pwrEmitter) *Inverter {
	inv := &Inverter{}

	inv.relay = NewRelay(NewBattery(), pin)

	// the Open Out is what gets the flipped state in an Inverter
	inv.relay.OpenOut.WireUp(inv.Transmit)

	return inv
}
