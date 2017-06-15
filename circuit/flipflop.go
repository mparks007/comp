package circuit

// Reset-Set Flip-Flop

// r l   q  !q
// 0 1   1   0
// 1 0   0   1
// 0 0   q  !q  (hold)
// 1 1   x   x  (invalid)

type RSFlipFlop struct {
	Q    *NORGate
	QBar *NORGate
}

func NewRSFlipFLop(rPin, sPin pwrEmitter) *RSFlipFlop {
	f := &RSFlipFlop{}

	// must ensure QBar gets built first to ensure a case of double false on inputs ends up with QBar true

	f.QBar = NewNORGate(sPin, nil) // f.Qs doesn't exist yet so cannot setup feedback loop (pin2) yet
	f.Q = NewNORGate(rPin, f.QBar)

	f.QBar.UpdatePin(2, 2, f.Q) // now f.Qs exists so wire up the feedback loop

	// perform sanity check on any input changes
	f.Q.WireUp(f.validateOutputRule)
	f.QBar.WireUp(f.validateOutputRule)

	return f
}

func (f *RSFlipFlop) validateOutputRule(newState bool) {
	if f.Q.isPowered == f.QBar.isPowered {
		panic("A Flip-Flop cannot have equivalent power status at both Qs and QBar")
	}
}
