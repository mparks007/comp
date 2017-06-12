package circuit

import (
	"sync"
)

// Half Adder
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

// Full Adder
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

// 8-bit Adder
// Handles a Carry bit in and holds potential Carry bit after summing all
//    10011101
// +  11010110
// = 101110011

type EightBitAdder struct {
	fullAdders [8]*FullAdder
	Sums       [8]pwrEmitter
	CarryOut   pwrEmitter
}

func NewEightBitAdder(addend1Pins, addend2Pins [8]pwrEmitter, carryIn pwrEmitter) *EightBitAdder {

	a := &EightBitAdder{}

	for i := 7; i >= 0; i-- {
		var f *FullAdder

		if i == 7 {
			f = NewFullAdder(addend1Pins[i], addend2Pins[i], carryIn)
		} else {
			f = NewFullAdder(addend1Pins[i], addend2Pins[i], a.fullAdders[i+1].Carry) // carry-in is the neighboring adders carry-out
		}

		a.fullAdders[i] = f

		// make Sums refer to each for easier, external access
		a.Sums[i] = f.Sum
	}

	// make CarryOut refer to the appropriate adder for easier, external access
	a.CarryOut = a.fullAdders[0].Carry

	return a
}

func (a *EightBitAdder) AsAnswerString() string {
	answer := ""

	for _, v := range a.fullAdders {

		if v.Sum.(*XORGate).GetIsPowered() {
			answer += "1"
		} else {
			answer += "0"
		}
	}

	return answer
}

func (a *EightBitAdder) CarryOutAsBool() bool {
	return a.CarryOut.(*ORGate).GetIsPowered()
}

// 16-bit Adder
// Handles a Carry bit in from the far right, chains the two, inner half-adder switchOn a Carry, and holds potential Carry bit after summing all 16 bits
//    1001110110011101
// +  1101011011010110
// = 10111010001110011

type SixteenBitAdder struct {
	rightAdder *EightBitAdder
	leftAdder  *EightBitAdder
	Sums       [16]pwrEmitter
	CarryOut   pwrEmitter
}

func NewSixteenBitAdder(addend1Pins, addend2Pins [16]pwrEmitter, carryIn pwrEmitter) *SixteenBitAdder {

	a := &SixteenBitAdder{}

	// convert incoming 16-bit array to two 8-bit arrays
	var addend1Right [8]pwrEmitter
	var addend2Right [8]pwrEmitter
	copy(addend1Right[:], addend1Pins[8:])
	copy(addend2Right[:], addend2Pins[8:])

	var addend1Left [8]pwrEmitter
	var addend2Left [8]pwrEmitter
	copy(addend1Left[:], addend1Pins[:8])
	copy(addend2Left[:], addend2Pins[:8])

	a.rightAdder = NewEightBitAdder(addend1Right, addend2Right, carryIn)
	a.leftAdder = NewEightBitAdder(addend1Left, addend2Left, a.rightAdder.CarryOut)

	// make Sums refer to each for easier, external access
	for i, la := range a.leftAdder.Sums {
		a.Sums[i] = la
	}
	for i, ra := range a.rightAdder.Sums {
		a.Sums[i+8] = ra
	}

	// make CarryOut refer to the appropriate adder for easier, external access
	a.CarryOut = a.leftAdder.CarryOut

	return a
}

func (a *SixteenBitAdder) AsAnswerString() string {
	answerLeft := ""
	answerRight := ""

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, v := range a.leftAdder.fullAdders {

			if v.Sum.(*XORGate).GetIsPowered() {
				answerLeft += "1"
			} else {
				answerLeft += "0"
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, v := range a.rightAdder.fullAdders {

			if v.Sum.(*XORGate).GetIsPowered() {
				answerRight += "1"
			} else {
				answerRight += "0"
			}
		}
	}()

	wg.Wait()

	return answerLeft + answerRight
}

func (a *SixteenBitAdder) CarryOutAsBool() bool {
	return a.CarryOut.(*ORGate).GetIsPowered()
}
