package circuit

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// go test
// go test -race -v (verbose)
// go test -race -cpu=1,2,4 (go max procs)
// go test -v
// go test -run TestOscillator (specific test)

func TestPwrSource(t *testing.T) {
	var want, got1, got2 bool

	pwr := &pwrSource{}

	// wire up two callbacks to prove both will get called
	pwr.WireUp(func(state bool) { got1 = state })
	pwr.WireUp(func(state bool) { got2 = state })

	pwr.Transmit(true)
	want = true

	if got1 != want {
		t.Errorf("Expected subscription 1 to be %t but got %t", want, got1)
	}

	if got2 != want {
		t.Errorf("Expected subscription 2 to be %t but got %t", want, got2)
	}

	pwr.Transmit(false)
	want = false

	if got1 != want {
		t.Errorf("Expected subscription 1 to be %t but got %t", want, got1)
	}

	if got2 != want {
		t.Errorf("Expected subscription 2 to be %t but got %t", want, got2)
	}
}

func TestBattery(t *testing.T) {
	var want, got bool

	bat := NewBattery()

	bat.WireUp(func(state bool) { got = state })

	// by default, a battery will be charged (on/true)
	want = true

	if got != want {
		t.Errorf("With a new battery, wanted the subscriber to see power as %t but got %t", want, got)
	}

	bat.Discharge()
	want = false

	if got != want {
		t.Errorf("With a discharged battery, wanted the subscriber'l IsPowered to be %t but got %t", want, got)
	}

	bat.Charge()
	want = true

	if got != want {
		t.Errorf("With a charged battery, wanted the subscriber'l IsPowered to be %t but got %t", want, got)
	}
}

func TestSwitch(t *testing.T) {
	var wantState, gotState bool
	var wantCount, gotCount int

	sw := NewSwitch(false)

	// register callback (will trigger immediate call to push isPowered at time of registration)
	sw.WireUp(func(state bool) {
		gotState = state
		gotCount += 1
	})

	// initial turn on
	sw.Set(true)
	wantState = true
	wantCount = 2

	if gotState != wantState {
		t.Errorf("With an off switch turned on, wanted the subscriber to see power as %t but got %t", wantState, gotState)
	}

	if gotCount != wantCount {
		t.Errorf("With an off switch turned on, wanted the subscriber call count to be %d but got %d", wantCount, gotCount)
	}

	// turn on again though already on
	sw.Set(true)
	wantCount = 2

	if gotCount != wantCount {
		t.Errorf("With an attempt to turn on an already on switch, wanted the subscriber'sw call count to remain %d but got %d", wantCount, gotCount)
	}

	// now off again
	sw.Set(false)
	wantState = false
	wantCount = 3

	if gotState != wantState {
		t.Errorf("With an on switch turned off, wanted the subscriber to see power as %t but got %t", wantState, gotState)
	}

	if gotCount != wantCount {
		t.Errorf("With an on switch turned off, wanted the subscriber call count to be %d but got %d", wantCount, gotCount)
	}
}

func TestNewNSwitchBank_BadInputs(t *testing.T) {
	testCases := []struct {
		input     string
		wantError string
	}{
		{"12345", "Input not in binary format: "},
		{"000X", "Input not in binary format: "},
		{"X000", "Input not in binary format: "},
		{"111X", "Input not in binary format: "},
		{"X111", "Input not in binary format: "},
		{"bad", "Input not in binary format: "},
		{"", "Input not in binary format: "},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting switches to %s", tc.input), func(t *testing.T) {
			sb, err := NewNSwitchBank(tc.input)

			if sb != nil {
				t.Error("Didn't expected a Switch Bank back but got one.")
			}

			tc.wantError += tc.input

			if err == nil || (err != nil && err.Error() != tc.wantError) {
				t.Errorf("Wanted error \"%s\" but got \"%v\"", tc.wantError, err)
			}
		})
	}
}

func TestNewNSwitchBank_GoodInputs(t *testing.T) {
	testCases := []struct {
		input string
		want  []bool
	}{
		{"0", []bool{false}},
		{"1", []bool{true}},
		{"101", []bool{true, false, true}},
		{"00000000", []bool{false, false, false, false, false, false, false, false}},
		{"11111111", []bool{true, true, true, true, true, true, true, true}},
		{"10101010", []bool{true, false, true, false, true, false, true, false}},
		{"10000001", []bool{true, false, false, false, false, false, false, true}},
		{"0000000000000000", []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
		{"1111111111111111", []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}},
		{"1010101010101010", []bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false}},
		{"1000000000000001", []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting switches to %s", tc.input), func(t *testing.T) {
			sb, err := NewNSwitchBank(tc.input)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
			}

			// try as actual switches
			for i, sw := range sb.Switches {
				got := sw.GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf("[As Switch] At index %d, wanted %v but got %v", i, want, got)
				}
			}

			// now try AsPwrEmitters
			for i, pwr := range sb.AsPwrEmitters() {
				got := pwr.(*Switch).GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf("[As PwrEmitter] At index %d, wanted %v but got %v", i, want, got)
				}
			}
		})
	}
}

func TestRelay_WithSwitches(t *testing.T) {
	testCases := []struct {
		aInPowered   bool
		bInPowered   bool
		wantAtOpen   bool
		wantAtClosed bool
	}{
		{true, true, false, true},
		{true, false, true, false},
		{false, true, false, false},
		{false, false, false, false},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)

	r := NewRelay(aSwitch, bSwitch)

	var gotOpenOut, gotClosedOut bool
	r.OpenOut.WireUp(func(state bool) { gotOpenOut = state })
	r.ClosedOut.WireUp(func(state bool) { gotClosedOut = state })

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip [%d]: Setting A power to %t and B power to %t", i+1, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			if gotOpenOut != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut)
			}

			if gotClosedOut != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut)
			}
		})
	}
}

func TestRelay_WithBatteries(t *testing.T) {
	testCases := []struct {
		aInPowered   bool
		bInPowered   bool
		wantAtOpen   bool
		wantAtClosed bool
	}{
		{false, false, false, false},
		{true, false, true, false},
		{false, true, false, false},
		{true, true, false, true},
	}

	var gotOpenOut, gotClosedOut bool
	var pin1Battery, pin2Battery *Battery

	pin1Battery = NewBattery()
	pin2Battery = NewBattery()

	rel := NewRelay(pin1Battery, pin2Battery)

	rel.OpenOut.WireUp(func(state bool) { gotOpenOut = state })
	rel.ClosedOut.WireUp(func(state bool) { gotClosedOut = state })

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %t and B to %t", tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			if tc.aInPowered {
				pin1Battery.Charge()
			} else {
				pin1Battery.Discharge()
			}
			if tc.bInPowered {
				pin2Battery.Charge()
			} else {
				pin2Battery.Discharge()
			}

			if gotOpenOut != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut)
			}

			if gotClosedOut != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut)
			}
		})
	}
}

func TestRelay_UpdatePinPanic_High(t *testing.T) {
	want := "Invalid relay pin number.  Relays have two pins and the requested pin was (3)"

	defer func() {
		if got := recover(); got != want {
			t.Errorf("Expected a panic of \"%s\" but got \"%s\"", want, got)
		}
	}()

	rel := NewRelay(NewBattery(), NewBattery())

	rel.UpdatePin(3, NewBattery())
}

func TestRelay_UpdatePinPanic_Low(t *testing.T) {
	want := "Invalid relay pin number.  Relays have two pins and the requested pin was (0)"

	defer func() {
		if got := recover(); got != want {
			t.Errorf("Expected a panic of \"%s\" but got \"%s\"", want, got)
		}
	}()

	rel := NewRelay(NewBattery(), NewBattery())

	rel.UpdatePin(0, NewBattery())
}

func TestANDGate(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		cInPowered bool
		want       bool
	}{
		{false, false, false, false},
		{true, false, false, false},
		{false, true, false, false},
		{true, true, false, false},
		{false, false, true, false},
		{true, false, true, false},
		{false, true, true, false},
		{true, true, true, true},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)
	cSwitch := NewSwitch(false)

	gate := NewANDGate(aSwitch, bSwitch, cSwitch)

	var got bool
	gate.WireUp(func(state bool) { got = state })

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip[%d]: Setting A power to %t and B power to %t and C power to %t", i+1, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			if got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestANDGate_UpdatePinPanic_TooHigh(t *testing.T) {
	want := "Invalid gate pin number.  Input pin count (2), requested pin (3)"

	defer func() {
		if got := recover(); got != want {
			t.Errorf("Expected a panic of \"%s\" but got \"%s\"", want, got)
		}
	}()

	gate := NewANDGate(NewBattery(), NewBattery())

	gate.UpdatePin(3, 1, NewBattery())
}

func TestANDGate_UpdatePinPanic_Low(t *testing.T) {
	want := "Invalid gate pin number.  Input pin count (2), requested pin (0)"

	defer func() {
		if got := recover(); got != want {
			t.Errorf("Expected a panic of \"%s\" but got \"%s\"", want, got)
		}
	}()

	gate := NewANDGate(NewBattery(), NewBattery())

	gate.UpdatePin(0, 1, NewBattery())
}

func TestORGate(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		cInPowered bool
		want       bool
	}{
		{false, false, false, false},
		{true, false, false, true},
		{false, true, false, true},
		{true, true, false, true},
		{false, false, true, true},
		{true, false, true, true},
		{false, true, true, true},
		{true, true, true, true},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)
	cSwitch := NewSwitch(false)

	gate := NewORGate(aSwitch, bSwitch, cSwitch)

	var got bool
	gate.WireUp(func(state bool) { got = state })

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip[%d]: Setting A power to %t and B power to %t and C power to %t", i+1, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			if got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestNANDGate(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		cInPowered bool
		want       bool
	}{
		{false, false, false, true},
		{true, false, false, true},
		{false, true, false, true},
		{true, true, false, true},
		{false, false, true, true},
		{true, false, true, true},
		{false, true, true, true},
		{true, true, true, false},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)
	cSwitch := NewSwitch(false)

	gate := NewNANDGate(aSwitch, bSwitch, cSwitch)

	var got bool
	gate.WireUp(func(state bool) { got = state })

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip[%d]: Setting A power to %t and B power to %t and C power to %t", i+1, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			if got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestNORGate(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		cInPowered bool
		want       bool
	}{
		{false, false, false, true},
		{true, false, false, false},
		{false, true, false, false},
		{true, true, false, false},
		{false, false, true, false},
		{true, false, true, false},
		{false, true, true, false},
		{true, true, true, false},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)
	cSwitch := NewSwitch(false)

	gate := NewNORGate(aSwitch, bSwitch, cSwitch)

	var got bool
	gate.WireUp(func(state bool) { got = state })

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip[%d]: Setting A power to %t and B power to %t and C power to %t", i+1, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			if got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestNORGate_UpdatePinPanic_TooHigh(t *testing.T) {
	want := "Invalid gate pin number.  Input pin count (2), requested pin (3)"

	defer func() {
		if got := recover(); got != want {
			t.Errorf("Expected a panic of \"%s\" but got \"%s\"", want, got)
		}
	}()

	gate := NewNORGate(NewBattery(), NewBattery())

	gate.UpdatePin(3, 1, NewBattery())
}

func TestNORGate_UpdatePinPanic_Low(t *testing.T) {
	want := "Invalid gate pin number.  Input pin count (2), requested pin (0)"

	defer func() {
		if got := recover(); got != want {
			t.Errorf("Expected a panic of \"%s\" but got \"%s\"", want, got)
		}
	}()

	gate := NewNORGate(NewBattery(), NewBattery())

	gate.UpdatePin(0, 1, NewBattery())
}

func TestXORGate(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		want       bool
	}{
		{false, false, false},
		{true, false, true},
		{false, true, true},
		{true, true, false},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)

	gate := NewXORGate(aSwitch, bSwitch)

	var got bool
	gate.WireUp(func(state bool) { got = state })

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip[%d]: Setting A power to %t and B power to %t", i+1, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			if got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestXNORGate(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		want       bool
	}{
		{false, false, true},
		{true, false, false},
		{false, true, false},
		{true, true, true},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)

	gate := NewXNORGate(aSwitch, bSwitch)

	var got bool
	gate.WireUp(func(state bool) { got = state })

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip[%d]: Setting A power to %t and B power to %t", i+1, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			if got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestInverter(t *testing.T) {
	testCases := []struct {
		inPowered bool
		wantOut   bool
	}{
		{false, true},
		{true, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Input as %t.", tc.inPowered), func(t *testing.T) {
			var pin1Battery *Battery

			pin1Battery = NewBattery()
			if !tc.inPowered {
				pin1Battery.Discharge()
			}

			inv := NewInverter(pin1Battery)

			var got bool
			inv.WireUp(func(state bool) { got = state })

			if got != tc.wantOut {
				t.Errorf("Input power was %t so wanted it inverted to %t but got %t", tc.inPowered, tc.wantOut, got)
			}
		})
	}
}

func TestHalfAdder(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		wantSum    bool
		wantCarry  bool
	}{
		{false, false, false, false},
		{true, false, true, false},
		{false, true, true, false},
		{true, true, false, true},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)

	half := NewHalfAdder(aSwitch, bSwitch)

	var gotSum, gotCarry bool
	half.Sum.WireUp(func(state bool) { gotSum = state })
	half.Carry.WireUp(func(state bool) { gotCarry = state })

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input source A to %t and source B to %t", tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			if gotSum != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, gotSum)
			}

			if gotCarry != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, gotCarry)
			}
		})
	}
}

func TestFullAdder(t *testing.T) {
	testCases := []struct {
		aInPowered     bool
		bInPowered     bool
		carryInPowered bool
		wantSum        bool
		wantCarry      bool
	}{
		{false, false, false, false, false},
		{true, false, false, true, false},
		{true, true, false, false, true},
		{true, true, true, true, true},
		{false, true, false, true, false},
		{false, true, true, false, true},
		{false, false, true, true, false},
		{true, false, true, false, true},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)
	cSwitch := NewSwitch(false)

	full := NewFullAdder(aSwitch, bSwitch, cSwitch)

	var gotSum, gotCarry bool
	full.Sum.WireUp(func(state bool) { gotSum = state })
	full.Carry.WireUp(func(state bool) { gotCarry = state })

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input source A to %t and source B to %t with carry in of %t", tc.aInPowered, tc.bInPowered, tc.carryInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.carryInPowered)

			if gotSum != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, gotSum)
			}

			if gotCarry != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, gotCarry)
			}
		})
	}
}

func TestNBitAdder_BadInputLengths(t *testing.T) {
	testCases := []struct {
		byte1     string
		byte2     string
		wantError string
	}{
		{"0", "00", "Mismatched addend lengths.  Addend1 len: 1, Addend2 len: 2"},
		{"00", "0", "Mismatched addend lengths.  Addend1 len: 2, Addend2 len: 1"},
		{"11111111", "111111111", "Mismatched addend lengths.  Addend1 len: 8, Addend2 len: 9"},
		{"111111111", "11111111", "Mismatched addend lengths.  Addend1 len: 9, Addend2 len: 8"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.byte1, tc.byte2), func(t *testing.T) {
			addend1Switches, _ := NewNSwitchBank(tc.byte1)
			addend2Switches, _ := NewNSwitchBank(tc.byte2)

			addr, err := NewNBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), nil)

			if addr != nil {
				t.Error("Did not expect an adder to be created, but got one")
			}

			if err == nil {
				t.Error("Expected an error on length mismatch but didn't get one")
			}

			if err.Error() != tc.wantError {
				t.Errorf("Wanted error %s, but got %s", tc.wantError, err.Error())
			}
		})
	}
}

func TestNBitAdder_EightBit_AsAnswerString(t *testing.T) {
	testCases := []struct {
		byte1          string
		byte2          string
		carryInPowered bool
		wantAnswer     string
		wantCarryOut   bool
	}{
		{"00000000", "00000000", false, "00000000", false},
		{"00000001", "00000000", false, "00000001", false},
		{"00000000", "00000001", false, "00000001", false},
		{"00000000", "00000000", true, "00000001", false},
		{"00000001", "00000000", true, "00000010", false},
		{"00000000", "00000001", true, "00000010", false},
		{"10000000", "10000000", false, "00000000", true},
		{"10000001", "10000000", false, "00000001", true},
		{"11111111", "11111111", false, "11111110", true},
		{"11111111", "11111111", true, "11111111", true},
		{"01111111", "11111111", false, "01111110", true},
		{"01111111", "11111111", true, "01111111", true},
		{"10101010", "01010101", false, "11111111", false},
		{"10101010", "01010101", true, "00000000", true},
	}

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, bit := range bits {
			switchBank.Switches[i].Set(bit == '1')
		}
	}

	// start with off switches
	addend1Switches, _ := NewNSwitchBank("00000000")
	addend2Switches, _ := NewNSwitchBank("00000000")
	carryInSwitch := NewSwitch(false)

	addr, _ := NewNBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), carryInSwitch)

	if addr == nil {
		t.Error("Expected an adder to return due to good inputs, but got addr nil one.")
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", tc.byte1, tc.byte2, tc.carryInPowered), func(t *testing.T) {

			updateSwitches(addend1Switches, tc.byte1)
			updateSwitches(addend2Switches, tc.byte2)
			carryInSwitch.Set(tc.carryInPowered)

			if got := addr.AsAnswerString(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := addr.CarryOutAsBool(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
	}
}

func TestNBitAdder_EightBit_AnswerViaRegistration(t *testing.T) {
	wantCarryOut := true
	var gotCarryOut bool

	wantAnswer := [8]bool{true, false, false, false, false, false, true, false}
	var gotAnswer [8]bool

	var callbackFuncs [8]func(state bool)
	callbackFuncs[0] = func(state bool) { gotAnswer[0] = state }
	callbackFuncs[1] = func(state bool) { gotAnswer[1] = state }
	callbackFuncs[2] = func(state bool) { gotAnswer[2] = state }
	callbackFuncs[3] = func(state bool) { gotAnswer[3] = state }
	callbackFuncs[4] = func(state bool) { gotAnswer[4] = state }
	callbackFuncs[5] = func(state bool) { gotAnswer[5] = state }
	callbackFuncs[6] = func(state bool) { gotAnswer[6] = state }
	callbackFuncs[7] = func(state bool) { gotAnswer[7] = state }

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	// start with off switches
	addend1Switches, _ := NewNSwitchBank("00000000")
	addend2Switches, _ := NewNSwitchBank("00000000")
	carryInSwitch := NewSwitch(false)

	addr, _ := NewNBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), carryInSwitch)

	for i, sum := range addr.Sums {
		sum.WireUp(callbackFuncs[i])
	}
	addr.CarryOut.WireUp(func(state bool) { gotCarryOut = state })

	updateSwitches(addend1Switches, "11000001")
	updateSwitches(addend2Switches, "11000000")
	carryInSwitch.Set(true)

	if gotAnswer != wantAnswer {
		t.Errorf("Wanted answer %v, but got %v", wantAnswer, gotAnswer)
	}

	if gotCarryOut != wantCarryOut {
		t.Errorf("Wanted carry %t, but got %t", wantCarryOut, gotCarryOut)
	}
}

func TestNBitAdder_SixteenBit_AsAnswerString(t *testing.T) {
	testCases := []struct {
		bytes1         string
		bytes2         string
		carryInPowered bool
		wantAnswer     string
		wantCarryOut   bool
	}{
		{"0000000000000000", "0000000000000000", false, "0000000000000000", false},
		{"0000000000000001", "0000000000000000", false, "0000000000000001", false},
		{"0000000000000000", "0000000000000001", false, "0000000000000001", false},
		{"0000000000000000", "0000000000000000", true, "0000000000000001", false},
		{"0000000000000001", "0000000000000000", true, "0000000000000010", false},
		{"0000000000000000", "0000000000000001", true, "0000000000000010", false},
		{"1000000000000000", "1000000000000000", false, "0000000000000000", true},
		{"1000000000000001", "1000000000000000", false, "0000000000000001", true},
		{"1111111111111111", "1111111111111111", false, "1111111111111110", true},
		{"1111111111111111", "1111111111111111", true, "1111111111111111", true},
		{"0000000001111111", "0000000011111111", false, "0000000101111110", false},
		{"0000000001111111", "0000000011111111", true, "0000000101111111", false},
		{"1010101010101010", "0101010101010101", false, "1111111111111111", false},
		{"1010101010101010", "0101010101010101", true, "0000000000000000", true},
		{"1001110110011101", "1101011011010110", false, "0111010001110011", true},
	}

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, bit := range bits {
			switchBank.Switches[i].Set(bit == '1')
		}
	}

	// start with off switches
	addend1Switches, _ := NewNSwitchBank("0000000000000000")
	addend2Switches, _ := NewNSwitchBank("0000000000000000")
	carryInSwitch := NewSwitch(false)

	addr, _ := NewNBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), carryInSwitch)

	if addr == nil {
		t.Error("Expected an adder to return due to good inputs, but got addr nil one.")
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", tc.bytes1, tc.bytes2, tc.carryInPowered), func(t *testing.T) {

			updateSwitches(addend1Switches, tc.bytes1)
			updateSwitches(addend2Switches, tc.bytes2)
			carryInSwitch.Set(tc.carryInPowered)

			if got := addr.AsAnswerString(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := addr.CarryOutAsBool(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
	}
}

func TestNBitAdder_SixteenBit_AnswerViaRegistration(t *testing.T) {
	wantCarryOut := true
	var gotCarryOut bool

	wantAnswer := [16]bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false}
	var gotAnswer [16]bool

	var callbackFuncs [16]func(state bool)
	callbackFuncs[0] = func(state bool) { gotAnswer[0] = state }
	callbackFuncs[1] = func(state bool) { gotAnswer[1] = state }
	callbackFuncs[2] = func(state bool) { gotAnswer[2] = state }
	callbackFuncs[3] = func(state bool) { gotAnswer[3] = state }
	callbackFuncs[4] = func(state bool) { gotAnswer[4] = state }
	callbackFuncs[5] = func(state bool) { gotAnswer[5] = state }
	callbackFuncs[6] = func(state bool) { gotAnswer[6] = state }
	callbackFuncs[7] = func(state bool) { gotAnswer[7] = state }
	callbackFuncs[8] = func(state bool) { gotAnswer[8] = state }
	callbackFuncs[9] = func(state bool) { gotAnswer[9] = state }
	callbackFuncs[10] = func(state bool) { gotAnswer[10] = state }
	callbackFuncs[11] = func(state bool) { gotAnswer[11] = state }
	callbackFuncs[12] = func(state bool) { gotAnswer[12] = state }
	callbackFuncs[13] = func(state bool) { gotAnswer[13] = state }
	callbackFuncs[14] = func(state bool) { gotAnswer[14] = state }
	callbackFuncs[15] = func(state bool) { gotAnswer[15] = state }

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, bit := range bits {
			switchBank.Switches[i].Set(bit == '1')
		}
	}

	// start with off switches
	addend1Switches, _ := NewNSwitchBank("0000000000000000")
	addend2Switches, _ := NewNSwitchBank("0000000000000000")
	carryInSwitch := NewSwitch(false)

	addr, _ := NewNBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), carryInSwitch)

	for i, sum := range addr.Sums {
		sum.WireUp(callbackFuncs[i])
	}
	addr.CarryOut.WireUp(func(state bool) { gotCarryOut = state })

	updateSwitches(addend1Switches, "1100000000000001")
	updateSwitches(addend2Switches, "1100000000000000")
	carryInSwitch.Set(true)

	if gotAnswer != wantAnswer {
		t.Errorf("Wanted answer %v, but got %v", wantAnswer, gotAnswer)
	}

	if gotCarryOut != wantCarryOut {
		t.Errorf("Wanted carry %t, but got %t", wantCarryOut, gotCarryOut)
	}
}

// -stopCh=XXX prevents the test running aspect from finding any tests
// go test -stopCh=XXX -bench=. -benchmem -count 5 > old.txt
// ---change some code---
// go test -stopCh=XXX -bench=. -benchmem -count 5 > new.txt

// go get golang.org/x/perf/cmd/benchstat
// benchstat old.txt new.txt

func BenchmarkNBitAdder_SixteenBit_AsAnswerString(b *testing.B) {
	benchmarks := []struct {
		name           string
		bytes1         string
		bytes2         string
		carryInPowered bool
	}{
		{"All zeros", "0000000000000000", "0000000000000000", false},
		{"All ones", "1111111111111111", "1111111111111111", false},
	}
	for _, bm := range benchmarks {
		carryInSwitch := NewSwitch(bm.carryInPowered)
		addend1Switches, _ := NewNSwitchBank(bm.bytes1)
		addend2Switches, _ := NewNSwitchBank(bm.bytes2)
		addend1BitPubs := addend1Switches.AsPwrEmitters()
		addend2BitPubs := addend2Switches.AsPwrEmitters()
		a, _ := NewNBitAdder(addend1BitPubs, addend2BitPubs, carryInSwitch)
		b.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", bm.bytes1, bm.bytes2, bm.carryInPowered), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				a.AsAnswerString()
			}
		})
	}
}

func TestOnesCompliment_AsComplimentString(t *testing.T) {

	testCases := []struct {
		bits            string
		signalIsPowered bool
		want            string
	}{
		{"0", false, "0"},
		{"0", true, "1"},
		{"1", false, "1"},
		{"1", true, "0"},
		{"00000000", false, "00000000"},
		{"00000000", true, "11111111"},
		{"11111111", true, "00000000"},
		{"10101010", false, "10101010"},
		{"10101010", true, "01010101"},
		{"1010101010101010101010101010101010101010", false, "1010101010101010101010101010101010101010"},
		{"1010101010101010101010101010101010101010", true, "0101010101010101010101010101010101010101"},
	}

	getInputs := func(bits string) []pwrEmitter {
		pubs := []pwrEmitter{}

		for _, bit := range bits {
			pubs = append(pubs, NewSwitch(bit == '1'))
		}

		return pubs
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Executing complementer against %s with readFromLatch of %t", tc.bits, tc.signalIsPowered), func(t *testing.T) {
			comp := NewOnesComplementer(getInputs(tc.bits), NewSwitch(tc.signalIsPowered))

			if comp == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
			}

			if got := comp.AsComplementString(); got != tc.want {
				t.Errorf("Wanted %s, but got %s", tc.want, got)
			}
		})
	}
}

func TestOnesCompliment_Compliments(t *testing.T) {

	testCases := []struct {
		bits            string
		signalIsPowered bool
		want            []bool
	}{
		{"0", false, []bool{false}},
		{"0", true, []bool{true}},
		{"1", false, []bool{true}},
		{"1", true, []bool{false}},
		{"00000000", false, []bool{false, false, false, false, false, false, false, false}},
		{"00000000", true, []bool{true, true, true, true, true, true, true, true}},
		{"11111111", true, []bool{false, false, false, false, false, false, false, false}},
		{"10101010", false, []bool{true, false, true, false, true, false, true, false}},
		{"10101010", true, []bool{false, true, false, true, false, true, false, true}},
		{"101010101010", false, []bool{true, false, true, false, true, false, true, false, true, false, true, false}},
		{"101010101010", true, []bool{false, true, false, true, false, true, false, true, false, true, false, true}},
	}

	getInputs := func(bits string) []pwrEmitter {
		pubs := []pwrEmitter{}

		for _, bit := range bits {
			pubs = append(pubs, NewSwitch(bit == '1'))
		}

		return pubs
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Executing complementer against %s with readFromLatch of %t", tc.bits, tc.signalIsPowered), func(t *testing.T) {
			comp := NewOnesComplementer(getInputs(tc.bits), NewSwitch(tc.signalIsPowered))

			if comp == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
			}

			for i, pub := range comp.Complements {
				got := pub.(*XORGate).GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf("At index %d, wanted %v but got %v", i, want, got)
				}
			}
		})
	}
}

func TestNBitSubtractor_BadInputLengths(t *testing.T) {
	testCases := []struct {
		byte1     string
		byte2     string
		wantError string
	}{
		{"0", "00", "Mismatched input lengths.  Minuend len: 1, Subtrahend len: 2"},
		{"00", "0", "Mismatched input lengths.  Minuend len: 2, Subtrahend len: 1"},
		{"11111111", "111111111", "Mismatched input lengths.  Minuend len: 8, Subtrahend len: 9"},
		{"111111111", "11111111", "Mismatched input lengths.  Minuend len: 9, Subtrahend len: 8"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.byte1, tc.byte2), func(t *testing.T) {
			addend1Switches, _ := NewNSwitchBank(tc.byte1)
			addend2Switches, _ := NewNSwitchBank(tc.byte2)

			sub, err := NewNBitSubtractor(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters())

			if sub != nil {
				t.Error("Did not expect a Subtractor to be created, but got one")
			}

			if err == nil {
				t.Error("Expected an error on length mismatch but didn't get one")
			}

			if err.Error() != tc.wantError {
				t.Errorf("Wanted error %sub, but got %s", tc.wantError, err.Error())
			}
		})
	}
}

func TestNBitSubtractor_EightBit_AsAnswerString(t *testing.T) {
	testCases := []struct {
		minuend      string
		subtrahend   string
		wantAnswer   string
		wantCarryOut bool
	}{
		{"00000000", "00000000", "00000000", true},  // 0 - 0 = 0
		{"00000001", "00000000", "00000001", true},  // 1 - 0 = 1
		{"00000001", "00000001", "00000000", true},  // 1 - 1 = 0
		{"00000011", "00000001", "00000010", true},  // 3 - 1 = 2
		{"10000000", "00000001", "01111111", true},  // -128 - 1 = 127 signed (or 128 - 1 = 127 unsigned)
		{"11111111", "11111111", "00000000", true},  // -1 - -1 = 0 signed (or 255 - 255 = 0 unsigned)
		{"11111111", "00000001", "11111110", true},  // -1 - 1 = -2 signed (or 255 - 1 = 254 unsigned)
		{"10000001", "00000001", "10000000", true},  // -127 - 1 = -128 signed (or 129 - 1 = 128 unsigned)
		{"11111110", "11111011", "00000011", true},  // -2 - -5 = 3 (or 254 - 251 = 3 unsigned)
		{"00000000", "00000001", "11111111", false}, // 0 - 1 = -1 signed (or 255 unsigned)
		{"00000010", "00000011", "11111111", false}, // 2 - 3 = -1 signed (or 255 unsigned)
		{"11111110", "11111111", "11111111", false}, // -2 - -1 = -1 signed or (254 - 255 = 255 unsigned)
		{"10000001", "01111110", "00000011", true},  // -127 - 126 = 3 signed or (129 - 126 = 3 unsigned)
	}

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, bit := range bits {
			switchBank.Switches[i].Set(bit == '1')
		}
	}

	// start with off switches
	minuendwitches, _ := NewNSwitchBank("00000000")
	subtrahendSwitches, _ := NewNSwitchBank("00000000")

	sub, _ := NewNBitSubtractor(minuendwitches.AsPwrEmitters(), subtrahendSwitches.AsPwrEmitters())

	if sub == nil {
		t.Error("Expected an subtractor to return due to good inputs, but gotAnswer c nil one.")
		return // cannot continue tests if no subtractor to test
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Subtracting %sub from %sub", tc.subtrahend, tc.minuend), func(t *testing.T) {

			updateSwitches(minuendwitches, tc.minuend)
			updateSwitches(subtrahendSwitches, tc.subtrahend)

			if gotAnswer := sub.AsAnswerString(); gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %sub, but gotAnswer %sub", tc.wantAnswer, gotAnswer)
			}

			if gotCarryOut := sub.CarryOutAsBool(); gotCarryOut != tc.wantCarryOut {
				t.Errorf("Wanted carry out %t, but gotAnswer %t", tc.wantCarryOut, gotCarryOut)
			}
		})
	}
}

func TestNBitSubtractor_EightBit_AnswerViaRegistration(t *testing.T) {
	wantCarryOut := true
	var gotCarryOut bool

	wantAnswer := [8]bool{false, false, false, false, false, false, true, true}
	var gotAnswer [8]bool

	var callbackFuncs [8]func(state bool)
	callbackFuncs[0] = func(state bool) { gotAnswer[0] = state }
	callbackFuncs[1] = func(state bool) { gotAnswer[1] = state }
	callbackFuncs[2] = func(state bool) { gotAnswer[2] = state }
	callbackFuncs[3] = func(state bool) { gotAnswer[3] = state }
	callbackFuncs[4] = func(state bool) { gotAnswer[4] = state }
	callbackFuncs[5] = func(state bool) { gotAnswer[5] = state }
	callbackFuncs[6] = func(state bool) { gotAnswer[6] = state }
	callbackFuncs[7] = func(state bool) { gotAnswer[7] = state }

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	// start with off switches
	minuendSwitches, _ := NewNSwitchBank("00000000")
	subtrahendSwitches, _ := NewNSwitchBank("00000000")

	sub, _ := NewNBitSubtractor(minuendSwitches.AsPwrEmitters(), subtrahendSwitches.AsPwrEmitters())

	for i, dif := range sub.Differences {
		dif.WireUp(callbackFuncs[i])
	}

	sub.CarryOut.WireUp(func(state bool) { gotCarryOut = state })

	updateSwitches(minuendSwitches, "10000001")
	updateSwitches(subtrahendSwitches, "01111110")

	if gotAnswer != wantAnswer {
		t.Errorf("Wanted answer %v, but got %v", wantAnswer, gotAnswer)
	}

	if gotCarryOut != wantCarryOut {
		t.Errorf("Wanted carry out %t, but got %t", wantCarryOut, gotCarryOut)
	}
}

// Fragile test due to timing of asking Oscillator vs. isPowered of Oscillator at the time being asked
func TestOscillator(t *testing.T) {
	testCases := []struct {
		initState   bool
		oscHertz    int
		wantResults string
	}{
		{false, 1, "FTF"},
		{true, 1, "TFT"},
		{false, 5, "FTFTFTFTFTF"},
		{true, 5, "TFTFTFTFTFT"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Oscillating at %d hertz, immediate start (%t)", tc.oscHertz, tc.initState), func(t *testing.T) {

			var gotResults string

			osc := NewOscillator(tc.initState)
			osc.Oscillate(tc.oscHertz)

			osc.WireUp(func(state bool) {
				if state {
					gotResults += "T"
				} else {
					gotResults += "F"
				}
			})

			time.Sleep(time.Second * 2)

			osc.Stop()

			if gotResults != tc.wantResults {
				t.Errorf("Wanted results %s but got %s.", tc.wantResults, gotResults)
			}
		})
	}
}

func TestRSFlipFlop(t *testing.T) {
	testCases := []struct {
		rPinPowered bool
		sPinPowered bool
		wantQ       bool
		wantQBar    bool
	}{ // contsruction of the flipflop will start with a default of rPin:false, sPin:false, which causes false on both inputs of the S nor, which causes QBar on (Qs off)
		{false, false, false, true}, // Un-Set should remember prior
		{false, true, true, false},  // Set causes Qs on (QBar off)
		{false, true, true, false},  // Set again should change nothing
		{false, false, true, false}, // Un-Set should remember prior
		{false, false, true, false}, // Un-Set again should change nothing
		{true, false, false, true},  // Reset causes Qs off (QBar on)
		{true, false, false, true},  // Reset again should change nothing
		{false, false, false, true}, // Un-Reset should remember prior
		{true, false, false, true},  // Un-Reset again should change nothing
		{false, true, true, false},  // Set causes Qs on (QBar off)
		{true, false, false, true},  // Reset causes Qs off (QBar on)
		{false, true, true, false},  // Set causes Qs on (QBar off)
	}

	testName := func(i int) string {
		var priorR bool
		var priorS bool

		if i == 0 {
			priorR = false
			priorS = false
		} else {
			priorR = testCases[i-1].rPinPowered
			priorS = testCases[i-1].sPinPowered
		}

		return fmt.Sprintf("Stage %d: Switching from [rInPowered (%t) sInPowered (%t)] to [rInPowered (%t) sInPowered (%t)]", i+1, priorR, priorS, testCases[i].rPinPowered, testCases[i].sPinPowered)
	}

	var rPinBattery, sPinBattery *Battery
	rPinBattery = NewBattery()
	sPinBattery = NewBattery()
	rPinBattery.Discharge()
	sPinBattery.Discharge()

	// starting with no input signals
	ff := NewRSFlipFLop(rPinBattery, sPinBattery)

	if gotQ := ff.Q.GetIsPowered(); gotQ != false {
		t.Errorf("Wanted power of %t at Qs, but got %t.", false, gotQ)
	}

	if gotQBar := ff.QBar.GetIsPowered(); gotQBar != true {
		t.Errorf("Wanted power of %t at QBar, but got %t.", true, gotQBar)
	}

	for i, tc := range testCases {
		t.Run(testName(i), func(t *testing.T) {

			// must discharge both first since power at R and S is disallowed
			rPinBattery.Discharge()
			sPinBattery.Discharge()

			if tc.rPinPowered {
				rPinBattery.Charge()
			}

			if tc.sPinPowered {
				sPinBattery.Charge()
			}

			if gotQ := ff.Q.GetIsPowered(); gotQ != tc.wantQ {
				t.Errorf("Wanted power of %t at Qs, but got %t.", tc.wantQ, gotQ)
			}

			if gotQBar := ff.QBar.GetIsPowered(); gotQBar != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar)
			}
		})
	}
}

func TestRSFlipFlop_Panic(t *testing.T) {
	want := "A Flip-Flop cannot have equivalent power status at both Qs and QBar"

	defer func() {
		if got := recover(); !strings.HasPrefix(got.(string), want) {
			t.Errorf("Expected a panic of \"%s\" but got \"%s\"", want, got)
		}
	}()

	// use two ON batteries to trigger invalid state
	NewRSFlipFLop(NewBattery(), NewBattery())
}

func TestLevelTriggeredDTypeLatch(t *testing.T) {
	testCases := []struct {
		clkIn    bool
		dataIn   bool
		wantQ    bool
		wantQBar bool
	}{ // construction of the latchStore will start with a default of clkIn:true, dataIn:true, which causes Qs on (QBar off)
		{false, false, true, false}, // clkIn off should cause no change
		{false, true, true, false},  // clkIn off should cause no change
		{true, true, true, false},   // clkIn with dataIn causes Qs on (QBar off)
		{false, false, true, false}, // clkIn off should cause no change
		{true, false, false, true},  // clkIn with no dataIn causes Qs off (QBar on)
		{false, false, false, true}, // clkIn off should cause no change
		{true, false, false, true},  // clkIn again with same dataIn should cause no change
		{true, true, true, false},   // clkIn with dataIn should cause Qs on (QBar off)
		{false, false, true, false}, // clkIn off should cause no change
	}

	testName := func(i int) string {
		var priorClkIn bool
		var priorDataIn bool

		if i == 0 {
			// trues since starting with charged batteries when Newing thew Latch initially
			priorClkIn = true
			priorDataIn = true
		} else {
			priorClkIn = testCases[i-1].clkIn
			priorDataIn = testCases[i-1].dataIn
		}

		return fmt.Sprintf("Stage %d: Switching from [clkIn (%t) dataIn (%t)] to [clkIn (%t) dataIn (%t)]", i+1, priorClkIn, priorDataIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	var holdBattery, dataBattery *Battery
	holdBattery = NewBattery()
	dataBattery = NewBattery()

	latch := NewLevelTriggeredDTypeLatch(holdBattery, dataBattery)

	for i, tc := range testCases {
		t.Run(testName(i), func(t *testing.T) {

			if tc.clkIn {
				holdBattery.Charge()
			} else {
				holdBattery.Discharge()
			}

			if tc.dataIn {
				dataBattery.Charge()
			} else {
				dataBattery.Discharge()
			}

			if gotQ := latch.Q.GetIsPowered(); gotQ != tc.wantQ {
				t.Errorf("Wanted power of %t at Qs, but got %t.", tc.wantQ, gotQ)
			}

			if gotQBar := latch.QBar.GetIsPowered(); gotQBar != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar)
			}
		})
	}
}

func TestNBitLatch(t *testing.T) {
	testCases := []struct {
		input string
		want  [8]bool
	}{
		{"00000001", [8]bool{false, false, false, false, false, false, false, true}},
		{"11111111", [8]bool{true, true, true, true, true, true, true, true}},
		{"10101010", [8]bool{true, false, true, false, true, false, true, false}},
		{"10000001", [8]bool{true, false, false, false, false, false, false, true}},
	}

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	latchSwitches, _ := NewNSwitchBank("00000000")
	clkSwitch := NewSwitch(false)
	latch := NewNBitLatch(clkSwitch, latchSwitches.AsPwrEmitters())

	priorWant := [8]bool{false, false, false, false, false, false, false, false}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Stage %d: Setting switches to %s", i+1, tc.input), func(t *testing.T) {

			// set to OFF to test that nothing will change in the latchStore store

			clkSwitch.Set(false)
			updateSwitches(latchSwitches, tc.input)

			for i, pwr := range latch.AsPwrEmitters() {
				got := pwr.(*NORGate).GetIsPowered()
				want := priorWant[i]

				if got != want {
					t.Errorf("[As PwrEmitter] At index %d, with clkSwitch off, wanted %v but got %v", i, want, got)
				}
			}

			// Now set to ON to test that requested changes did occur in the latchStore store

			clkSwitch.Set(true)

			for i, pwr := range latch.AsPwrEmitters() {
				got := pwr.(*NORGate).GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf("[As PwrEmitter] At index %d, with clkSwitch on, wanted %v but got %v", i, want, got)
				}
			}

			// now update the prior tracker bools to ensure next pass (with cklIn as OFF at the top) proves it didn't change (so matches prior)
			for i, q := range latch.Qs {
				priorWant[i] = q.(*NORGate).GetIsPowered()
			}
		})
	}
}

func TestTwoToOneSelector_BadInputLengths(t *testing.T) {
	testCases := []struct {
		byte1     string
		byte2     string
		wantError string
	}{
		{"1111", "000", "Mismatched input lengths. aPins len: 4, bPins len: 3"},
		{"000", "1111", "Mismatched input lengths. aPins len: 3, bPins len: 4"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.byte1, tc.byte2), func(t *testing.T) {
			addend1Switches, _ := NewNSwitchBank(tc.byte1)
			addend2Switches, _ := NewNSwitchBank(tc.byte2)

			sel, err := NewTwoToOneSelector(nil, addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters())

			if sel != nil {
				t.Error("Did not expect a Selector to be created, but got one")
			}

			if err == nil {
				t.Error("Expected an error on length mismatch but didn't get one")
			}

			if err.Error() != tc.wantError {
				t.Errorf("Wanted error %s, but got %s", tc.wantError, err.Error())
			}
		})
	}
}

func TestTwoToOneSelector(t *testing.T) {
	testCases := []struct {
		aIn     string
		bIn     string
		selectB bool
		want    []bool
	}{
		{"000", "111", false, []bool{false, false, false}},
		{"000", "111", true, []bool{true, true, true}},
		{"111", "000", true, []bool{false, false, false}},
		{"111", "000", false, []bool{true, true, true}},
	}

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, bit := range bits {
			switchBank.Switches[i].Set(bit == '1')
		}
	}

	// start with off switches
	aInSwitches, _ := NewNSwitchBank("000")
	bInSwitches, _ := NewNSwitchBank("000")
	selectBSwitch := NewSwitch(false)

	sel, _ := NewTwoToOneSelector(selectBSwitch, aInSwitches.AsPwrEmitters(), bInSwitches.AsPwrEmitters())

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("With aIn as %s and bIn as %s, selecting bIn (%t)", tc.aIn, tc.bIn, tc.selectB), func(t *testing.T) {

			updateSwitches(aInSwitches, tc.aIn)
			updateSwitches(bInSwitches, tc.bIn)
			selectBSwitch.Set(tc.selectB)

			for i, out := range sel.Outs {
				got := out.(*ORGate).GetIsPowered()

				if got != tc.want[i] {
					t.Errorf("At index %d, with signal %t, wanted %t but got %t", i, tc.selectB, tc.want[i], got)
				}
			}
		})
	}
}

func TestClunkyAdder_MismatchInputs(t *testing.T) {
	wantError := "Mismatched input lengths. Switchbank 1 switch count: 8, Switchbank 2 switch count: 4"

	aInSwitches, _ := NewNSwitchBank("00000000")
	bInSwitches, _ := NewNSwitchBank("0000")

	addr, err := NewClunkyAdder(aInSwitches, bInSwitches)

	if addr != nil {
		t.Error("Did not expect an adder back but got one.")
	}
	if err != nil && err.Error() != wantError {
		t.Errorf("Wanted error %s, but got %", wantError, err.Error())
	}
}

func TestClunkyAdder_InitialAnswer(t *testing.T) {
	testCases := []struct {
		aIn        string
		bIn        string
		wantAnswer string
		wantCarry  bool
	}{
		{"00000000", "00000001", "00000001", false},
		{"00000001", "00000010", "00000011", false},
		{"10000001", "10000000", "00000001", true},
		{"11111111", "11111111", "11111110", true},
	}

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, bit := range bits {
			switchBank.Switches[i].Set(bit == '1')
		}
	}

	aInSwitches, _ := NewNSwitchBank("00000000")
	bInSwitches, _ := NewNSwitchBank("00000000")
	addr, _ := NewClunkyAdder(aInSwitches, bInSwitches)

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.aIn, tc.bIn), func(t *testing.T) {

			updateSwitches(aInSwitches, tc.aIn)
			updateSwitches(bInSwitches, tc.bIn)

			if gotAnswer := addr.AsAnswerString(); gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s but %s", tc.wantAnswer, gotAnswer)
			}

			if gotCarry := addr.CarryOutAsBool(); gotCarry != tc.wantCarry {
				t.Errorf("Wanted carry %t, but %t", tc.wantCarry, gotCarry)
			}
		})
	}
}

func TestClunkyAdder_MultiAdd(t *testing.T) {

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *NSwitchBank, bits string) {
		for i, bit := range bits {
			switchBank.Switches[i].Set(bit == '1')
		}
	}

	aInSwitches, _ := NewNSwitchBank("00000000")
	bInSwitches, _ := NewNSwitchBank("00000001")
	addr, _ := NewClunkyAdder(aInSwitches, bInSwitches)

	//addr.SaveToLatch.Set(true)
	addr.ReadFromLatch.Set(true)
	updateSwitches(aInSwitches, "00000010")

	wantAnswer := "00000011"
	wantCarry := false

	if gotAnswer := addr.AsAnswerString(); gotAnswer != wantAnswer {
		t.Errorf("Wanted answer %s but %s", wantAnswer, gotAnswer)
	}

	if gotCarry := addr.CarryOutAsBool(); gotCarry != wantCarry {
		t.Errorf("Wanted carry %t, but %t", wantCarry, gotCarry)
	}
}
