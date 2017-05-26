package circuit

type rsFlipFlop struct {
	rIn     emitter
	sIn     emitter
	rNor    emitter
	sNor    emitter
	qOut    emitter
	qBarOut emitter
}

func newRSFlipFLop(rPin, sPin emitter) *rsFlipFlop {
	f := &rsFlipFlop{}

	f.rIn = rPin
	f.sIn = sPin
	f.rNor = &norGate{}
	f.sNor = &norGate{}

	rNorGate, _ := f.rNor.(*norGate)
	sNorGate, _ := f.sNor.(*norGate)

	rNorGate.relay1 = &relay{}
	rNorGate.relay2 = &relay{}

	rNorGate.relay1 = &relay{}
	rNorGate.relay1.openOut = newXContact(&Battery{}, sNorGate.relay1.openOut)

	rNorGate.relay2 = &relay{}
	rNorGate.relay2.openOut = newXContact(&Battery{}, sNorGate.relay1.openOut)
	rNorGate.relay2.closedOut = newANDContact(&Battery{}, sNorGate.relay1.openOut)

	rNorGate.relay1.aIn = f.rIn
	rNorGate.relay1.bIn = sNorGate.relay2.openOut

	sNorGate.relay1.aIn = f.sIn
	sNorGate.relay1.bIn = rNorGate.relay2.openOut

	f.qOut = rNorGate.relay2.openOut
	f.qBarOut = sNorGate.relay2.openOut

	return f
}
