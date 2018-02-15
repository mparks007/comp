package circuit

import (
	"errors"
	"fmt"
	"time"
)

// Half Adder
// A and B in result in Sum and Carry out (doesn't handle carry in, needs full-adder)
// A B		Sum		Carry
// 0 0		0		0
// 1 0		1		0
// 0 1		1		0
// 1 1		0		1

type HalfAdder struct {
	Sum   *XORGate
	Carry *ANDGate
}

func NewHalfAdder(pin1, pin2 pwrEmitter) *HalfAdder {
	h := &HalfAdder{}

	h.Sum = NewXORGate(pin1, pin2)
	h.Carry = NewANDGate(pin1, pin2)

	return h
}

// Full Adder
// A, B, and Carry in result in Sum and Carry out (can handle "1 + 1 = 0 carry the 1")

type FullAdder struct {
	halfAdder1 *HalfAdder
	halfAdder2 *HalfAdder
	Sum        *XORGate
	Carry      *ORGate
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
// Handles a Carry bit in and holds potential Carry bit after summing all
//    10011101
// +  11010110
// = 101110011

type NBitAdder struct {
	fullAdders []*FullAdder
	Sums       []pwrEmitter
	CarryOut   *ORGate
}

func NewNBitAdder(addend1Pins, addend2Pins []pwrEmitter, carryInPin pwrEmitter) (*NBitAdder, error) {

	if len(addend1Pins) != len(addend2Pins) {
		return nil, fmt.Errorf("Mismatched addend lengths.  Addend1 len: %d, Addend2 len: %d", len(addend1Pins), len(addend2Pins))
	}

	addr := &NBitAdder{}

	for i := len(addend1Pins) - 1; i >= 0; i-- {
		var full *FullAdder

		// if at least significant pin
		if i == len(addend1Pins)-1 {
			full = NewFullAdder(addend1Pins[i], addend2Pins[i], carryInPin) // carry-in is the actual (potential) carry from an adjoining circuit
		} else {
			// [carry-in is the neighboring, more significant adder's carry-out]
			// since insert at the front of the slice, the neighbor is always the one at the front per the prior insert
			full = NewFullAdder(addend1Pins[i], addend2Pins[i], addr.fullAdders[0].Carry)
		}

		// prepend since going in reverse order
		addr.fullAdders = append([]*FullAdder{full}, addr.fullAdders...)

		// us external Sums for easier external access (pre-pending here too)
		addr.Sums = append([]pwrEmitter{full.Sum}, addr.Sums...)
	}

	// make CarryOut refer to the appropriate (most significant) adder for easier external access
	addr.CarryOut = addr.fullAdders[0].Carry

	return addr, nil
}

type ThreeNumberAdder struct {
	latchStore    *NBitLatch
	selector      *TwoToOneSelector
	adder         *NBitAdder
	SaveToLatch   *Switch
	ReadFromLatch *Switch
	Sums          []pwrEmitter
	CarryOut      *ORGate
}

func NewThreeNumberAdder(aSwitchBank, bSwitchBank *NSwitchBank) (*ThreeNumberAdder, error) {

	if len(aSwitchBank.Switches) != len(bSwitchBank.Switches) {
		return nil, errors.New(fmt.Sprintf("Mismatched input lengths. Addend1 len: %d, Addend2 len: %d", len(aSwitchBank.Switches), len(bSwitchBank.Switches)))
	}

	addr := &ThreeNumberAdder{}

	// set of wires that will lead from the adder outputs back up to the latch inputs
	loopRibbon := NewRibbonCable(uint(len(aSwitchBank.Switches)), 10)

	// build the latch, handing it the wires from the adder output
	addr.SaveToLatch = NewSwitch(false)
	addr.latchStore = NewNBitLatch(addr.SaveToLatch, loopRibbon.Wires)

	// build the selector
	addr.ReadFromLatch = NewSwitch(false)
	addr.selector, _ = NewTwoToOneSelector(addr.ReadFromLatch, bSwitchBank.Switches, addr.latchStore.Qs)

	// build the adder, handing it the selector for the B pins
	addr.adder, _ = NewNBitAdder(aSwitchBank.Switches, addr.selector.Outs, NewSwitch(false)) // no carry-in

	// set adder sums to be the input to the loopback ribbon cable
	loopRibbon.SetInputs(addr.adder.Sums)

	// give the SetInputs inner go funcs time to spin up and get their states
	time.Sleep(time.Millisecond * 10)

	// refer to the appropriate adder innards for easier external access
	addr.Sums = addr.adder.Sums
	addr.CarryOut = addr.adder.CarryOut

	return addr, nil
}

type NNumberAdder struct {
	latches  *NBitLatchWithClear
	adder    *NBitAdder
	Clear    *Switch
	Add      *Switch
	Sums     []pwrEmitter
}

func NewNNumberAdder(switchBank *NSwitchBank) (*NNumberAdder, error) {

	addr := &NNumberAdder{}

	addr.Clear = NewSwitch(false)
	addr.Add = NewSwitch(false)

	loopRibbon := NewRibbonCable(uint(len(switchBank.Switches)), 10)

	addr.adder, _ = NewNBitAdder(switchBank.Switches, loopRibbon.Wires, nil)

	addr.latches = NewNBitLatchWithClear(addr.Clear, addr.Add, addr.adder.Sums)

	loopRibbon.SetInputs(addr.latches.Qs)

	// refer to the appropriate adder innards for easier external access
	addr.Sums = addr.latches.Qs

	return addr, nil
}
