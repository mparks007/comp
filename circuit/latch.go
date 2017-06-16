package circuit

// Level-triggered D-Type Latch ("Level" = clock high/low, "D" = data 0/1)

// d clk   q  !q
// 0 1     0  1
// 1 1     1  0
// X 0     q  !q  (data doesn't matter, no clock high to trigger s store-it action)

type LevelTriggeredDTypeLatch struct {
	rs   *RSFlipFlop
	rAnd *ANDGate
	sAnd *ANDGate
	Q    *NORGate
	QBar *NORGate
}

func NewLevelTriggeredDTypeLatch(clkIn, dataIn pwrEmitter) *LevelTriggeredDTypeLatch {
	latch := &LevelTriggeredDTypeLatch{}

	latch.rAnd = NewANDGate(clkIn, NewInverter(dataIn))
	latch.sAnd = NewANDGate(clkIn, dataIn)

	latch.rs = NewRSFlipFLop(latch.rAnd, latch.sAnd)

	// refer to the inner-flipflop's outputs for easier external access
	latch.Q = latch.rs.Q
	latch.QBar = latch.rs.QBar

	return latch
}

type NBitLatch struct {
	latches []*LevelTriggeredDTypeLatch
	Qs      []*NORGate
}

func NewNBitLatch(clkIn pwrEmitter, dataIn []pwrEmitter) *NBitLatch {
	latch := &NBitLatch{}

	for _, data := range dataIn {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatch(clkIn, data))

		// refer to the inner-latch's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
}

// AsPwrEmitters will return pwrEmitter versions of the internal latch's Qs out
func (l *NBitLatch) AsPwrEmitters() []pwrEmitter {
	pwrEmits := []pwrEmitter{}

	for _, latch := range l.latches {
		pwrEmits = append(pwrEmits, latch.Q)
	}

	return pwrEmits
}
