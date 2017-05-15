package circuit

import "testing"

/**********************************/
/*************** Pin **************/
/**********************************/

func TestPin_InPin_NoPower(t *testing.T) {
	p := inPin{}

	want := false
	got := p.Emitting()
	if got != want {
		t.Error("Expected an input pin, with no power supplied, to return no power (off)")
	}
}

func TestPin_InPin_HasPower(t *testing.T) {
	p := inPin{&battery{}}

	want := true
	got := p.Emitting()
	if got != want {
		t.Error("Expected an input pin, with power supplied, to return power (on)")
	}
}

func TestPin_OutOpenPin_NoPower(t *testing.T) {
	p := outOpenPin{}

	want := false
	got := p.Emitting()
	if got != want {
		t.Error("Expected an out (open) pin, with no power supplied at either source pin, to return no power (off)")
	}
}

func TestPin_OutOpenPin_SourceAPowerOnly_HasPower(t *testing.T) {
	p := outOpenPin{&battery{}, nil}

	want := true
	got := p.Emitting()
	if got != want {
		t.Error("Expected an out (open) pin, with power supplied only at source pin A, to return power (on)")
	}
}

func TestPin_OutOpenPin_SourceBPowerOnly_NoPower(t *testing.T) {
	p := outOpenPin{nil, &battery{}}

	want := false
	got := p.Emitting()
	if got != want {
		t.Error("Expected an out (open) pin, with power supplied only at source pin B, to return no power (off)")
	}
}

func TestPin_OutOpenPin_BothSourcesPowered_NoPower(t *testing.T) {
	p := outOpenPin{&battery{}, &battery{}}

	want := false
	got := p.Emitting()
	if got != want {
		t.Error("Expected an out (open) pin, with power supplied to both source pins, to return no power (off)")
	}
}

func TestPin_OutClosedPin_NoPower(t *testing.T) {
	p := outClosedPin{}

	want := false
	got := p.Emitting()
	if got != want {
		t.Error("Expected an out (closed) pin, with no power supplied at either source pin, to return no power (off)")
	}
}

func TestPin_OutClosedPin_SourceAPowerOnly_HasPower(t *testing.T) {
	p := outClosedPin{&battery{}, nil}

	want := false
	got := p.Emitting()
	if got != want {
		t.Error("Expected an out (closed) pin, with power supplied only at source pin A, to return no power (off)")
	}
}

func TestPin_OutClosedPin_SourceBPowerOnly_NoPower(t *testing.T) {
	p := outClosedPin{nil, &battery{}}

	want := false
	got := p.Emitting()
	if got != want {
		t.Error("Expected an out (closed) pin, with power supplied only at source pin B, to return no power (off)")
	}
}

func TestPin_OutClosedPin_BothSourcesPowered_NoPower(t *testing.T) {
	p := outClosedPin{&battery{}, &battery{}}

	want := true
	got := p.Emitting()
	if got != want {
		t.Error("Expected an out (closed) pin, with power supplied to both source pins, to return power (on)")
	}
}

/**********************************/
/************* Relay **************/
/**********************************/

func TestRelay_AInPowerOnly_HasOpenPowerOnly(t *testing.T) {
	r := newRelay(&battery{}, nil)

	want := true
	got := r.emittingOpen() && !r.emittingClosed()
	if got != want {
		t.Error("Expected a relay, with a battery only on AIn, to output power out the open position and only out that position.")
	}
}

func TestRelay_BInPowerOnly_HasNoPower(t *testing.T) {
	r := newRelay(nil, &battery{})

	want := false
	got := r.emittingOpen() || r.emittingClosed()
	if got != want {
		t.Error("Expected a relay, with a battery only on BIn, to output no power.")
	}
}

func TestRelay_AInAndBInPower_HasClosedPowerOnly(t *testing.T) {
	r := newRelay(&battery{}, &battery{})

	want := true
	got := r.emittingClosed() && !r.emittingOpen()
	if got != want {
		t.Error("Expected a relay, with a battery on AIn and BIn, to output power out the closed position and only out that position.")
	}
}

func TestRelay_AInAndBInNoPower_HasNoPower(t *testing.T) {
	r := &relay{}

	want := false
	got := r.emittingOpen() || r.emittingClosed()
	if got != want {
		t.Error("Expected a relay, with no input power, to output no power.")
	}
}

/**********************************/
/************** AND ***************/
/**********************************/

func TestANDGate_BothInputsUnpowered_NoPower(t *testing.T) {
	g := newANDGate(nil, nil)

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected an AND gate, with no powered inputs, to output no power (off).")
	}
}

func TestANDGate_OnlyFirstInputPowered_NoPower(t *testing.T) {
	g := newANDGate(&battery{}, nil)

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected an AND gate, with only its first input pin powered, to outuput no power (off).")
	}
}

func TestANDGate_OnlySecondInputPowered_NoPower(t *testing.T) {
	g := newANDGate(nil, &battery{})

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected an AND gate, with only its second input pin powered, to output no power (off).")
	}
}

func TestANDGate_BothInputsPowered_HasPower(t *testing.T) {
	g := newANDGate(&battery{}, &battery{})

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected an AND gate, with both its input pins powered, to output power (on).")
	}
}

/**********************************/
/*************  OR  ***************/
/**********************************/

func TestORGate_BothInputsUnpowered_NoPower(t *testing.T) {
	g := newORGate(nil, nil)

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected an OR gate, with no powered inputs, to output no power (off).")
	}
}

func TestORGate_OnlyFirstInputPowered_HasPower(t *testing.T) {
	g := newORGate(&battery{}, nil)

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected an OR gate, with only its first input pin powered, to output power (on).")
	}
}

func TestORGate_OnlySecondInputPowered_HasPower(t *testing.T) {
	g := newORGate(nil, &battery{})

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected an OR gate, with only its second input pin powered, to output power (on).")
	}
}

func TestORGate_BothInputsPowered_HasPower(t *testing.T) {
	g := newORGate(&battery{}, &battery{})

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected an OR gate, with both its input pins powered, to output power (on).")
	}
}

/**********************************/
/************* NAND ***************/
/**********************************/

func TestNANDGate_BothInputsUnpowered_HasPower(t *testing.T) {
	g := newNANDGate(nil, nil)

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected a NAND gate, with no powered inputs, to output power (on).")
	}
}

func TestNANDGate_OnlyFirstInputPowered_HasPower(t *testing.T) {
	g := newNANDGate(&battery{}, nil)

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected a NAND gate, with only its first input pin powered, to outuput power (on).")
	}
}

func TestNANDGate_OnlySecondInputPowered_HasPower(t *testing.T) {
	g := newNANDGate(nil, &battery{})

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected a NAND gate, with only its second input pin powered, to output power (on).")
	}
}

func TestNANDGate_BothInputsPowered_NoPower(t *testing.T) {
	g := newNANDGate(&battery{}, &battery{})

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected a NAND gate, with both its input pins powered, to output no power (off).")
	}
}

/**********************************/
/************** NOR ***************/
/**********************************/

func TestNORGate_BothInputsUnpowered_HasPower(t *testing.T) {
	g := newNORGate(nil, nil)

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected a NOR gate, with no powered inputs, to output power (on).")
	}
}

func TestNORGate_OnlyFirstInputPowered_NoPower(t *testing.T) {
	g := newNORGate(&battery{}, nil)

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected a NOR gate, with only its first input pin powered, to outuput no power (off).")
	}
}

func TestNORGate_OnlySecondInputPowered_NoPower(t *testing.T) {
	g := newNORGate(nil, &battery{})

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected a NOR gate, with only its second input pin powered, to output no power (off).")
	}
}

func TestNORGate_BothInputsPowered_NoPower(t *testing.T) {
	g := newNORGate(&battery{}, &battery{})

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected a NOR gate, with both its input pins powered, to output no power (off).")
	}
}

/**********************************/
/************** XOR ***************/
/**********************************/

func TestXORGate_BothInputsUnpowered_NoPower(t *testing.T) {
	g := newXORGate(nil, nil)

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected an XOR gate, with no powered inputs, to output no power (off).")
	}
}

func TestXORGate_OnlyFirstInputPowered_HasPower(t *testing.T) {
	g := newXORGate(&battery{}, nil)

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected an XOR gate, with only its first input pin powered, to outuput power (on).")
	}
}

func TestXORGate_OnlySecondInputPowered_HasPower(t *testing.T) {
	g := newXORGate(nil, &battery{})

	want := true
	got := g.Emitting()
	if got != want {
		t.Error("Expected an XOR gate, with only its second input pin powered, to output power (on).")
	}
}

func TestXORGate_BothInputsPowered_NoPower(t *testing.T) {
	g := newXORGate(&battery{}, &battery{})

	want := false
	got := g.Emitting()
	if got != want {
		t.Error("Expected an XOR gate, with both its input pins powered, to output no power (off).")
	}
}

/*********************************/
/********** HalfAdder ************/
/*********************************/

func TestHalfAdder_BothInputsUnPowered_NoSumNoCarry(t *testing.T) {
	h := newHalfAdder(nil, nil)

	want := false
	got := h.Sum()
	if got != want {
		t.Error("Expected a HalfAdder, with both its input pins unpowered, to return no sum.")
	}

	want = false
	got = h.Carry()
	if got != want {
		t.Error("Expected a HalfAdder, with both its input pins unpowered, to return no carry.")
	}
}

func TestHalfAdder_OnlyFirstInputsPowered_YesSumNoCarry(t *testing.T) {
	h := newHalfAdder(&battery{}, nil)

	want := true
	got := h.Sum()
	if got != want {
		t.Error("Expected a HalfAdder, with only its first input powered, to return a sum.")
	}

	want = false
	got = h.Carry()
	if got != want {
		t.Error("Expected a HalfAdder, with only its first input powered, to return no carry.")
	}
}

func TestHalfAdder_OnlySecondInputsPowered_YesSumNoCarry(t *testing.T) {
	h := newHalfAdder(nil, &battery{})

	want := true
	got := h.Sum()
	if got != want {
		t.Error("Expected a HalfAdder, with only its second input powered, to return a sum.")
	}

	want = false
	got = h.Carry()
	if got != want {
		t.Error("Expected a HalfAdder, with only its second input powered, to return no carry.")
	}
}

func TestHalfAdder_BothInputsAsOnes_NoSumYesCarry(t *testing.T) {
	h := newHalfAdder(&battery{}, &battery{})

	want := false
	got := h.Sum()
	if got != want {
		t.Error("Expected a HalfAdder, with both its input pins as ones, to return no sum.")
	}

	want = true
	got = h.Carry()
	if got != want {
		t.Error("Expected a HalfAdder, with both its input pins as ones, to return a carry.")
	}
}

/*********************************/
/********** HalfAdder ************/
/*********************************/

func TestFullAdder_AllInputsUnPowered_NoSumNoCarry(t *testing.T) {
	f := newFullAdder(nil, nil, nil)

	want := false
	got := f.Sum()
	if got != want {
		t.Error("Expected a FullAdder, with all its input pins unpowered, to return no sum.")
	}

	want = false
	got = f.Carry()
	if got != want {
		t.Error("Expected a FullAdder, with all its input pins unpowered, to return no carry.")
	}
}

func TestFullAdder_OnlyFirstInputPowered_YesSumNoCarry(t *testing.T) {
	f := newFullAdder(&battery{}, nil, nil)

	want := true
	got := f.Sum()
	if got != want {
		t.Error("Expected a FullAdder, with only its first input pins powered, to return a sum.")
	}

	want = false
	got = f.Carry()
	if got != want {
		t.Error("Expected a FullAdder, with only its first input pins powered, to return no carry.")
	}
}

func TestFullAdder_OnlySecondInputPowered_YesSumNoCarry(t *testing.T) {
	f := newFullAdder(nil, &battery{}, nil)

	want := true
	got := f.Sum()
	if got != want {
		t.Error("Expected a FullAdder, with only its first input pins powered, to return a sum.")
	}

	want = false
	got = f.Carry()
	if got != want {
		t.Error("Expected a FullAdder, with only its first input pins powered, to return no carry.")
	}
}

func TestFullAdder_OnlyCarryPowered_YesSumNoCarry(t *testing.T) {
	f := newFullAdder(nil, nil, &battery{})

	want := true
	got := f.Sum()
	if got != want {
		t.Error("Expected a FullAdder, with only its carry input pin powered, to return a sum.")
	}

	want = false
	got = f.Carry()
	if got != want {
		t.Error("Expected a FullAdder, with only its carry input pin powered, to return no carry.")
	}
}

func TestFullAdder_FirstInputAndCarryPowered_NoSumYesCarry(t *testing.T) {
	f := newFullAdder(&battery{}, nil, &battery{})

	want := false
	got := f.Sum()
	if got != want {
		t.Error("Expected a FullAdder, with its first input and carry pins powered, to return no sum.")
	}

	want = true
	got = f.Carry()
	if got != want {
		t.Error("Expected a FullAdder, with its first input and carry pins powered, to return a carry.")
	}
}

func TestFullAdder_OnlySecondInputAndCarryPowered_NoSumYesCarry(t *testing.T) {
	f := newFullAdder(nil, &battery{}, &battery{})

	want := false
	got := f.Sum()
	if got != want {
		t.Error("Expected a FullAdder, with its second input and carry pins powered, to return no sum.")
	}

	want = true
	got = f.Carry()
	if got != want {
		t.Error("Expected a FullAdder, with its second input and carry pins powered, to return carry.")
	}
}

func TestFullAdder_AllInputsPowered_YesSumYesCarry(t *testing.T) {
	f := newFullAdder(&battery{}, &battery{}, &battery{})

	want := true
	got := f.Sum()
	if got != want {
		t.Error("Expected a FullAdder, with all its input pins powered, to return a sum.")
	}

	want = true
	got = f.Carry()
	if got != want {
		t.Error("Expected a FullAdder, with all its input pins powered, to return a carry.")
	}
}

/*********************************/
/*********** bitsAdder ***********/
/*********************************/
/*
func TestBitAdder_ThreeBitAllZeros_ZeroSum(t *testing.T) {
	b := NewBitAdder("000", "000")

	want := "000"
	got := b.String()
	if got != want {
		t.Errorf("Adding three zeros to three zeros.  Wanted %v, got %v", want, got)
	}
}

func TestBitAdder_ThreeBitAllOnes_ZeroSum(t *testing.T) {
	b := NewBitAdder("111", "111")

	want := "1000"
	got := b.String()
	if got != want {
		t.Errorf("Adding three bits to three bits.  Wanted %v, got %v", want, got)
	}
}

//    0110011101
// +  1011010110
// = 10001110011

func TestBitAdder_TenBitNumbers_SumWithCarry(t *testing.T) {
	b := NewBitAdder("0110011101", "1011010110")

	want := "10001110011"
	got := b.String()
	if got != want {
		t.Errorf("Adding two, 10 bit numbers with final carry.  Wanted %v, got %v", want, got)
	}
}

//   0100011101
// + 1011010110
// = 1101110011

func TestBitAdder_TenBitNumbers_SumWithNoCarry(t *testing.T) {
	b := NewBitAdder("0100011101", "1011010110")

	want := "1101110011"
	got := b.String()
	if got != want {
		t.Errorf("Adding two, 10 bit numbers with no final carry.  Wanted %v, got %v", want, got)
	}
}
*/
