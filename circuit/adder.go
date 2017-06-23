package circuit

import (
	"errors"
	"fmt"
)

// Half NBitAdder
// A and B in result in Sum and Carry out (doesn't handle carry in, needs full-adder)
// A B		Sum		Carry
// 0 0		0		0
// 1 0		1		0
// 0 1		1		0
// 1 1		0		1

type HalfAdder struct {
	Sum   pwrEmitter
	Carry pwrEmitter
}

func NewHalfAdder(pin1, pin2 pwrEmitter) *HalfAdder {
	h := &HalfAdder{}

	h.Sum = NewXORGate(pin1, pin2)
	h.Carry = NewANDGate(pin1, pin2)

	return h
}

// Full NBitAdder
// A, B, and Carry in result in Sum and Carry out (can handle "1 + 1 = 0 carry the 1")

type FullAdder struct {
	halfAdder1 *HalfAdder
	halfAdder2 *HalfAdder
	Sum        pwrEmitter
	Carry      pwrEmitter
}

func NewFullAdder(pin1, pin2, carryInPin pwrEmitter) *FullAdder {
	f := &FullAdder{}

	f.halfAdder1 = NewHalfAdder(pin1, pin2)
	f.halfAdder2 = NewHalfAdder(f.halfAdder1.Sum, carryInPin)
	f.Sum = f.halfAdder2.Sum
	f.Carry = NewORGate(f.halfAdder1.Carry, f.halfAdder2.Carry)

	return f
}

// N-bit NAdder
// Handles s Carry bit in and holds potential Carry bit after summing all
//    10011101
// +  11010110
// = 101110011

type NBitAdder struct {
	fullAdders []*FullAdder
	Sums       []pwrEmitter
	CarryOut   pwrEmitter
}

func NewNBitAdder(addend1Pins, addend2Pins []pwrEmitter, carryInPin pwrEmitter) (*NBitAdder, error) {

	if len(addend1Pins) != len(addend2Pins) {
		return nil, errors.New(fmt.Sprintf("Mismatched addend lengths.  Addend1 len: %d, Addend2 len: %d", len(addend1Pins), len(addend2Pins)))
	}

	addr := &NBitAdder{}

	for i := len(addend1Pins) - 1; i >= 0; i-- {
		var full *FullAdder

		if i == len(addend1Pins)-1 {
			full = NewFullAdder(addend1Pins[i], addend2Pins[i], carryInPin) // carry-in is the actual (potential) carry from an adjoining circuit
		} else {
			// [carry-in is the neighboring adders carry-out]
			// since insert at the front of the slice, the neighbor is always the one at the front per the prior insert
			full = NewFullAdder(addend1Pins[i], addend2Pins[i], addr.fullAdders[0].Carry)
		}

		// prepend since going in reverse order
		addr.fullAdders = append([]*FullAdder{full}, addr.fullAdders...)

		// make Sums refer to each for easier external access (pre-pending here)
		addr.Sums = append([]pwrEmitter{full.Sum}, addr.Sums...)
	}

	// make CarryOut refer to the appropriate adder for easier external access
	addr.CarryOut = addr.fullAdders[0].Carry

	return addr, nil
}

func (a *NBitAdder) AsAnswerString() string {
	answer := ""

	for _, full := range a.fullAdders {

		if full.Sum.(*XORGate).GetIsPowered() {
			answer += "1"
		} else {
			answer += "0"
		}
	}

	return answer
}

func (a *NBitAdder) CarryOutAsBool() bool {
	return a.CarryOut.(*ORGate).GetIsPowered()
}

type ThreeNumberAdder struct {
	latchStore    *NBitLatch
	selector      *TwoToOneSelector
	adder         *NBitAdder
	SaveToLatch   *Switch
	ReadFromLatch *Switch
	Sums          []pwrEmitter
	CarryOut      pwrEmitter
}

func NewThreeNumberAdder(aSwitchBank, bSwitchBank *NSwitchBank) (*ThreeNumberAdder, error) {

	if len(aSwitchBank.Switches) != len(bSwitchBank.Switches) {
		return nil, errors.New(fmt.Sprintf("Mismatched input lengths. Switchbank 1 switch count: %d, Switchbank 2 switch count: %d", len(aSwitchBank.Switches), len(bSwitchBank.Switches)))
	}

	addr := &ThreeNumberAdder{}

	// build the selector
	addr.ReadFromLatch = NewSwitch(false)
	addr.selector, _ = NewTwoToOneSelector(addr.ReadFromLatch, bSwitchBank.AsPwrEmitters(), nil) // we don't have a latch store yet so cannot set bPins

	// build the adder, handing it the selector for the B pins
	addr.adder, _ = NewNBitAdder(aSwitchBank.AsPwrEmitters(), addr.selector.Outs, nil)

	// build the latch, handing it the adder for its input pins (for the loop)
	addr.SaveToLatch = NewSwitch(false)
	addr.latchStore = NewNBitLatch(addr.SaveToLatch, addr.adder.Sums)

	// now refresh the selectors b inputs with the latch store's output
	addr.selector.UpdateBPins(addr.latchStore.Qs)

	// refer to the appropriate adder innards for easier external access
	addr.Sums = addr.adder.Sums
	addr.CarryOut = addr.adder.CarryOut

	return addr, nil
}

func (a *ThreeNumberAdder) AsAnswerString() string {
	return a.adder.AsAnswerString()
}

func (a *ThreeNumberAdder) CarryOutAsBool() bool {
	return a.adder.CarryOutAsBool()
}

type NNumberAdder struct {
	latchStore *NBitLatchWithClear
	adder      *NBitAdder
	Clear      *Switch
	Add        *Switch
	Sums       []pwrEmitter
	CarryOut   pwrEmitter
}

func NewNNumberAdder(switchBank *NSwitchBank) (*NNumberAdder, error) {

	addr := &NNumberAdder{}

	// build the latch
	addr.Clear = NewSwitch(false)
	addr.Add = NewSwitch(false)
	addr.latchStore = NewNBitLatchWithClear(addr.Clear, addr.Add, make([]pwrEmitter, len(switchBank.Switches))) // we don't have an adder yet so cannot set data pins correctly yet

	// build the adder, handing it the selector for the B pins
	addr.adder, _ = NewNBitAdder(switchBank.AsPwrEmitters(), addr.latchStore.Qs, nil)

	// now refresh the latch's inputs with the adder's output
	addr.latchStore.UpdatePins(addr.adder.Sums)

	// refer to the appropriate adder innards for easier external access
	addr.Sums = addr.latchStore.Qs

	return addr, nil
}

func (a *NNumberAdder) AsAnswerString() string {
	answer := ""

	for _, q := range a.latchStore.Qs {

		if q.(*NORGate).GetIsPowered() {
			answer += "1"
		} else {
			answer += "0"
		}
	}

	return answer
}
