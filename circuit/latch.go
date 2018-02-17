package circuit

import "time"

// RSFlipFlop (Reset-Set) Flip-Flop is a standard flipflop circuit controlled by Set and Reset to output power at Q or "QBar" (QBar being opposite of Q)
// 	...or is this an SR (Set-Reset) Flip-Flop, or is this just an RS Latch?  SR Latch?  No matter for my purposes thus far...
//  This circuit is core to more complicated FlipFops/Latches to make even further complicated components (memory, counters, more?)
//
// Truth Table
// r s   q  !q
// 0 1   1   0
// 1 0   0   1
// 0 0   q  !q  (hold prior)
// 1 1   x   x  (generally deemed invalid)
type RSFlipFlop struct {
	wireQOut    *Wire
	wireQBarOut *Wire
	Q           *NORGate
	QBar        *NORGate
}

// NewRSFlipFLop returns an RSFlipFlop circuit which will be controlled by the passed in Reset/Set pins, resulting in varying states of its two outputs, Q and QBar
func NewRSFlipFLop(rPin, sPin pwrEmitter) *RSFlipFlop {
	ff := &RSFlipFlop{}

	ff.wireQOut = NewWire(10)
	ff.wireQBarOut = NewWire(10)

	ff.QBar = NewNORGate(sPin, ff.wireQOut)
	ff.QBar.WireUp(ff.wireQBarOut.Input)

	time.Sleep(time.Millisecond * 10)

	ff.Q = NewNORGate(rPin, ff.wireQBarOut)
	ff.Q.WireUp(ff.wireQOut.Input)

	time.Sleep(time.Millisecond * 10)

	return ff
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (f *RSFlipFlop) Shutdown() {
	f.Q.Shutdown()
	f.QBar.Shutdown()
	f.wireQBarOut.Shutdown()
	f.wireQOut.Shutdown()
}

// LevelTriggeredDTypeLatch is a type of latch/flipflop which will store the value of data only when the clock is high (on) ("Level" = clock high/low, "D" = data 0/1)
//
// Truth Table
// clk d   q  !q
//  1  0   0  1
//  1  1   1  0
//  0  X   q  !q  (data doesn't matter, no clock high to trigger a store-it action)
type LevelTriggeredDTypeLatch struct {
	rs       *RSFlipFlop
	rAnd     *ANDGate
	sAnd     *ANDGate
	inverter *Inverter
	Q        *NORGate
	QBar     *NORGate
}

// NewLevelTriggeredDTypeLatch returns a LevelTriggeredDTypeLatch circuit which will be controlled by the passed in Clock and Data pins, resulting in varying states of its two outputs, Q and QBar
func NewLevelTriggeredDTypeLatch(clkInPin, dataInPin pwrEmitter) *LevelTriggeredDTypeLatch {
	latch := &LevelTriggeredDTypeLatch{}

	latch.inverter = NewInverter(dataInPin)

	latch.rAnd = NewANDGate(clkInPin, latch.inverter)
	latch.sAnd = NewANDGate(clkInPin, dataInPin)

	latch.rs = NewRSFlipFLop(latch.rAnd, latch.sAnd)

	// refer to the inner-flipflop's outputs for easier external access
	latch.Q = latch.rs.Q
	latch.QBar = latch.rs.QBar

	return latch
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (l *LevelTriggeredDTypeLatch) Shutdown() {
	l.rs.Shutdown()
	l.sAnd.Shutdown()
	l.rAnd.Shutdown()
	l.inverter.Shutdown()
}

// NBitLevelTriggeredDTypeLatch is a component made up of a slice of LevelTriggeredDTypeLatch components
type NBitLevelTriggeredDTypeLatch struct {
	latches []*LevelTriggeredDTypeLatch
	Qs      []pwrEmitter
}

// NewNBitLevelTriggeredDTypeLatch returns an NBitLevelTriggeredDTypeLatch whose storing of the data pin value of EVERY internal latch will occur when the clock pin is on
func NewNBitLevelTriggeredDTypeLatch(clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLevelTriggeredDTypeLatch {
	latch := &NBitLevelTriggeredDTypeLatch{}

	for _, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatch(clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each latch, to exit
func (l *NBitLevelTriggeredDTypeLatch) Shutdown() {
	for i, _ := range l.latches {
		l.latches[i].Shutdown()
	}
}

// LevelTriggeredDTypeLatchWithClear is a type of latch/flipflop which will store the value of data only when the clock is high (on) and Clear is off ("Level" = clock high/low, "D" = data 0/1)
//	This is almost like a NewLevelTriggeredDTypeLatch, BUT it adds a Clear input, which if on, will force Q off no matter what else is going on
//
// Truth Table
// clk d clr   q  !q
//  1  0  0    0  1
//  1  1  0    1  0
//  0  X  0    q  !q  (data doesn't matter, no clock high to trigger a store-it action)
//  X  X  1    0  1   (clear forces reset so data and clock do not matter)
//  1  1  1    X  X   (generally deemed invalid)
type LevelTriggeredDTypeLatchWithClear struct {
	rs       *RSFlipFlop
	rAnd     *ANDGate
	sAnd     *ANDGate
	clrOR    *ORGate
	inverter *Inverter
	Q        *NORGate
	QBar     *NORGate
}

// NewLevelTriggeredDTypeLatchWithClear returns a LevelTriggeredDTypeLatchWithClear component controlled by a Clear, a Clock, both of which will control how the Data pin is handled
func NewLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin, dataInPin pwrEmitter) *LevelTriggeredDTypeLatchWithClear {
	latch := &LevelTriggeredDTypeLatchWithClear{}

	latch.inverter = NewInverter(dataInPin)

	latch.rAnd = NewANDGate(clkInPin, latch.inverter)
	latch.sAnd = NewANDGate(clkInPin, dataInPin)

	latch.clrOR = NewORGate(clrPin, latch.rAnd)

	wireSAndToQBarNor := NewWire(125) // need to ensure the Clear state resolves the QNor first, by slowing down the AND->QBarNor wire (using a "longer wire" between them)

	latch.sAnd.WireUp(wireSAndToQBarNor.Input)

	latch.rs = NewRSFlipFLop(latch.clrOR, wireSAndToQBarNor)
	//latch.rs = NewRSFlipFLop(latch.clrOR, latch.sAnd)

	// refer to the inner-flipflop's outputs for easier external access
	latch.Q = latch.rs.Q
	latch.QBar = latch.rs.QBar

	return latch
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (l *LevelTriggeredDTypeLatchWithClear) Shutdown() {
	l.rs.Shutdown()
	l.clrOR.Shutdown()
	l.sAnd.Shutdown()
	l.rAnd.Shutdown()
	l.inverter.Shutdown()
}

// NBitLevelTriggeredDTypeLatchWithClear is a component made up of a slice of LevelTriggeredDTypeLatchWithClear components
type NBitLevelTriggeredDTypeLatchWithClear struct {
	latches []*LevelTriggeredDTypeLatchWithClear
	Qs      []pwrEmitter
}

// NewNBitLevelTriggeredDTypeLatchWithClear returns an NBitLevelTriggeredDTypeLatchWithClear whose storing of the data pin value of EVERY internal latch will occur when the clock pin is on BUT the clear is off
//  If Clear is on, Q is forced off regardless of any other circuit power
func NewNBitLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLevelTriggeredDTypeLatchWithClear {
	latch := &NBitLevelTriggeredDTypeLatchWithClear{}

	for i, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[i].Q)
	}

	return latch
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each latch, to exit
func (l *NBitLevelTriggeredDTypeLatchWithClear) Shutdown() {
	for i, _ := range l.latches {
		l.latches[i].Shutdown()
	}
}

/*
// Edge-triggered D-Type Latch ("Level" = clock high/low, "D" = data 0/1)

// d clk   q  !q
// 0  ^    0  1
// 1  ^    1  0
// X  1    q  !q  (data doesn't matter, no clock transition to trigger a store-it action)
// X  0    q  !q  (data doesn't matter, no clock transition to trigger a store-it action)

type EdgeTriggeredDTypeLatch struct {
	lRAnd *ANDGate
	lSAnd *ANDGate
	lRS   *RSFlipFlop
	rRAnd *ANDGate
	rSAnd *ANDGate
	rRS   *RSFlipFlop
	Q     *NORGate
	QBar  *NORGate
}

func NewEdgeTriggeredDTypeLatch(clkInPin, dataInPin pwrEmitter) *EdgeTriggeredDTypeLatch {
	latch := &EdgeTriggeredDTypeLatch{}

	// for this to work, the clock wiring up has to be done against the right-side flipflop aspects FIRST
	latch.rRAnd = NewANDGate(clkInPin, nil)
	latch.rSAnd = NewANDGate(clkInPin, nil)
	latch.rRS = NewRSFlipFLop(latch.rRAnd, latch.rSAnd)

	latch.lRAnd = NewANDGate(NewInverter(clkInPin), dataInPin)
	latch.lSAnd = NewANDGate(NewInverter(clkInPin), NewInverter(dataInPin))
	latch.lRS = NewRSFlipFLop(latch.lRAnd, latch.lSAnd)

	latch.rRAnd.UpdatePin(2, 2, latch.lRS.Q)
	latch.rSAnd.UpdatePin(2, 2, latch.lRS.QBar)

	// refer to the inner-right-flipflop's outputs for easier external access
	latch.Q = latch.rRS.Q
	latch.QBar = latch.rRS.QBar

	return latch
}
/*
type FrequencyDivider struct {
	latch *EdgeTriggeredDTypeLatch
	Q     *NORGate
	QBar  *NORGate
}

func NewFrequencyDivider(oscillator pwrEmitter) *FrequencyDivider {
	freqDiv := &FrequencyDivider{}

	freqDiv.latch = NewSynchronizedEdgeTriggeredDTypeLatch(oscillator, nil)
	freqDiv.latch.UpdateDataPin(freqDiv.latch.QBar)

	// refer to the inner-right-flipflop's outputs for easier external access
	freqDiv.Q = freqDiv.latch.Q
	freqDiv.QBar = freqDiv.latch.QBar

	return freqDiv
}

/*
type NBitRippleCounter struct {
	freqDivs []*FrequencyDivider
	Qs       []*NORGate
}

func NewNBitRippleCounter(oscillator pwrEmitter, size int) *NBitRippleCounter {
	counter := &NBitRippleCounter{}

	for i := size - 1; i >= 0; i-- {
		var freqDiv *FrequencyDivider

		if i == size-1 {
			freqDiv = NewFrequencyDivider(oscillator)
		} else {
			freqDiv = NewFrequencyDivider(counter.freqDivs[0].QBar)
		}

		// prepend since going in reverse order
		counter.freqDivs = append([]*FrequencyDivider{freqDiv}, counter.freqDivs...)

		// make Qs refer to each for easier external access (pre-pending here too)
		counter.Qs = append([]*NORGate{freqDiv.Q}, counter.Qs...)
	}

	return counter
}
*/
