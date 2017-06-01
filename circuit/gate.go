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

type AndGate2 struct {
	pin1Powered bool
	pin2Powered bool
	relay1      *Relay2
	relay2      *Relay2
}

func NewANDGate2(pin1, pin2 powerPublisher) *AndGate2 {
	g := &AndGate2{}

	g.relay1 = NewRelay2(&Battery{}, pin1)
	g.relay2 = NewRelay2(g.relay1.ClosedOut, pin2)

	pin1.Subscribe(g.pin1PowerChange)
	pin2.Subscribe(g.pin2PowerChange)

	return g
}

func (g *AndGate2) pin1PowerChange(state bool) {
	if g.pin1Powered != state {
		g.pin1Powered = state
		//r.OpenOut.Publish(state && !r.bInPowered)
	}
}

func (g *AndGate2) pin2PowerChange(state bool) {
	if g.pin2Powered != state {
		g.pin2Powered = state
		//r.ClosedOut.Publish(state && r.aInPowered)
		//r.OpenOut.Publish(state && !r.bInPowered)
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
