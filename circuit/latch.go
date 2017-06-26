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

func NewLevelTriggeredDTypeLatch(clkInPin, dataInPin pwrEmitter) *LevelTriggeredDTypeLatch {
	latch := &LevelTriggeredDTypeLatch{}

	latch.rAnd = NewANDGate(clkInPin, NewInverter(dataInPin))
	latch.sAnd = NewANDGate(clkInPin, dataInPin)

	latch.rs = NewRSFlipFLop(latch.rAnd, latch.sAnd)

	// refer to the inner-flipflop's outputs for easier external access
	latch.Q = latch.rs.Q
	latch.QBar = latch.rs.QBar

	return latch
}

type NBitLatch struct {
	latches []*LevelTriggeredDTypeLatch
	Qs      []pwrEmitter
}

func NewNBitLatch(clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLatch {
	latch := &NBitLatch{}

	for _, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatch(clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
}

type LevelTriggeredDTypeLatchWithClear struct {
	rs    *RSFlipFlop
	rAnd  *ANDGate
	sAnd  *ANDGate
	clrOR *ORGate
	Q     *NORGate
	QBar  *NORGate
}

func NewLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin, dataInPin pwrEmitter) *LevelTriggeredDTypeLatchWithClear {
	latch := &LevelTriggeredDTypeLatchWithClear{}

	latch.rAnd = NewANDGate(clkInPin, NewInverter(dataInPin))
	latch.sAnd = NewANDGate(clkInPin, dataInPin)

	latch.clrOR = NewORGate(clrPin, latch.rAnd)

	latch.rs = NewRSFlipFLop(latch.clrOR, latch.sAnd)

	// refer to the inner-flipflop's outputs for easier external access
	latch.Q = latch.rs.Q
	latch.QBar = latch.rs.QBar

	return latch
}

func (l *LevelTriggeredDTypeLatchWithClear) UpdateDataPin(dataPin pwrEmitter) {
	l.rAnd.UpdatePin(2, 2, NewInverter(dataPin))
	l.sAnd.UpdatePin(2, 2, dataPin)
}

type NBitLatchWithClear struct {
	latches []*LevelTriggeredDTypeLatchWithClear
	Qs      []pwrEmitter
}

func NewNBitLatchWithClear(clrPin, clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLatchWithClear {
	latch := &NBitLatchWithClear{}

	for _, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
}

func (l *NBitLatchWithClear) UpdateDataPins(dataPins []pwrEmitter) {

	// TODO: validate dataPins is same length as the latches slice

	for i, latch := range l.latches {
		latch.UpdateDataPin(dataPins[i])
	}
}
