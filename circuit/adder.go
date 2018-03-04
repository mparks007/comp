package circuit

import "fmt"

// HalfAdder is a standard Half Adder circuit
//	A and B inputs result in Sum and Carry outputs (as normal, this circuit doesn't handle a carry in (needs FullAdder for that)
//
// Truth Table
// A B		Sum		Carry
// 0 0		0		0
// 1 0		1		0
// 0 1		1		0
// 1 1		0		1
type HalfAdder struct {
	Sum   *XORGate
	Carry *ANDGate
}

// NewHalfAdder returns a HalfAdder which can add up the values of the two input pins
func NewHalfAdder(name string, pin1, pin2 pwrEmitter) *HalfAdder {
	h := &HalfAdder{}

	h.Sum = NewXORGate(fmt.Sprintf("%s-XORGate", name), pin1, pin2)
	h.Carry = NewANDGate(fmt.Sprintf("%s-ANDGate", name), pin1, pin2)

	return h
}

// Shutdown will allow the go funcs, which are handling listen/transmit on the inner gates, to exit
func (h *HalfAdder) Shutdown() {
	h.Carry.Shutdown()
	h.Sum.Shutdown()
}

// FullAdder is a standard Full Adder circuit
// 	A, B, and Carry in result in Sum and Carry out (can handle "1 + 1 = 0 carry the 1"!)
//
// Truth Table
// A B CarrIn  Sum	Carry
// 0 0	 0		0     0
// 0 0	 1		1     0
// 1 0	 0		1     0
// 1 0	 1		0     1
// 0 1	 0		1     0
// 0 1	 1		0     1
// 1 1	 0		0     1
// 1 1	 1		1     1
type FullAdder struct {
	halfAdder1 *HalfAdder
	halfAdder2 *HalfAdder
	Sum        *XORGate
	Carry      *ORGate
}

// NewFullAdder returns a FullAdder which can add up the values of the two input pins, but also accepts a carry-in pin to include in the addition
func NewFullAdder(name string, pin1, pin2, carryInPin pwrEmitter) *FullAdder {
	f := &FullAdder{}

	f.halfAdder1 = NewHalfAdder(fmt.Sprintf("%s-LeftHalfAdder", name), pin1, pin2)
	f.halfAdder2 = NewHalfAdder(fmt.Sprintf("%s-RightHalfAdder", name), f.halfAdder1.Sum, carryInPin)
	f.Sum = f.halfAdder2.Sum
	f.Carry = NewORGate(fmt.Sprintf("%s-ORGate", name), f.halfAdder1.Carry, f.halfAdder2.Carry)

	return f
}

// Shutdown will allow the go funcs, which are handling listen/transmit on the inner components, to exit
func (f *FullAdder) Shutdown() {
	f.Carry.Shutdown()
	f.halfAdder2.Shutdown()
	f.halfAdder1.Shutdown()
}
/*
// NBitAdder allows the summing of two binary numbers
//	Handles a Carry bit in and holds potential Carry bit after summing all
//
//    10011101
// +  11010110
// = 101110011
type NBitAdder struct {
	fullAdders []*FullAdder
	Sums       []pwrEmitter
	CarryOut   *ORGate
}

// NewNBitAdder returns a NBitAdder which will add up the values of the two sets of input pins, but also accepts a carry-in pin to include in the addition
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
			// [carry-in (third param in NewFullAdder) is the neighboring, less significant adder's carry-out]
			//   since inserting at the front of the slice, the neighbor is always the one at the current front per the prior insert
			full = NewFullAdder(addend1Pins[i], addend2Pins[i], addr.fullAdders[0].Carry)
		}

		// prepend since going in reverse order
		addr.fullAdders = append([]*FullAdder{full}, addr.fullAdders...)

		// us external Sums field for easier external access (pre-pending here too)
		addr.Sums = append([]pwrEmitter{full.Sum}, addr.Sums...)
	}

	// make external CarryOut field refer to the appropriate (most significant) adder for easier external access
	addr.CarryOut = addr.fullAdders[0].Carry

	return addr, nil
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each full adder, to exit
func (a *NBitAdder) Shutdown() {
	for i := range a.fullAdders {
		a.fullAdders[i].Shutdown()
	}
}

// ThreeNumberAdder allows the summing of three binary numbers
type ThreeNumberAdder struct {
	latchStore    *NBitLevelTriggeredDTypeLatch
	selector      *TwoToOneSelector
	adder         *NBitAdder
	carryIn       *Switch
	loopRibbon    *RibbonCable
	SaveToLatch   *Switch
	ReadFromLatch *Switch
	Sums          []pwrEmitter
	CarryOut      *ORGate
}

// NewThreeNumberAdder returns a ThreeNumberAdder which will allow the addition of up to three sets of binary numbers
//
//	Steps:
//		1. On construction, send in the intial two numbers to add
//		2. Set SaveToLatch to true to store the answer from step 1
//		3. Set SaveToLatch back to false to prevent a chance for adding in a fourth number (which ThreeNumberAdder is not designed to do)
//		4. Update the first parameter inputs to be the third number to add in
//		5. Set ReadFromLatch to true to allow the saved original sum to be added to the new number entered in step 4
func NewThreeNumberAdder(aInputs, bInputs []pwrEmitter) (*ThreeNumberAdder, error) {

	if len(aInputs) != len(bInputs) {
		return nil, fmt.Errorf("Mismatched input lengths. Addend1 len: %d, Addend2 len: %d", len(aInputs), len(bInputs))
	}

	addr := &ThreeNumberAdder{}

	// set of wires that will lead from the adder outputs back up to the latch inputs
	addr.loopRibbon = NewRibbonCable(uint(len(aInputs)), 10)

	// build the latch, handing it the wires from the adder output
	addr.SaveToLatch = NewSwitch(false)
	addr.latchStore = NewNBitLevelTriggeredDTypeLatch(addr.SaveToLatch, addr.loopRibbon.Wires)

	// build the selector
	addr.ReadFromLatch = NewSwitch(false)
	addr.selector, _ = NewTwoToOneSelector(addr.ReadFromLatch, bInputs, addr.latchStore.Qs)

	// build the adder, handing it the selector for the B pins
	addr.carryIn = NewSwitch(false)
	addr.adder, _ = NewNBitAdder(aInputs, addr.selector.Outs, addr.carryIn)

	// set adder sums to be the input to the loopback ribbon cable
	addr.loopRibbon.SetInputs(addr.adder.Sums...)

	// refer to the appropriate adder innards for easier external access
	addr.Sums = addr.adder.Sums
	addr.CarryOut = addr.adder.CarryOut

	return addr, nil
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (a *ThreeNumberAdder) Shutdown() {
	a.adder.Shutdown()
	a.carryIn.Shutdown()
	a.selector.Shutdown()
	a.ReadFromLatch.Shutdown()
	a.latchStore.Shutdown()
	a.SaveToLatch.Shutdown()
	a.loopRibbon.Shutdown()
}

// NNumberAdder allows the summing of any number of binary numbers
type NNumberAdder struct {
	latches    *NBitLevelTriggeredDTypeLatchWithClear
	adder      *NBitAdder
	loopRibbon *RibbonCable
	Clear      *Switch
	Add        *Switch
	Sums       []pwrEmitter
}

// NewNNumberAdder returns an NNumberAdder which will allow the addition of any number of binary numbers
//	However, due to the answer latch being level-triggered and not edge-triggered, once active, it will kick off an infinite loop of addition.
//	This component is just a proof of concept on how to add multiple numbers without needing an internal 2-to-1 Selector
//
//	Steps:
//		1. On construction, send in the intial numbers
//		2. TODO: finish the steps once I have the TestNewNNumberAdder unit test working
//		3. TODO:
//		4. TODO: ...
func NewNNumberAdder(inputs []pwrEmitter) (*NNumberAdder, error) {

	addr := &NNumberAdder{}

	addr.Clear = NewSwitch(false)
	addr.Add = NewSwitch(false)

	addr.loopRibbon = NewRibbonCable(uint(len(inputs)), 10)

	addr.adder, _ = NewNBitAdder(inputs, addr.loopRibbon.Wires, nil)

	addr.latches = NewNBitLevelTriggeredDTypeLatchWithClear(addr.Clear, addr.Add, addr.adder.Sums)

	addr.loopRibbon.SetInputs(addr.latches.Qs...)

	// refer to the appropriate adder innards for easier external access
	addr.Sums = addr.latches.Qs

	return addr, nil
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (a *NNumberAdder) Shutdown() {
	a.latches.Shutdown()
	a.adder.Shutdown()
	a.loopRibbon.Shutdown()
	a.Add.Shutdown()
	a.Clear.Shutdown()
}
*/
