package circuit

// AND
// 0 0 0
// 1 0 0
// 0 1 0
// 1 1 1

type andGate struct {
	relay1 *relay
	relay2 *relay
}

func newANDGate(pin1, pin2 emitter) *andGate {
	g := &andGate{}

	g.relay1 = newRelay(&Battery{}, pin1)
	g.relay2 = newRelay(g.relay1.closedOut, pin2)

	return g
}

func (g *andGate) Emitting() bool {
	return g.relay2.closedOut.Emitting()
}

type ANDGate2 struct {
	relays []*Relay2
	publication
}

func NewANDGate2(pins ...publisher) *ANDGate2 {
	g := &ANDGate2{}

	for i, p := range pins {
		if i == 0 {
			g.relays = append(g.relays, NewRelay2(&Battery{}, p))
		} else {
			g.relays = append(g.relays, NewRelay2(&g.relays[i-1].ClosedOut, p))
		}
	}

	g.relays[len(pins)-1].ClosedOut.Register(g.powerChange)

	return g
}

func (g *ANDGate2) powerChange(state bool) {
	if g.state != state {
		g.state = state
		g.Publish()
	}
}

// OR
// 0 0 0
// 1 0 1
// 0 1 1
// 1 1 1

type orGate struct {
	relay1 *relay
	relay2 *relay
}

func newORGate(pin1, pin2 emitter) *orGate {
	return &orGate{
		newRelay(&Battery{}, pin1),
		newRelay(&Battery{}, pin2),
	}
}

func (g *orGate) Emitting() bool {
	return g.relay1.closedOut.Emitting() || g.relay2.closedOut.Emitting()
}

type ORGate2 struct {
	relays []*Relay2
	publication
}

func NewORGate2(pins ...publisher) *ORGate2 {
	g := &ORGate2{}

	for i, p := range pins {
		g.relays = append(g.relays, NewRelay2(&Battery{}, p))
		g.relays[i].ClosedOut.Register(g.powerChange)
	}

	return g
}

func (g *ORGate2) powerChange(state bool) {
	newState := false

	// check to see if ANY of the relays are closed
	for _, r := range g.relays {
		if r.ClosedOut.state {
			newState = true
			break
		}
	}

	if g.state != newState {
		g.state = newState
		g.Publish()
	}
}

// NAND
// 0 0 1
// 1 0 1
// 0 1 1
// 1 1 0

type nandGate struct {
	relay1 *relay
	relay2 *relay
}

func newNANDGate(pin1, pin2 emitter) *nandGate {
	return &nandGate{
		newRelay(&Battery{}, pin1),
		newRelay(&Battery{}, pin2),
	}
}

func (g *nandGate) Emitting() bool {
	return g.relay1.openOut.Emitting() || g.relay2.openOut.Emitting()
}

type NANDGate2 struct {
	relays []*Relay2
	publication
}

func NewNANDGate2(pins ...publisher) *NANDGate2 {
	g := &NANDGate2{}

	for i, p := range pins {
		g.relays = append(g.relays, NewRelay2(&Battery{}, p))
		g.relays[i].OpenOut.Register(g.powerChange)
	}

	return g
}

func (g *NANDGate2) powerChange(state bool) {
	newState := false

	// check to see if ANY of the relays are open
	for _, r := range g.relays {
		if r.OpenOut.state {
			newState = true
			break
		}
	}

	if g.state != newState {
		g.state = newState
		g.Publish()
	}
}

// NOR
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 0

type norGate struct {
	relay1 *relay
	relay2 *relay
}

func newNORGate(pin1, pin2 emitter) *norGate {
	g := &norGate{}

	g.relay1 = newRelay(&Battery{}, pin1)
	g.relay2 = newRelay(g.relay1.openOut, pin2)

	return g
}

func (g *norGate) Emitting() bool {
	return g.relay2.openOut.Emitting()
}

// XOR
// 0 0 0
// 1 0 1
// 0 1 1
// 1 1 0

type xorGate struct {
	orGate   emitter
	nandGate emitter
	andGate  emitter
}

func newXORGate(pin1, pin2 emitter) *xorGate {
	g := &xorGate{}

	g.orGate = newORGate(pin1, pin2)
	g.nandGate = newNANDGate(pin1, pin2)
	g.andGate = newANDGate(g.orGate, g.nandGate)

	return g
}

func (g *xorGate) Emitting() bool {
	return g.andGate.Emitting()
}
