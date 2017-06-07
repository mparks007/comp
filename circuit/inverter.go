package circuit

// Inverter
// 0 -> 1
// 1 -> 0

type Inverter struct {
	relay *Relay
	bitPublication
}

func NewInverter(pin bitPublisher) *Inverter {
	inv := &Inverter{}

	inv.relay = NewRelay(NewBattery(), pin)

	// the Open Out is what gets the flipped state in an Inverter
	inv.relay.OpenOut.Register(inv.Publish)

	return inv
}
