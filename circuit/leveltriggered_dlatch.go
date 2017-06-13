package circuit

// Level-triggered D-Type Latch ("Level" = clock high/low, "D" = data 0/1)

// d clk   q  !q
// 0 1     0  1
// 1 1     1  0
// X 0     q  !q  (data doesn't matter, no clock high to trigger a store-it action)

type LevTrigDLatch struct {
	rs   *RSFlipFlop
	rAnd *ANDGate
	sAnd *ANDGate
	Q    *NORGate
	QBar *NORGate
}

func NewLtDLatch(clkIn, dataIn pwrEmitter) *LevTrigDLatch {
	l := &LevTrigDLatch{}

	l.rAnd = NewANDGate(clkIn, NewInverter(dataIn))
	l.sAnd = NewANDGate(clkIn, dataIn)

	l.rs = NewRSFlipFLop(l.rAnd, l.sAnd)

	// refer to the inner-flipflop's outputs for easier, external access
	l.Q = l.rs.Q
	l.QBar = l.rs.QBar

	return l
}
