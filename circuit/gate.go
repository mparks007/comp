package circuit

// AND
// 0 0 0
// 1 0 0
// 0 1 0
// 1 1 1

type ANDGate struct {
	relays []*Relay
	pwrSource
}

func NewANDGate(pins ...pwrEmitter) *ANDGate {
	g := &ANDGate{}

	for i, pin := range pins {
		if i == 0 {
			g.relays = append(g.relays, NewRelay(NewBattery(), pin))
		} else {
			g.relays = append(g.relays, NewRelay(&g.relays[i-1].ClosedOut, pin))
		}
	}

	// the last gate in the chain is the final answer for an AND
	g.relays[len(pins)-1].ClosedOut.WireUp(g.Transmit)

	return g
}

// OR
// 0 0 0
// 1 0 1
// 0 1 1
// 1 1 1

type ORGate struct {
	relays []*Relay
	pwrSource
}

func NewORGate(pins ...pwrEmitter) *ORGate {
	g := &ORGate{}

	for i, pin := range pins {
		g.relays = append(g.relays, NewRelay(NewBattery(), pin))

		// every relay can trigger state in a chain of ORs
		g.relays[i].ClosedOut.WireUp(g.powerUpdate)
	}

	return g
}

func (g *ORGate) powerUpdate(newState bool) {
	newState = false

	// check to see if ANY of the relays are closed
	for _, r := range g.relays {
		if r.ClosedOut.isPowered {
			newState = true
			break
		}
	}

	g.Transmit(newState)
}

// NAND
// 0 0 1
// 1 0 1
// 0 1 1
// 1 1 0

type NANDGate struct {
	relays []*Relay
	pwrSource
}

func NewNANDGate(pins ...pwrEmitter) *NANDGate {
	g := &NANDGate{}

	for i, pin := range pins {
		g.relays = append(g.relays, NewRelay(NewBattery(), pin))

		// every relay can trigger state in a chain of NANDs
		g.relays[i].OpenOut.WireUp(g.powerUpdate)
	}

	return g
}

func (g *NANDGate) powerUpdate(newState bool) {
	newState = false

	// check to see if ANY of the relays are open
	for _, r := range g.relays {
		if r.OpenOut.isPowered {
			newState = true
			break
		}
	}

	g.Transmit(newState)

}

// NOR
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 0

type NORGate struct {
	relays []*Relay
	pwrSource
}

func NewNORGate(pins ...pwrEmitter) *NORGate {
	g := &NORGate{}

	for i, pin := range pins {
		if i == 0 {
			g.relays = append(g.relays, NewRelay(NewBattery(), pin))
		} else {
			g.relays = append(g.relays, NewRelay(&g.relays[i-1].OpenOut, pin))
		}
	}

	// the last gate in the chain is the final answer for a NOR
	g.relays[len(pins)-1].OpenOut.WireUp(g.Transmit)

	return g
}

// XOR
// 0 0 0
// 1 0 1
// 0 1 1
// 1 1 0

type XORGate struct {
	orGate   pwrEmitter
	nandGate pwrEmitter
	andGate  pwrEmitter
	pwrSource
}

func NewXORGate(pin1, pin2 pwrEmitter) *XORGate {
	g := &XORGate{}

	g.orGate = NewORGate(pin1, pin2)
	g.nandGate = NewNANDGate(pin1, pin2)
	g.andGate = NewANDGate(g.orGate, g.nandGate)

	// the state of the shared AND is the answer for an XOR
	g.andGate.WireUp(g.Transmit)

	return g
}

// XNOR (aka equivalence gate) (using Inverter on an XOR gate)
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 1

type XNORGate struct {
	inverter pwrEmitter
	pwrSource
}

func NewXNORGate(pin1, pin2 pwrEmitter) *XNORGate {
	g := &XNORGate{}

	g.inverter = NewInverter(NewXORGate(pin1, pin2))

	// the inverter owns the final answer in this approach to an XNOR
	g.inverter.WireUp(g.Transmit)

	return g
}
