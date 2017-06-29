package circuit

// Reset-Set Flip-Flop

// r s   q  !q
// 0 1   1   0
// 1 0   0   1
// 0 0   q  !q  (hold)
// 1 1   x   x  (invalid)

type RSFlipFlop struct {
	Q    *NORGate
	QBar *NORGate
}

func NewRSFlipFLop(rPin, sPin pwrEmitter) *RSFlipFlop {
	ff := &RSFlipFlop{}

	// WARNING: must ensure QBar gets built first to ensure a case of double false on inputs ends up with QBar true

	ff.QBar = NewNORGate(sPin, nil) // ff.Qs doesn't exist yet so cannot setup feedback loop (pin2) yet
	ff.Q = NewNORGate(rPin, ff.QBar)

	ff.QBar.UpdatePin(2, 2, ff.Q) // now ff.Qs exists so wire up the feedback loop

	// perform sanity check on any input changes (can't do this really since things are transmitted sequential will can cause both inputs to be the same for a fraction of a second before stabilizing)
	//ff.Q.WireUp(ff.validateOutputRule)
	//ff.QBar.WireUp(ff.validateOutputRule)

	return ff
}
/*
func (f *RSFlipFlop) validateOutputRule(newState bool) {
	if f.Q.isPowered == f.QBar.isPowered {
		panic("A Flip-Flop cannot have equivalent power status at both Q and QBar")
	}
}
*/