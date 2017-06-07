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

type HalfAdder2 struct {
	Sum   bitPublisher
	Carry bitPublisher
}

func NewHalfAdder2(pin1, pin2 bitPublisher) *HalfAdder2 {
	h := &HalfAdder2{}

	h.Sum = NewXORGate2(pin1, pin2)
	h.Carry = NewANDGate2(pin1, pin2)

	return h
}

// Full Adder
// A, B, and Carry in result in Sum and Carry out (can handle "1 + 1 = 0 carry the 1")

type FullAdder2 struct {
	halfAdder1 *HalfAdder2
	halfAdder2 *HalfAdder2
	Sum        bitPublisher
	Carry      bitPublisher
}

func NewFullAdder2(pin1, pin2, carryIn bitPublisher) *FullAdder2 {
	f := &FullAdder2{}

	f.halfAdder1 = NewHalfAdder2(pin1, pin2)
	f.halfAdder2 = NewHalfAdder2(f.halfAdder1.Sum, carryIn)
	f.Sum = f.halfAdder2.Sum
	f.Carry = NewORGate2(f.halfAdder1.Carry, f.halfAdder2.Carry)

	return f
}

// 8-bit Adder
// Handles a Carry bit in and holds potential Carry bit after summing all
//    10011101
// +  11010110
// = 101110011

type EightBitAdder2 struct {
	fullAdders [8]*FullAdder2
	Sums       [8]bitPublisher
	CarryOut   bitPublisher
}

func NewEightBitAdder2(addend1Pins, addend2Pins [8]bitPublisher, carryIn bitPublisher) *EightBitAdder2 {

	a := &EightBitAdder2{}

	for i := 7; i >= 0; i-- {
		var f *FullAdder2

		if i == 7 {
			f = NewFullAdder2(addend1Pins[i], addend2Pins[i], carryIn)
		} else {
			f = NewFullAdder2(addend1Pins[i], addend2Pins[i], a.fullAdders[i+1].Carry) // carry-in is the neighboring adders carry-out
		}

		a.Sums[i] = f.Sum

		a.fullAdders[i] = f
	}

	a.CarryOut = a.fullAdders[0].Carry

	return a
}

func (a *EightBitAdder2) AsAnswerString() string {
	answer := ""

	for _, v := range a.fullAdders {

		if v.Sum.(*XORGate2).isPowered {
			answer += "1"
		} else {
			answer += "0"
		}
	}

	return answer
}

func (a *EightBitAdder2) CarryOutAsBool() bool {
	return a.CarryOut.(*ORGate2).isPowered
}

// 16-bit Adder
// Handles a Carry bit in from the far right, chains the two, inner half-adder switchOn a Carry, and holds potential Carry bit after summing all 16 bits
//    1001110110011101
// +  1101011011010110
// = 10111010001110011

type SixteenBitAdder2 struct {
	rightAdder *EightBitAdder2
	leftAdder  *EightBitAdder2
	Sums       [16]bitPublisher
	CarryOut   bitPublisher
}

func NewSixteenBitAdder2(addend1Pins, addend2Pins [16]bitPublisher, carryIn bitPublisher) *SixteenBitAdder2 {

	a := &SixteenBitAdder2{}

	var addend1Right [8]bitPublisher
	var addend2Right [8]bitPublisher
	copy(addend1Right[:], addend1Pins[8:])
	copy(addend2Right[:], addend2Pins[8:])

	var addend1Left [8]bitPublisher
	var addend2Left [8]bitPublisher
	copy(addend1Left[:], addend1Pins[:8])
	copy(addend2Left[:], addend2Pins[:8])

	a.rightAdder = NewEightBitAdder2(addend1Right, addend2Right, carryIn)
	a.leftAdder = NewEightBitAdder2(addend1Left, addend2Left, a.rightAdder.CarryOut)

	for i, la := range a.leftAdder.Sums {
		a.Sums[i] = la
	}
	for i, ra := range a.rightAdder.Sums {
		a.Sums[i+8] = ra
	}

	a.CarryOut = a.leftAdder.CarryOut

	return a
}

func (a *SixteenBitAdder2) AsAnswerString() string {
	answerLeft := ""
	answerRight := ""

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, v := range a.leftAdder.fullAdders {

			if v.Sum.(*XORGate2).isPowered {
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

			if v.Sum.(*XORGate2).isPowered {
				answerRight += "1"
			} else {
				answerRight += "0"
			}
		}
	}()

	wg.Wait()

	return answerLeft + answerRight
}

func (a *SixteenBitAdder2) CarryOutAsBool() bool {
	return a.CarryOut.(*ORGate2).isPowered
}
