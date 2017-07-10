package circuit

import "fmt"

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

	// the Open Outs is what gets the flipped state in an Inverter
	//inv.relay.OpenOut.WireUp(inv.Transmit)
	inv.relay.OpenOut.WireUp(inv.PowerUpdate)

	return inv
}

func (i *Inverter) PowerUpdate(newState bool) {

	fmt.Printf("Inverter address: %v\n", i)
	i.Transmit(newState)
}
