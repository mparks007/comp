package circuit

// Level-triggered D-Type Latch ("Level" = clock high/low, "D" = data 0/1)

// d clk   q  !q
// 0 1     0  1
// 1 1     1  0
// X 0     q  !q  (data doesn't matter, no clock high to trigger a store-it action)

type LevelTriggeredDTypeLatch struct {
	rs   *RSFlipFlop
	rAnd *ANDGate
	sAnd *ANDGate
	Q    *NORGate
	QBar *NORGate
}

func NewLevelTriggeredDTypeLatch(clkIn, dataIn pwrEmitter) *LevelTriggeredDTypeLatch {
	l := &LevelTriggeredDTypeLatch{}

	l.rAnd = NewANDGate(clkIn, NewInverter(dataIn))
	l.sAnd = NewANDGate(clkIn, dataIn)

	l.rs = NewRSFlipFLop(l.rAnd, l.sAnd)

	// refer to the inner-flipflop's outputs for easier, external access
	l.Q = l.rs.Q
	l.QBar = l.rs.QBar

	return l
}

type EightBitLatch struct {
	latches [8]*LevelTriggeredDTypeLatch
	Qs      [8]*NORGate
}

func NewEightBitLatch(clkIn pwrEmitter, dataIn [8]pwrEmitter) *EightBitLatch {
	l := &EightBitLatch{}

	for i, d := range dataIn {
		l.latches[i] = NewLevelTriggeredDTypeLatch(clkIn, d)

		// refer to the inner-latch's Qs output for easier, external access
		l.Qs[i] = l.latches[i].Q
	}

	return l
}

// AsPwrEmitters will return pwrEmitter versions of the internal latch's Qs out
func (l *EightBitLatch) AsPwrEmitters() [8]pwrEmitter {
	pwrEmits := [8]pwrEmitter{}

	for i, latch := range l.latches {
		pwrEmits[i] = latch.Q
	}

	return pwrEmits
}
