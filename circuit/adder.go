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

func NewFullAdder(pin1, pin2, carryIn pwrEmitter) *FullAdder {
	f := &FullAdder{}

	f.halfAdder1 = NewHalfAdder(pin1, pin2)
	f.halfAdder2 = NewHalfAdder(f.halfAdder1.Sum, carryIn)
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

func NewNBitAdder(addend1Pins, addend2Pins []pwrEmitter, carryIn pwrEmitter) (*NBitAdder, error) {

	if len(addend1Pins) != len(addend2Pins) {
		return nil, errors.New(fmt.Sprintf("Mismatched addend lengths.  Addend1 len: %d, Addend2 len: %d", len(addend1Pins), len(addend2Pins)))
	}

	addr := &NBitAdder{}

	for i := len(addend1Pins) - 1; i >= 0; i-- {
		var full *FullAdder

		if i == len(addend1Pins)-1 {
			full = NewFullAdder(addend1Pins[i], addend2Pins[i], carryIn) // carry-in is the actual (potential) carry from an adjoining circuit
		} else {
			// [carry-in is the neighboring adders carry-out]
			// since insert at the front of the slice, the neigher is always the one at the front per the prior insert
			full = NewFullAdder(addend1Pins[i], addend2Pins[i], addr.fullAdders[0].Carry)
		}

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
