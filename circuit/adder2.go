package circuit

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

func NewEightBitAdder2(minuendBits, subtrahendBits [8]bitPublisher, carryIn bitPublisher) *EightBitAdder2 {

	a := &EightBitAdder2{}

	for i := 7; i >= 0; i-- {
		var f *FullAdder2

		if i == 7 {
			f = NewFullAdder2(minuendBits[i], subtrahendBits[i], carryIn)
		} else {
			f = NewFullAdder2(minuendBits[i], subtrahendBits[i], a.fullAdders[i+1].Carry) // carry-in is the neighboring adders carry-out
		}

		a.Sums[i] = f.Sum

		a.fullAdders[i] = f
	}

	a.CarryOut = a.fullAdders[0].Carry

	return a
}

func (a *EightBitAdder2) AsString() string {
	answer := ""

	if a.CarryOut.(bitPublication).isPowered {
		answer += "1"
	}

	for _, v := range a.fullAdders {

		if v.Sum.(bitPublication).isPowered {
			answer += "1"
		} else {
			answer += "0"
		}
	}

	return answer
}
