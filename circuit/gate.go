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
