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
	f := &RSFlipFlop{}

	f.QBar = NewNORGate(sPin, nil) // f.Q doesn't exist yet so cannot setup feedback loop yet
	f.Q = NewNORGate(rPin, f.QBar)

	f.QBar.UpdatePin(2, 2, f.Q) // now f.Q exists so wire up the feedback loop

	f.Q.WireUp(f.validateOutputRule)
	f.QBar.WireUp(f.validateOutputRule)

	return f
}

func (f *RSFlipFlop) validateOutputRule(newState bool) {
	if f.Q.isPowered == f.QBar.isPowered {
		panic("A Flip-Flop cannot have equivalent power status at both Q and QBar")
	}
}
