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

func NewRSFlipFLop(rPin, sPin bitPublisher) *RSFlipFlop {
	f := &RSFlipFlop{}

	var t bitPublisher
	f.QBar = NewNORGate(sPin, t)
	f.Q = NewNORGate(rPin, f.QBar)
	f.QBar = NewNORGate(sPin, f.Q)

//	f.Q.Register(f.ValidateOutputRule)
//	f.QBar.Register(f.ValidateOutputRule)

	return f
}

func (f *RSFlipFlop) ValidateOutputRule(newState bool) {
	if f.Q.isPowered == f.QBar.isPowered {
		panic("A Flip-Flop cannot be powered simultaneously at both Q and QBar")
	}
}
