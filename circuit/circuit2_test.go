package circuit

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// go test
// go test -race -v (verbose)
// go test -race -cpu=1,2,4 (go max prox)
// go test -v

func TestPwrSource(t *testing.T) {
	var want, got1, got2 bool

	p := &pwrSource{}

	p.WireUp(func(state bool) { got1 = state })
	p.WireUp(func(state bool) { got2 = state })

	p.Transmit(true)
	want = true

	if got1 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 1 to be %t but got %t", want, got1))
	}

	if got2 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 2 to be %t but got %t", want, got2))
	}

	p.Transmit(false)
	want = false

	if got1 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 1 to be %t but got %t", want, got1))
	}

	if got2 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 2 to be %t but got %t", want, got2))
	}
}

func TestBattery(t *testing.T) {
	var want, got bool

	b := NewBattery()

	b.WireUp(func(state bool) { got = state })

	want = true

	if got != want {
		t.Errorf(fmt.Sprintf("With a new battery, wanted the subscriber to see power as %t but got %t", want, got))
	}

	b.Discharge()
	want = false

	if got != want {
		t.Errorf(fmt.Sprintf("With a discharged battery, wanted the subscriber'l IsPowered to be %t but got %t", want, got))
	}

	b.Charge()
	want = true

	if got != want {
		t.Errorf(fmt.Sprintf("With a charged battery, wanted the subscriber'l IsPowered to be %t but got %t", want, got))
	}
}

func TestSwitch(t *testing.T) {
	var wantState, gotState bool
	var wantCount, gotCount int

	// start with switch being off
	s := NewSwitch(false)

	// register callback (will trigger immediate call to push isPowered at time of registration)
	s.WireUp(func(state bool) {
		gotState = state
		gotCount += 1
	})

	// initial turn on
	s.Set(true)
	wantState = true
	wantCount = 2

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an off switch turned on, wanted the subscriber to see power as %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an off switch turned on, wanted the subscriber to see power as %d but got %d", wantCount, gotCount))
	}

	// turn on again though already on
	s.Set(true)
	wantCount = 2

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an attempt to turn on an already on switch, wanted the subscriber'c call count to remain %d but got %d", wantCount, gotCount))
	}
}

func TestNewEightSwitchBank_BadInputs(t *testing.T) {
	testCases := []struct {
		input     string
		wantError string
	}{
		{"0000000", "Input not in 8-bit binary format: "},
		{"00000000X", "Input not in 8-bit binary format: "},
		{"X00000000", "Input not in 8-bit binary format: "},
		{"bad", "Input not in 8-bit binary format: "},
		{"", "Input not in 8-bit binary format: "},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting switches to %s", tc.input), func(t *testing.T) {
			sb, err := NewEightSwitchBank(tc.input)

			if sb != nil {
				t.Error("Didn't expected a Switch Bank back but got one.")
			}

			tc.wantError += tc.input

			if err == nil || (err != nil && err.Error() != tc.wantError) {
				t.Error(fmt.Sprintf("Wanted error \"%s\" but got \"%v\"", tc.wantError, err))
			}
		})
	}
}

func TestNewEightSwitchBank_GoodInputs(t *testing.T) {
	testCases := []struct {
		input string
		want  [8]bool
	}{
		{"00000000", [8]bool{false, false, false, false, false, false, false, false}},
		{"11111111", [8]bool{true, true, true, true, true, true, true, true}},
		{"10101010", [8]bool{true, false, true, false, true, false, true, false}},
		{"10000001", [8]bool{true, false, false, false, false, false, false, true}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting switches to %s", tc.input), func(t *testing.T) {
			sb, err := NewEightSwitchBank(tc.input)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
			}

			// try as actual switches
			for i, s := range sb.Switches {
				got := s.GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As Switch] At index %d, wanted %v but got %v", i, want, got))
				}
			}

			// now try AsPwrEmitters
			for i, pwr := range sb.AsPwrEmitters() {
				got := pwr.(*Switch).GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As PwrEmitter] At index %d, wanted %v but got %v", i, want, got))
				}
			}
		})
	}
}

func TestNewSixteenSwitchBank_BadInputs(t *testing.T) {
	testCases := []struct {
		input     string
		wantError string
	}{
		{"000000000000000", "Input not in 16-bit binary format: "},
		{"0000000000000000X", "Input not in 16-bit binary format: "},
		{"X0000000000000000", "Input not in 16-bit binary format: "},
		{"bad", "Input not in 16-bit binary format: "},
		{"", "Input not in 16-bit binary format: "},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting switches to %s", tc.input), func(t *testing.T) {
			sb, err := NewSixteenSwitchBank(tc.input)

			if sb != nil {
				t.Error("Didn't expected a Switch Bank back but got one.")
			}

			tc.wantError += tc.input

			if err == nil || (err != nil && err.Error() != tc.wantError) {
				t.Error(fmt.Sprintf("Wanted error \"%s\" but got \"%v\"", tc.wantError, err))
			}
		})
	}
}

func TestNewSixteenSwitchBank_GoodInputs(t *testing.T) {
	testCases := []struct {
		input string
		want  [16]bool
	}{
		{"0000000000000000", [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
		{"1111111111111111", [16]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}},
		{"1010101010101010", [16]bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false}},
		{"1000000000000001", [16]bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting switches to %s", tc.input), func(t *testing.T) {
			sb, err := NewSixteenSwitchBank(tc.input)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
			}

			// try as actual switches
			for i, s := range sb.Switches {
				got := s.GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As Switch] At index %d, wanted %v but got %v", i, want, got))
				}
			}

			// now try AsPwrEmitters
			for i, pwr := range sb.AsPwrEmitters() {
				got := pwr.(*Switch).GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As PwrEmitter] At index %d, wanted %v but got %v", i, want, got))
				}
			}
		})
	}
}

func TestNewNSwitchBank_BadInputs(t *testing.T) {
	testCases := []struct {
		input     string
		wantError string
	}{
		{"000X", "Input not in binary format: "},
		{"X000", "Input not in binary format: "},
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
				t.Error(fmt.Sprintf("Wanted error \"%s\" but got \"%v\"", tc.wantError, err))
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
		{"10000001", []bool{true, false, false, false, false, false, false, true}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting switches to %s", tc.input), func(t *testing.T) {
			sb, err := NewNSwitchBank(tc.input)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
			}

			// try as actual switches
			for i, s := range sb.Switches {
				got := s.GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As Switch] At index %d, wanted %v but got %v", i, want, got))
				}
			}

			// now try AsPwrEmitters
			for i, pwr := range sb.AsPwrEmitters() {
				got := pwr.(*Switch).GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As PwrEmitter] At index %d, wanted %v but got %v", i, want, got))
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

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %t and B to %t", tc.aInPowered, tc.bInPowered), func(t *testing.T) {
			var gotOpenOut, gotClosedOut bool
			var pin1Battery, pin2Battery *Battery

			pin1Battery = NewBattery()
			pin2Battery = NewBattery()

			if !tc.aInPowered {
				pin1Battery.Discharge()
			}
			if !tc.bInPowered {
				pin2Battery.Discharge()
			}

			r := NewRelay(pin1Battery, pin2Battery)

			r.OpenOut.WireUp(func(state bool) { gotOpenOut = state })
			r.ClosedOut.WireUp(func(state bool) { gotClosedOut = state })

			if gotOpenOut != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut)
			}

			if gotClosedOut != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut)
			}
		})
	}
}

func TestRelay_UpdatePinPanic(t *testing.T) {
	want := "Invalid relay pin number.  Relays have two pins and the requested pin was (3)"

	defer func() {
		if got := recover(); got != want {
			t.Errorf(fmt.Sprintf("Expected a panic of \"%s\" but got \"%s\"", want, got))
		}
	}()

	r := NewRelay(NewBattery(), NewBattery())

	r.UpdatePin(3, NewBattery())
	r.UpdatePin(0, NewBattery())
}

func TestANDGate_TwoPin(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		want       bool
	}{
		{false, false, false},
		{true, false, false},
		{false, true, false},
		{true, true, true},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)

	g := NewANDGate(aSwitch, bSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

func TestANDGate_ThreePin(t *testing.T) {
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

	g := NewANDGate(aSwitch, bSwitch, cSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

func TestORGate_TwoPin(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		want       bool
	}{
		{false, false, false},
		{true, false, true},
		{false, true, true},
		{true, true, true},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)

	g := NewORGate(aSwitch, bSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

func TestORGate_ThreePin(t *testing.T) {
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

	g := NewORGate(aSwitch, bSwitch, cSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

func TestNANDGate_TwoPin(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		want       bool
	}{
		{false, false, true},
		{true, false, true},
		{false, true, true},
		{true, true, false},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)

	g := NewNANDGate(aSwitch, bSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

func TestNANDGate_ThreePin(t *testing.T) {
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

	g := NewNANDGate(aSwitch, bSwitch, cSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

func TestNORGate_TwoPin(t *testing.T) {
	testCases := []struct {
		aInPowered bool
		bInPowered bool
		want       bool
	}{
		{false, false, true},
		{true, false, false},
		{false, true, false},
		{true, true, false},
	}

	aSwitch := NewSwitch(false)
	bSwitch := NewSwitch(false)

	g := NewNORGate(aSwitch, bSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

func TestNORGate_ThreePin(t *testing.T) {
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

	g := NewNORGate(aSwitch, bSwitch, cSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

func TestNORGate_UpdatePinPanic(t *testing.T) {
	want := "Invalid gate pin number.  Input pin count (2), requested pin (3)"

	defer func() {
		if got := recover(); got != want {
			t.Errorf(fmt.Sprintf("Expected a panic of \"%s\" but got \"%s\"", want, got))
		}
	}()

	g := NewNORGate(NewBattery(), NewBattery())

	g.UpdatePin(3, 1, NewBattery())
	g.UpdatePin(0, 1, NewBattery())
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

	g := NewXORGate(aSwitch, bSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

	g := NewXNORGate(aSwitch, bSwitch)

	var got bool
	g.WireUp(func(state bool) { got = state })

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

	h := NewHalfAdder(aSwitch, bSwitch)

	var gotSum, gotCarry bool
	h.Sum.WireUp(func(state bool) { gotSum = state })
	h.Carry.WireUp(func(state bool) { gotCarry = state })

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

	h := NewFullAdder(aSwitch, bSwitch, cSwitch)

	var gotSum, gotCarry bool
	h.Sum.WireUp(func(state bool) { gotSum = state })
	h.Carry.WireUp(func(state bool) { gotCarry = state })

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

func TestEightBitAdder_AsAnswerString(t *testing.T) {
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
	updateSwitches := func(switchBank *EightSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	// start with off switches
	addend1Switches, _ := NewEightSwitchBank("00000000")
	addend2Switches, _ := NewEightSwitchBank("00000000")
	carryInSwitch := NewSwitch(false)

	a := NewEightBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), carryInSwitch)

	if a == nil {
		t.Error("Expected an adder to return due to good inputs, but got a nil one.")
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", tc.byte1, tc.byte2, tc.carryInPowered), func(t *testing.T) {

			updateSwitches(addend1Switches, tc.byte1)
			updateSwitches(addend2Switches, tc.byte2)
			carryInSwitch.Set(tc.carryInPowered)

			if got := a.AsAnswerString(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := a.CarryOutAsBool(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
	}
}

func TestEightBitAdder_AnswerViaRegistration(t *testing.T) {
	wantCarryOut := true
	var gotCarryOut bool

	wantAnswer := [8]bool{true, false, false, false, false, false, true, false}
	var gotAnswer [8]bool

	var f [8]func(state bool)
	f[0] = func(state bool) { gotAnswer[0] = state }
	f[1] = func(state bool) { gotAnswer[1] = state }
	f[2] = func(state bool) { gotAnswer[2] = state }
	f[3] = func(state bool) { gotAnswer[3] = state }
	f[4] = func(state bool) { gotAnswer[4] = state }
	f[5] = func(state bool) { gotAnswer[5] = state }
	f[6] = func(state bool) { gotAnswer[6] = state }
	f[7] = func(state bool) { gotAnswer[7] = state }

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *EightSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	// start with off switches
	addend1Switches, _ := NewEightSwitchBank("00000000")
	addend2Switches, _ := NewEightSwitchBank("00000000")
	carryInSwitch := NewSwitch(false)

	a := NewEightBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), carryInSwitch)

	for i, s := range a.Sums {
		s.WireUp(f[i])
	}
	a.CarryOut.WireUp(func(state bool) { gotCarryOut = state })

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

func TestSixteenBitAdder_AsAnswerString(t *testing.T) {
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
	updateSwitches := func(switchBank *SixteenSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	// start with off switches
	addend1Switches, _ := NewSixteenSwitchBank("0000000000000000")
	addend2Switches, _ := NewSixteenSwitchBank("0000000000000000")
	carryInSwitch := NewSwitch(false)

	a := NewSixteenBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), carryInSwitch)

	if a == nil {
		t.Error("Expected an adder to return due to good inputs, but got a nil one.")
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", tc.bytes1, tc.bytes2, tc.carryInPowered), func(t *testing.T) {

			updateSwitches(addend1Switches, tc.bytes1)
			updateSwitches(addend2Switches, tc.bytes2)
			carryInSwitch.Set(tc.carryInPowered)

			if got := a.AsAnswerString(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := a.CarryOutAsBool(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
	}
}

func TestSixteenBitAdder_AnswerViaRegistration(t *testing.T) {
	wantCarryOut := true
	var gotCarryOut bool

	wantAnswer := [16]bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false}
	var gotAnswer [16]bool

	var f [16]func(state bool)
	f[0] = func(state bool) { gotAnswer[0] = state }
	f[1] = func(state bool) { gotAnswer[1] = state }
	f[2] = func(state bool) { gotAnswer[2] = state }
	f[3] = func(state bool) { gotAnswer[3] = state }
	f[4] = func(state bool) { gotAnswer[4] = state }
	f[5] = func(state bool) { gotAnswer[5] = state }
	f[6] = func(state bool) { gotAnswer[6] = state }
	f[7] = func(state bool) { gotAnswer[7] = state }
	f[8] = func(state bool) { gotAnswer[8] = state }
	f[9] = func(state bool) { gotAnswer[9] = state }
	f[10] = func(state bool) { gotAnswer[10] = state }
	f[11] = func(state bool) { gotAnswer[11] = state }
	f[12] = func(state bool) { gotAnswer[12] = state }
	f[13] = func(state bool) { gotAnswer[13] = state }
	f[14] = func(state bool) { gotAnswer[14] = state }
	f[15] = func(state bool) { gotAnswer[15] = state }

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *SixteenSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	// start with off switches
	addend1Switches, _ := NewSixteenSwitchBank("0000000000000000")
	addend2Switches, _ := NewSixteenSwitchBank("0000000000000000")
	carryInSwitch := NewSwitch(false)

	a := NewSixteenBitAdder(addend1Switches.AsPwrEmitters(), addend2Switches.AsPwrEmitters(), carryInSwitch)

	for i, s := range a.Sums {
		s.WireUp(f[i])
	}
	a.CarryOut.WireUp(func(state bool) { gotCarryOut = state })

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

func BenchmarkNewSixteenBitAdder(b *testing.B) {
	benchmarks := []struct {
		bytes1         string
		bytes2         string
		carryInPowered bool
	}{
		{"0000000000000000", "0000000000000000", false},
		{"1111111111111111", "1111111111111111", false},
		{"0000000000000000", "0000000000000000", true},
		{"1111111111111111", "1111111111111111", true},
	}

	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", bm.bytes1, bm.bytes2, bm.carryInPowered), func(b *testing.B) {
			carryInSwitch := NewSwitch(bm.carryInPowered)
			addend1Switches, _ := NewSixteenSwitchBank(bm.bytes1)
			addend2Switches, _ := NewSixteenSwitchBank(bm.bytes2)
			addend1BitPubs := addend1Switches.AsPwrEmitters()
			addend2BitPubs := addend2Switches.AsPwrEmitters()
			for i := 0; i < b.N; i++ {
				NewSixteenBitAdder(addend1BitPubs, addend2BitPubs, carryInSwitch)
			}
		})
	}
}

func BenchmarkSixteenBitAdder_AsAnswerString(b *testing.B) {
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
		addend1Switches, _ := NewSixteenSwitchBank(bm.bytes1)
		addend2Switches, _ := NewSixteenSwitchBank(bm.bytes2)
		addend1BitPubs := addend1Switches.AsPwrEmitters()
		addend2BitPubs := addend2Switches.AsPwrEmitters()
		a := NewSixteenBitAdder(addend1BitPubs, addend2BitPubs, carryInSwitch)
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

		for _, b := range bits {
			pubs = append(pubs, NewSwitch(b == '1'))
		}

		return pubs
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Executing complementer against %s with signal of %t", tc.bits, tc.signalIsPowered), func(t *testing.T) {
			c := NewOnesComplementer(getInputs(tc.bits), NewSwitch(tc.signalIsPowered))

			if c == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
			}

			if got := c.AsComplementString(); got != tc.want {
				t.Errorf(fmt.Sprintf("Wanted %s, but got %s", tc.want, got))
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

		for _, b := range bits {
			pubs = append(pubs, NewSwitch(b == '1'))
		}

		return pubs
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Executing complementer against %s with signal of %t", tc.bits, tc.signalIsPowered), func(t *testing.T) {
			c := NewOnesComplementer(getInputs(tc.bits), NewSwitch(tc.signalIsPowered))

			if c == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
			}

			for i, pub := range c.Complements {
				got := pub.(*XORGate).GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("At index %d, wanted %v but got %v", i, want, got))
				}
			}
		})
	}
}

func TestEightBitSubtracter_AsAnswerString(t *testing.T) {
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
	updateSwitches := func(switchBank *EightSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	// start with off switches
	minuendwitches, _ := NewEightSwitchBank("00000000")
	subtrahendSwitches, _ := NewEightSwitchBank("00000000")

	s := NewEightBitSubtracter(minuendwitches.AsPwrEmitters(), subtrahendSwitches.AsPwrEmitters())

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Subtracting %s from %s", tc.subtrahend, tc.minuend), func(t *testing.T) {

			updateSwitches(minuendwitches, tc.minuend)
			updateSwitches(subtrahendSwitches, tc.subtrahend)

			if s == nil {
				t.Error("Expected an subtractor to return due to good inputs, but gotAnswer c nil one.")
				return // cannot continue tests if no subtractor to test
			}

			if gotAnswer := s.AsAnswerString(); gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but gotAnswer %s", tc.wantAnswer, gotAnswer)
			}

			if gotCarryOut := s.CarryOutAsBool(); gotCarryOut != tc.wantCarryOut {
				t.Errorf("Wanted carry out %t, but gotAnswer %t", tc.wantCarryOut, gotCarryOut)
			}
		})
	}
}

func TestEightBitSubtracter_AnswerViaRegistration(t *testing.T) {
	wantCarryOut := true
	var gotCarryOut bool

	wantAnswer := [8]bool{false, false, false, false, false, false, true, true}
	var gotAnswer [8]bool

	var f [8]func(state bool)
	f[0] = func(state bool) { gotAnswer[0] = state }
	f[1] = func(state bool) { gotAnswer[1] = state }
	f[2] = func(state bool) { gotAnswer[2] = state }
	f[3] = func(state bool) { gotAnswer[3] = state }
	f[4] = func(state bool) { gotAnswer[4] = state }
	f[5] = func(state bool) { gotAnswer[5] = state }
	f[6] = func(state bool) { gotAnswer[6] = state }
	f[7] = func(state bool) { gotAnswer[7] = state }

	// flip switches to match bit pattern
	updateSwitches := func(switchBank *EightSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	// start with off switches
	minuendSwitches, _ := NewEightSwitchBank("00000000")
	subtrahendSwitches, _ := NewEightSwitchBank("00000000")

	s := NewEightBitSubtracter(minuendSwitches.AsPwrEmitters(), subtrahendSwitches.AsPwrEmitters())

	for i, s := range s.Differences {
		s.WireUp(f[i])
	}

	s.CarryOut.WireUp(func(state bool) { gotCarryOut = state })

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

			o := NewOscillator(tc.initState)
			o.Oscillate(tc.oscHertz)

			o.WireUp(func(state bool) {
				if state {
					gotResults += "T"
				} else {
					gotResults += "F"
				}
			})

			time.Sleep(time.Second * 2)

			o.Stop()

			if gotResults != tc.wantResults {
				t.Errorf(fmt.Sprintf("Wanted results %s but got %s.", tc.wantResults, gotResults))
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
		t.Errorf(fmt.Sprintf("Wanted power of %t at Qs, but got %t.", false, gotQ))
	}

	if gotQBar := ff.QBar.GetIsPowered(); gotQBar != true {
		t.Errorf(fmt.Sprintf("Wanted power of %t at QBar, but got %t.", true, gotQBar))
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
				t.Errorf(fmt.Sprintf("Wanted power of %t at Qs, but got %t.", tc.wantQ, gotQ))
			}

			if gotQBar := ff.QBar.GetIsPowered(); gotQBar != tc.wantQBar {
				t.Errorf(fmt.Sprintf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar))
			}
		})
	}
}

func TestRSFlipFlop_Panic(t *testing.T) {
	want := "A Flip-Flop cannot have equivalent power status at both Qs and QBar"

	defer func() {
		if got := recover(); !strings.HasPrefix(got.(string), want) {
			t.Errorf(fmt.Sprintf("Expected a panic of \"%s\" but got \"%s\"", want, got))
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
	}{ // construction of the latch will start with a default of clkIn:true, dataIn:true, which causes Qs on (QBar off)
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
				t.Errorf(fmt.Sprintf("Wanted power of %t at Qs, but got %t.", tc.wantQ, gotQ))
			}

			if gotQBar := latch.QBar.GetIsPowered(); gotQBar != tc.wantQBar {
				t.Errorf(fmt.Sprintf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar))
			}
		})
	}
}

func TestEightBitLatch(t *testing.T) {
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
	updateSwitches := func(switchBank *EightSwitchBank, bits string) {
		for i, b := range bits {
			switchBank.Switches[i].Set(b == '1')
		}
	}

	latchSwitches, _ := NewEightSwitchBank("00000000")
	clkSwitch := NewSwitch(false)
	latch := NewEightBitLatch(clkSwitch, latchSwitches.AsPwrEmitters())

	priorWant := [8]bool{false, false, false, false, false, false, false, false}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Stage %d: Setting switches to %s", i+1, tc.input), func(t *testing.T) {

			// set to OFF to test that nothing will change in the latch store

			clkSwitch.Set(false)
			updateSwitches(latchSwitches, tc.input)

			// try as actual Qs
			for i, q := range latch.Qs {
				got := q.GetIsPowered()
				want := priorWant[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As Q] At index %d, with clkSwitch off, wanted %v but got %v", i, want, got))
				}
			}

			// now try AsPwrEmitters
			for i, pwr := range latch.AsPwrEmitters() {
				got := pwr.(*NORGate).GetIsPowered()
				want := priorWant[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As PwrEmitter] At index %d, with clkSwitch off, wanted %v but got %v", i, want, got))
				}
			}

			// Now set to ON to test that requested changes did occur in the latch store

			clkSwitch.Set(true)

			// try as actual Qs
			for i, q := range latch.Qs {
				got := q.GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As Q] At index %d, with clkSwitch on, wanted %v but got %v", i, want, got))
				}
			}

			// now try AsPwrEmitters
			for i, pwr := range latch.AsPwrEmitters() {
				got := pwr.(*NORGate).GetIsPowered()
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("[As PwrEmitter] At index %d, with clkSwitch on, wanted %v but got %v", i, want, got))
				}
			}

			// now update the prior tracker bools to ensure next pass (with cklIn as OFF at the top) proves it didn't change (so matches prior)
			for i, q := range latch.Qs {
				priorWant[i] = q.GetIsPowered()
			}
		})
	}
}
