package circuit

// Level-triggered D-Type Latch ("Level" = clock high/low, "D" = data 0/1)

// d clk   q  !q
// 0 1     0  1
// 1 1     1  0
// X 0     q  !q  (data doesn't matter, no clock high to trigger a store-it action)

type ltDLatch struct {
	dataIn emitter
	clkIn  emitter
	rs     *rsFlipFlop
	rAnd   *andGate
	sAnd   *andGate
}

func newLtDLatch(clkIn, dataIn emitter) (*ltDLatch, error) {
	l := &ltDLatch{}

	l.updateInputs(clkIn, dataIn)

	l.rs, _ = newRSFlipFLop(nil, nil) // make defaulted inner flipflop. setupComponents will set it up fully

	err := l.setupComponents()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (l *ltDLatch) updateInputs(clkIn, dataIn emitter) {
	l.dataIn = dataIn
	l.clkIn = clkIn
}

func (l *ltDLatch) setupComponents() error {
	l.rAnd = newANDGate(newInverter(l.dataIn), l.clkIn)
	l.sAnd = newANDGate(l.dataIn, l.clkIn)

	// pass along the new current input states to the inner flipflop
	err := l.rs.updateInputs(l.rAnd, l.sAnd)
	if err != nil {
		return err
	}

	return nil
}

func (l *ltDLatch) qEmitting() (bool, error) {

	err := l.setupComponents()
	if err != nil {
		return false, err
	}

	if qEmitting, err := l.rs.qEmitting(); err != nil {
		return qEmitting, err
	} else {
		return qEmitting, nil
	}
}

func (l *ltDLatch) qBarEmitting() (bool, error) {
	if qEmitting, err := l.qEmitting(); err != nil {
		return !qEmitting, err
	} else {
		return !qEmitting, nil
	}
}
