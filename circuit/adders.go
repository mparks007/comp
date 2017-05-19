package circuit

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
)

// Half Adder
// A and B in result in Sum and Carry out (doesn't handle carry in, needs full-adder)
// A B		Sum		Carry
// 0 0		0		0
// 1 0		1		0
// 0 1		1		0
// 1 1		0		1

type halfAdder struct {
	sum   emitter
	carry emitter
}

func newHalfAdder(pin1, pin2 emitter) *halfAdder {
	return &halfAdder{
		newXORGate(pin1, pin2),
		newANDGate(pin1, pin2),
	}
}

// Full Adder
// A, B, and Carry in result in Sum and Carry out (handles 1 + 1 + 'carried over + 1')

type fullAdder struct {
	halfAdder1 *halfAdder
	halfAdder2 *halfAdder
	sum        emitter
	carry      emitter
}

func newFullAdder(pin1, pin2, carryIn emitter) *fullAdder {
	f := &fullAdder{}

	f.halfAdder1 = newHalfAdder(pin1, pin2)
	f.halfAdder2 = newHalfAdder(f.halfAdder1.sum, carryIn)
	f.sum = f.halfAdder2.sum
	f.carry = newORGate(f.halfAdder1.carry, f.halfAdder2.carry)

	return f
}

// 8-bit Adder
// Handles a Carry bit in and holds potential Carry bit after summing all
//    10011101
// +  11010110
// = 101110011

type EightbitAdder struct {
	fullAdders [8]*fullAdder
	carryOut   emitter
}

func NewEightBitAdder(byte1, byte2 string, carryIn emitter) (*EightbitAdder, error) {
	match, err := regexp.MatchString("^[01]{8}$", byte1)
	if err != nil {
		return nil, err
	}
	if !match {
		err = errors.New(fmt.Sprint("First input not in 8-bit binary format: " + byte1))
		return nil, err
	}

	match, err = regexp.MatchString("^[01]{8}$", byte2)
	if err != nil {
		return nil, err
	}

	if !match {
		err = errors.New(fmt.Sprint("Second input not in 8-bit binary format: " + byte2))
		return nil, err
	}

	a := &EightbitAdder{}

	for i := 7; i >= 0; i-- {
		var f *fullAdder
		var pin1 emitter
		var pin2 emitter

		switch byte1[i] {
		case '0':
			pin1 = nil
		case '1':
			pin1 = &battery{}
		}

		switch byte2[i] {
		case '0':
			pin2 = nil
		case '1':
			pin2 = &battery{}
		}

		if i == 7 {
			f = newFullAdder(pin1, pin2, carryIn)
		} else {
			f = newFullAdder(pin1, pin2, a.fullAdders[i+1].carry) // carry-in is the neighboring adders carry-out
		}

		a.fullAdders[i] = f
	}

	a.carryOut = a.fullAdders[0].carry

	return a, nil
}

func (a *EightbitAdder) String() string {
	answer := ""

	if a.carryOut.Emitting() {
		answer += "1"
	}

	for _, v := range a.fullAdders {

		if v.sum.Emitting() {
			answer += "1"
		} else {
			answer += "0"
		}
	}

	return answer
}

// 16-bit Adder
// Handles a Carry bit in from the far right, chains the two, inner half-adder on a Carry, and holds potential Carry bit after summing all 16 bits
//    1001110110011101
// +  1101011011010110
// = 10111010001110011

type SixteenBitAdder struct {
	rightAdder *EightbitAdder
	leftAdder  *EightbitAdder
	carryOut   emitter
}

func NewSixteenBitAdder(bytes1, bytes2 string, carryIn emitter) (*SixteenBitAdder, error) {
	match, err := regexp.MatchString("^[01]{16}$", bytes1)
	if err != nil {
		return nil, err
	}
	if !match {
		err = errors.New(fmt.Sprint("First input not in 16-bit binary format: " + bytes1))
		return nil, err
	}

	match, err = regexp.MatchString("^[01]{16}$", bytes2)
	if err != nil {
		return nil, err
	}

	if !match {
		err = errors.New(fmt.Sprint("Second input not in 16-bit binary format: " + bytes2))
		return nil, err
	}

	a := &SixteenBitAdder{}

	a.rightAdder, err = NewEightBitAdder(bytes1[8:], bytes2[8:], carryIn)
	if err != nil {
		return nil, err
	}

	a.leftAdder, err = NewEightBitAdder(bytes1[:8], bytes2[:8], a.rightAdder.carryOut)
	if err != nil {
		return nil, err
	}

	a.carryOut = a.leftAdder.carryOut

	return a, err
}

func (a *SixteenBitAdder) String() string {
	answerCarry := ""
	answerLeft := ""
	answerRight := ""

	if a.leftAdder.carryOut.Emitting() {
		answerCarry += "1"
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, v := range a.leftAdder.fullAdders {

			if v.sum.Emitting() {
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

			if v.sum.Emitting() {
				answerRight += "1"
			} else {
				answerRight += "0"
			}
		}
	}()

	wg.Wait()

	return answerCarry + answerLeft + answerRight
}
