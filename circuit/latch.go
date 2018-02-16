package circuit

import (
	"fmt"
	"time"
)

//import "fmt"

// RS/Reset-Set (or SR/Set-Reset) Flip-Flop

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

	fmt.Println("Creating loopback wires")
	wireQOut := NewWire(10)
	wireQBarOut := NewWire(10)

	fmt.Println("Creating QBar NOR")
	ff.QBar = NewNORGate(sPin, wireQOut)
	fmt.Println("Wiring channel up to QBar")
	ff.QBar.WireUp(wireQBarOut.Input)

	// give the WireUp's inner transmit time to wrap up
	time.Sleep(time.Millisecond * 10)

	fmt.Println("\nCreating Q NOR")
	ff.Q = NewNORGate(rPin, wireQBarOut)
	fmt.Println("Wiring channel up to Q")
	ff.Q.WireUp(wireQOut.Input)

	// give the WireUp's inner transmit time to wrap up
	time.Sleep(time.Millisecond * 10)

	return ff
}

// Level-triggered D-Type Latch ("Level" = clock high/low, "D" = data 0/1)

// d clk    q  !q
// 0  1     0  1
// 1  1     1  0
// X  0     q  !q  (data doesn't matter, no clock high to trigger a store-it action)

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

type NBitLevelTriggeredDTypeLatch struct {
	latches []*LevelTriggeredDTypeLatch
	Qs      []pwrEmitter
}

func NewNBitLevelTriggeredDTypeLatch(clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLevelTriggeredDTypeLatch {
	latch := &NBitLevelTriggeredDTypeLatch{}

	for _, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatch(clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
}

// Level-triggered D-Type Latch With Clear ("Level" = clock high/low, "D" = data 0/1)

// d clk clr   q  !q
// 0  1   0    0  1
// 1  1   0    1  0
// X  0   0    q  !q  (data doesn't matter, no clock high to trigger a store-it action)
// X  X   1    0  1   (clear forces reset so data and clock do not matter)
// 1  1   1    X  X   (invalid)

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

	wireSAndToQBarNor := NewWire(125) // need to ensure the Clear state resolves the QNor first, by slowing down the AND->QBarNor wire (using a "longer wire" between them)

	latch.sAnd.WireUp(wireSAndToQBarNor.Input)

	latch.rs = NewRSFlipFLop(latch.clrOR, wireSAndToQBarNor)

	// refer to the inner-flipflop's outputs for easier external access
	latch.Q = latch.rs.Q
	latch.QBar = latch.rs.QBar

	return latch
}

type NBitLevelTriggeredDTypeLatchWithClear struct {
	latches []*LevelTriggeredDTypeLatchWithClear
	Qs      []pwrEmitter
}

func NewNBitLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLevelTriggeredDTypeLatchWithClear {
	latch := &NBitLevelTriggeredDTypeLatchWithClear{}

	for _, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatchWithClear(clrPin, clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
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
