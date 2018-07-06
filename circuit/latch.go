package circuit

import "fmt"

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
func NewRSFlipFlop(name string, rPin, sPin pwrEmitter) *RSFlipFlop {
	ff := &RSFlipFlop{}

	ff.wireQOut = NewWire(fmt.Sprintf("%s-QOutWire", name), 0)
	ff.wireQBarOut = NewWire(fmt.Sprintf("%s-QBarOutWire", name), 0)

	ff.QBar = NewNORGate(fmt.Sprintf("%s-QBarNORGate", name), sPin, ff.wireQOut)
	ff.QBar.WireUp(ff.wireQBarOut.Input)

	ff.Q = NewNORGate(fmt.Sprintf("%s-QNORGate", name), rPin, ff.wireQBarOut)
	ff.Q.WireUp(ff.wireQOut.Input)

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
func NewLevelTriggeredDTypeLatch(name string, clkInPin, dataInPin pwrEmitter) *LevelTriggeredDTypeLatch {
	latch := &LevelTriggeredDTypeLatch{}

	latch.inverter = NewInverter(fmt.Sprintf("%s-Inverter", name), dataInPin)

	latch.rAnd = NewANDGate(fmt.Sprintf("%s-rANDGate", name), clkInPin, latch.inverter)
	latch.sAnd = NewANDGate(fmt.Sprintf("%s-sANDGate", name), clkInPin, dataInPin)

	latch.rs = NewRSFlipFlop(fmt.Sprintf("%s-RSFlipFlop", name), latch.rAnd, latch.sAnd)

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
func NewNBitLevelTriggeredDTypeLatch(name string, clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLevelTriggeredDTypeLatch {
	latch := &NBitLevelTriggeredDTypeLatch{}

	for i, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatch(fmt.Sprintf("%s-Latches[%d]", name, i), clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[len(latch.latches)-1].Q)
	}

	return latch
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each latch, to exit
func (l *NBitLevelTriggeredDTypeLatch) Shutdown() {
	for _, l := range l.latches {
		l.Shutdown()
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
func NewLevelTriggeredDTypeLatchWithClear(name string, clrPin, clkInPin, dataInPin pwrEmitter) *LevelTriggeredDTypeLatchWithClear {
	latch := &LevelTriggeredDTypeLatchWithClear{}

	latch.inverter = NewInverter(fmt.Sprintf("%s-Inverter", name), dataInPin)

	latch.rAnd = NewANDGate(fmt.Sprintf("%s-rANDGate", name), clkInPin, latch.inverter)
	latch.sAnd = NewANDGate(fmt.Sprintf("%s-sANDGate", name), clkInPin, dataInPin)

	latch.clrOR = NewORGate(fmt.Sprintf("%s-clrORGate", name), clrPin, latch.rAnd)

	latch.rs = NewRSFlipFlop(fmt.Sprintf("%s-RSFlipFlop", name), latch.clrOR, latch.sAnd) // if remove wires (for pause), this line makes the correct latch

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
func NewNBitLevelTriggeredDTypeLatchWithClear(name string, clrPin, clkInPin pwrEmitter, dataInPins []pwrEmitter) *NBitLevelTriggeredDTypeLatchWithClear {
	latch := &NBitLevelTriggeredDTypeLatchWithClear{}

	for i, dataInPin := range dataInPins {
		latch.latches = append(latch.latches, NewLevelTriggeredDTypeLatchWithClear(fmt.Sprintf("%s-Latches[%d]", name, i), clrPin, clkInPin, dataInPin))

		// refer to the inner-latches's Qs output for easier external access
		latch.Qs = append(latch.Qs, latch.latches[i].Q)
	}

	return latch
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each latch, to exit
func (l *NBitLevelTriggeredDTypeLatchWithClear) Shutdown() {
	for _, l := range l.latches {
		l.Shutdown()
	}
}

// Edge-triggered D-Type Latch is like a Level-triggered, but the ouputs only change when the clock goes from 0 to 1 ("Edge" = clock going high, "D" = data 0/1)
//
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

// NewEdgeTriggeredDTypeLatch returns an EdgeTriggeredDTypeLatch component controlled by a Clock pin, which will control how the Data pin is handled
//   (where Data will only get transferred to the output when the Clock transitions from 0 to 1)
func NewEdgeTriggeredDTypeLatch(name string, clkInPin, dataInPin pwrEmitter) *EdgeTriggeredDTypeLatch {
	latch := &EdgeTriggeredDTypeLatch{}

	// setup the left-side flipflop aspects
	latch.lRAnd = NewANDGate(fmt.Sprintf("%s-lRANDGate", name), NewInverter(fmt.Sprintf("%s-ClockInverter-lRANDGate", name), clkInPin), dataInPin)
	latch.lSAnd = NewANDGate(fmt.Sprintf("%s-lSANDGate", name), NewInverter(fmt.Sprintf("%s-ClockInverter-lSANDGate", name), clkInPin), NewInverter(fmt.Sprintf("%s-DataInverter-lSANDGate", name), dataInPin))
	latch.lRS = NewRSFlipFlop(fmt.Sprintf("%s-lRSFlipFlop", name), latch.lRAnd, latch.lSAnd)

	// setup the right-side flipflop aspects
	latch.rRAnd = NewANDGate(fmt.Sprintf("%s-rRANDGate", name), clkInPin, latch.lRS.Q)
	latch.rSAnd = NewANDGate(fmt.Sprintf("%s-rSANDGate", name), clkInPin, latch.lRS.QBar)
	latch.rRS = NewRSFlipFlop(fmt.Sprintf("%s-rRSFlipFlop", name), latch.rRAnd, latch.rSAnd)

	// refer to the inner-right-flipflop's outputs for easier external access
	latch.Q = latch.rRS.Q
	latch.QBar = latch.rRS.QBar

	return latch
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (l *EdgeTriggeredDTypeLatch) Shutdown() {
	l.rRS.Shutdown()
	l.rSAnd.Shutdown()
	l.rRAnd.Shutdown()
	l.lRS.Shutdown()
	l.lSAnd.Shutdown()
	l.lRAnd.Shutdown()
}

// FrequencyDivider is a special type of EdgeTriggeredDTypeLatch whose Clock pin is controlled by an Oscillator and whose Data pin is fed from its own QBar output
//    Every time the clock oscillates to 1, the Q/QBar outputs will output their state change, where Q is the FrequencyDivider's output and QBar feeds back to flip the value of Q for the next clock oscillation
//    On its own (though fragile), it basically acts like an oscillator itself, running at half the rate of the Clock pin's oscillator input (Sooooo, we should be able to chain this component with more to make a ripple counter)
type FrequencyDivider struct {
	wireLoopBack *Wire
	latch        *EdgeTriggeredDTypeLatch
	Q            *NORGate
	QBar         *NORGate
}

// NewFrequencyDivider returns a FrequencyDivider controlled by the passed in Oscillator
func NewFrequencyDivider(name string, oscillator pwrEmitter) *FrequencyDivider {
	freqDiv := &FrequencyDivider{}

	freqDiv.wireLoopBack = NewWire(fmt.Sprintf("%s-wireQBarLoopBack", name), 0)
	freqDiv.latch = NewEdgeTriggeredDTypeLatch(fmt.Sprintf("%s-Latch", name), oscillator, freqDiv.wireLoopBack)
	freqDiv.latch.QBar.WireUp(freqDiv.wireLoopBack.Input)

	// refer to the inner-right-flipflop's outputs for easier external access
	freqDiv.Q = freqDiv.latch.Q
	freqDiv.QBar = freqDiv.latch.QBar

	return freqDiv
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (d *FrequencyDivider) Shutdown() {
	d.latch.Shutdown()
	d.wireLoopBack.Shutdown()
}

// NBitRippleCounter is a chain of N number of frequency dividers where each divider's clock will be controlled by the left-side neighboring divider's QBar out
//   This should allow a rudimentary binary counter operation to occur, ticking at the rate of the outside driver clock
type NBitRippleCounter struct {
	freqDivs []*FrequencyDivider
	Qs       []*NORGate
}

// NewNBitRippleCounter returns an NBitRippleCounter which will user the oscillator pin as the driving counter rate to control a string of chained frequency dividers (the width of the size input)
func NewNBitRippleCounter(name string, oscillator pwrEmitter, size int) *NBitRippleCounter {
	counter := &NBitRippleCounter{}

	for i := size - 1; i >= 0; i-- {
		var freqDiv *FrequencyDivider

		if i == size-1 {
			freqDiv = NewFrequencyDivider(fmt.Sprintf("%s-Dividers[%d]", name, i), oscillator) // setup the outer oscillation aspect
		} else {
			freqDiv = NewFrequencyDivider(fmt.Sprintf("%s-Dividers[%d]", name, i), counter.freqDivs[0].QBar) // the rest chain together
		}

		// prepend since going in reverse order
		counter.freqDivs = append([]*FrequencyDivider{freqDiv}, counter.freqDivs...)

		// make Qs refer to each for easier external access (pre-pending here too)
		counter.Qs = append([]*NORGate{freqDiv.Q}, counter.Qs...)
	}

	return counter
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (c *NBitRippleCounter) Shutdown() {
	for _, d := range c.freqDivs {
		d.Shutdown()
	}
}

// Edge-triggered D-Type Latch with Preset and Clear is like an Edge-triggered D-Type Latch, but with an added Preset and Clear input to force the state to 1 or 0 regardless of clock/data
//
// pre clr d clk   q  !q
//  1   0  X  X    1  0   preset makes data and clock not matter, forces Q
//  0   1  X  X    0  1   clear makes data and clock not matter, forces QBar
//  0   0  0  ^    0  1	  should take the data value since clock was raised (transitioned to 1)
//  0   0  1  ^    1  0   should take the data value since clock was raised (transitioned to 1)
//  0   0  X  0    q  !q  data doesn't matter, no clock raise (to 1) transition to trigger a store-it action
type EdgeTriggeredDTypeLatchWithPresetAndClear struct {
	wireluQOut    *Wire
	wireluQBarOut *Wire
	luQ           *NORGate
	luQBar        *NORGate
	wirellQOut    *Wire
	wirellQBarOut *Wire
	llQ           *NORGate
	llQBar        *NORGate
	wireQOut      *Wire
	wireQBarOut   *Wire
	Q             *NORGate
	QBar          *NORGate
}

// NewEdgeTriggeredDTypeLatchWithPresetAndClear returns an EdgeTriggeredDTypeLatch component with an added Preset and Clear input
func NewEdgeTriggeredDTypeLatchWithPresetAndClear(name string, presetPin, clrPin, clkInPin, dataInPin pwrEmitter) *EdgeTriggeredDTypeLatchWithPresetAndClear {
	latch := &EdgeTriggeredDTypeLatchWithPresetAndClear{}

	latch.wireluQOut = NewWire(fmt.Sprintf("%s-LeftUpperQOutWire", name), 0)
	latch.wireluQBarOut = NewWire(fmt.Sprintf("%s-LeftUpperQBarOutWire", name), 0)
	latch.wirellQOut = NewWire(fmt.Sprintf("%s-LeftLowerQOutWire", name), 0)
	latch.wirellQBarOut = NewWire(fmt.Sprintf("%s-LeftLowerQBarOutWire", name), 0)
	latch.wireQOut = NewWire(fmt.Sprintf("%s-RightQOutWire", name), 0)
	latch.wireQBarOut = NewWire(fmt.Sprintf("%s-RightQBarOutWire", name), 0)

	// setup the left-side upper flipflop aspects
	latch.luQ = NewNORGate(fmt.Sprintf("%s-LeftUpperQNORGate", name), clrPin, latch.wirellQBarOut, latch.wireluQBarOut)
	latch.luQ.WireUp(latch.wireluQOut.Input)

	latch.luQBar = NewNORGate(fmt.Sprintf("%s-LeftUpperQBarNORGate", name), latch.luQ, presetPin, NewInverter(fmt.Sprintf("%s-ClockInverter-LeftUpperQBarNORGate", name), clkInPin))
	latch.luQBar.WireUp(latch.wireluQBarOut.Input)

	// setup the left-side lower flipflop aspects
	latch.llQ = NewNORGate(fmt.Sprintf("%s-LeftLowerQNORGate", name), latch.luQBar, NewInverter(fmt.Sprintf("%s-ClockInverter-LeftLowerQNORGate", name), clkInPin), latch.wirellQBarOut)
	latch.llQ.WireUp(latch.wirellQOut.Input)

	latch.llQBar = NewNORGate(fmt.Sprintf("%s-LeftLowerQBarNORGate", name), latch.llQ, presetPin, dataInPin)
	latch.llQBar.WireUp(latch.wirellQBarOut.Input)

	// setup the right-side flipflop aspects
	latch.Q = NewNORGate(fmt.Sprintf("%s-RightQNORGate", name), clrPin, latch.luQBar, latch.wireQBarOut)
	latch.Q.WireUp(latch.wireQOut.Input)

	latch.QBar = NewNORGate(fmt.Sprintf("%s-RightQBarNORGate", name), latch.wireQOut, presetPin, latch.llQBar)
	latch.QBar.WireUp(latch.wireQBarOut.Input)

	return latch
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (l *EdgeTriggeredDTypeLatchWithPresetAndClear) Shutdown() {
		l.QBar.Shutdown()
		l.Q.Shutdown()
		l.llQBar.Shutdown()
		l.llQ.Shutdown()
		l.luQBar.Shutdown()
		l.luQ.Shutdown()
		l.wireQBarOut.Shutdown()
		l.wireQOut.Shutdown()
		l.wirellQBarOut.Shutdown()
		l.wirellQOut.Shutdown()
		l.wireluQBarOut.Shutdown()
		l.wireluQOut.Shutdown()
}
