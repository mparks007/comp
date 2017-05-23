package circuit

// NAND (using Inverter on an AND gate emit)
// 0 0 1
// 1 0 1
// 0 1 1
// 1 1 0

type nandGate2 struct {
	inverter emitter
}

func newNANDGate2(pin1, pin2 emitter) *nandGate2 {
	g := &nandGate2{}

	g.inverter = newInverter(newANDGate(pin1, pin2))

	return g
}

func (g *nandGate2) Emitting() bool {
	return g.inverter.Emitting()
}

// NOR (using Inverter on an OR gate emit)
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 0

type norGate2 struct {
	inverter emitter
}

func newNORGate2(pin1, pin2 emitter) *norGate2 {
	g := &norGate2{}

	g.inverter = newInverter(newORGate(pin1, pin2))

	return g
}

func (g *norGate2) Emitting() bool {
	return g.inverter.Emitting()
}

// XNOR (aka equivalence gate) (using Inverter on an XOR gate emit)
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 1

type xnorGate struct {
	inverter emitter
}

func newXNORGate(pin1, pin2 emitter) *xnorGate {
	g := &xnorGate{}

	g.inverter = newInverter(newXORGate(pin1, pin2))

	return g
}

func (g *xnorGate) Emitting() bool {
	return g.inverter.Emitting()
}
