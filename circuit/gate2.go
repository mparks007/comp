package circuit

// AND
// 0 0 0
// 1 0 0
// 0 1 0
// 1 1 1

type ANDGate2 struct {
	relays []*Relay2
	bitPublication
}

func NewANDGate2(pins ...bitPublisher) *ANDGate2 {
	g := &ANDGate2{}

	for i, pin := range pins {
		if i == 0 {
			g.relays = append(g.relays, NewRelay2(&Battery{}, pin))
		} else {
			g.relays = append(g.relays, NewRelay2(&g.relays[i-1].ClosedOut, pin))
		}
	}

	// the last gate in the chain is the final answer for an AND
	g.relays[len(pins)-1].ClosedOut.Register(g.Publish)

	return g
}

// OR
// 0 0 0
// 1 0 1
// 0 1 1
// 1 1 1

type ORGate2 struct {
	relays []*Relay2
	bitPublication
}

func NewORGate2(pins ...bitPublisher) *ORGate2 {
	g := &ORGate2{}

	for i, pin := range pins {
		g.relays = append(g.relays, NewRelay2(&Battery{}, pin))

		// every relay can trigger state in a chain of ORs
		g.relays[i].ClosedOut.Register(g.powerUpdate)
	}

	return g
}

func (g *ORGate2) powerUpdate(newState bool) {
	newState = false

	// check to see if ANY of the relays are closed
	for _, r := range g.relays {
		if r.ClosedOut.isPowered {
			newState = true
			break
		}
	}

	g.Publish(newState)
}

// NAND
// 0 0 1
// 1 0 1
// 0 1 1
// 1 1 0

type NANDGate2 struct {
	relays []*Relay2
	bitPublication
}

func NewNANDGate2(pins ...bitPublisher) *NANDGate2 {
	g := &NANDGate2{}

	for i, pin := range pins {
		g.relays = append(g.relays, NewRelay2(&Battery{}, pin))

		// every relay can trigger state in a chain of NANDs
		g.relays[i].OpenOut.Register(g.powerUpdate)
	}

	return g
}

func (g *NANDGate2) powerUpdate(newState bool) {
	newState = false

	// check to see if ANY of the relays are open
	for _, r := range g.relays {
		if r.OpenOut.isPowered {
			newState = true
			break
		}
	}

	g.Publish(newState)

}

// NOR
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 0

type NORGate2 struct {
	relays []*Relay2
	bitPublication
}

func NewNORGate2(pins ...bitPublisher) *NORGate2 {
	g := &NORGate2{}

	for i, pin := range pins {
		if i == 0 {
			g.relays = append(g.relays, NewRelay2(&Battery{}, pin))
		} else {
			g.relays = append(g.relays, NewRelay2(&g.relays[i-1].OpenOut, pin))
		}
	}

	// the last gate in the chain is the final answer for a NOR
	g.relays[len(pins)-1].OpenOut.Register(g.Publish)

	return g
}

// XOR
// 0 0 0
// 1 0 1
// 0 1 1
// 1 1 0

type XORGate2 struct {
	orGate   bitPublisher
	nandGate bitPublisher
	andGate  bitPublisher
	bitPublication
}

func NewXORGate2(pin1, pin2 bitPublisher) *XORGate2 {
	g := &XORGate2{}

	g.orGate = NewORGate2(pin1, pin2)
	g.nandGate = NewNANDGate2(pin1, pin2)
	g.andGate = NewANDGate2(g.orGate, g.nandGate)

	// the state of the shared AND is the answer for an XOR
	g.andGate.Register(g.Publish)

	return g
}

// XNOR (aka equivalence gate) (using Inverter on an XOR gate)
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 1

type XNORGate struct {
	inverter bitPublisher
	bitPublication
}

func NewXNORGate(pin1, pin2 bitPublisher) *XNORGate {
	g := &XNORGate{}

	g.inverter = NewInverter(NewXORGate2(pin1, pin2))

	// the inverter owns the final answer in this approach to an XNOR
	g.inverter.Register(g.Publish)

	return g
}
