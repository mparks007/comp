package circuit

type relay struct {
	aIn       emitter
	bIn       emitter
	openOut   emitter
	closedOut emitter
}

func newRelay(a emitter, b emitter) *relay {
	return &relay{
		a,
		b,
		newOutOpenPin(a, b),
		newOutClosedPin(a, b),
	}
}

// DELETE THESE TWO
func (r *relay) emittingOpen() bool {
	return r.openOut != nil && r.openOut.Emitting()
}

func (r *relay) emittingClosed() bool {
	return r.closedOut != nil && r.closedOut.Emitting()
}
