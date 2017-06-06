package circuit

import (
	"fmt"
	"testing"
)

// go test
// go test -race -v (verbose)
// go test -race -cpu=1,2,4 (go max prox)
// go test -v

func TestBitPublication(t *testing.T) {
	var want, got1, got2 bool

	p := &bitPublication{}

	p.Register(func(state bool) { got1 = state })
	p.Register(func(state bool) { got2 = state })

	p.Publish(true)
	want = true

	if got1 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 1 to be %t but got %t", want, got1))
	}

	if got2 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 2 to be %t but got %t", want, got2))
	}

	p.Publish(false)
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

	b.Register(func(state bool) { got = state })

	want = true

	if got != want {
		t.Errorf(fmt.Sprintf("With a new battery, wanted the subscriber to see power as %t but got %t", want, got))
	}

	b.Discharge()
	want = false

	if got != want {
		t.Errorf(fmt.Sprintf("With a discharged battery, wanted the subscriber'c IsPowered to be %t but got %t", want, got))
	}

	b.Charge()
	want = true

	if got != want {
		t.Errorf(fmt.Sprintf("With a charged battery, wanted the subscriber'c IsPowered to be %t but got %t", want, got))
	}
}

func TestSwitch(t *testing.T) {
	var wantState, gotState bool
	var wantCount, gotCount int

	// start with switch being off
	s := NewSwitch(false)

	// register callback (will trigger immediate call to push isPowered at time of registration)
	s.Register(func(state bool) {
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

			for i, s := range sb.Switches {
				got := s.isPowered
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("At index %d, wanted %v but got %v", i, want, got))
				}
			}
		})
	}
}

func TestEightSwitchBank_AsBitPublishers(t *testing.T) {
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

			for i, pub := range sb.AsBitPublishers() {
				got := pub.(*Switch).isPowered
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("At index %d, wanted %v but got %v", i, want, got))
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

			for i, s := range sb.Switches {
				got := s.isPowered
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("At index %d, wanted %v but got %v", i, want, got))
				}
			}
		})
	}
}

func TestSixteenSwitchBank_AsBitPublishers(t *testing.T) {
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

			for i, pub := range sb.AsBitPublishers() {
				got := pub.(*Switch).isPowered
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("At index %d, wanted %v but got %v", i, want, got))
				}
			}
		})
	}
}

func TestRelay2_WithSwitches(t *testing.T) {
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

	r := NewRelay2(aSwitch, bSwitch)

	var gotOpenOut, gotClosedOut bool
	r.OpenOut.Register(func(state bool) { gotOpenOut = state })
	r.ClosedOut.Register(func(state bool) { gotClosedOut = state })

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

func TestRelay2_WithBatteries(t *testing.T) {
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

			r := NewRelay2(pin1Battery, pin2Battery)

			r.OpenOut.Register(func(state bool) { gotOpenOut = state })
			r.ClosedOut.Register(func(state bool) { gotClosedOut = state })

			if gotOpenOut != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut)
			}

			if gotClosedOut != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut)
			}
		})
	}
}

func TestANDGate2_TwoPin(t *testing.T) {
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

	g := NewANDGate2(aSwitch, bSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestANDGate2_ThreePin(t *testing.T) {
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

	g := NewANDGate2(aSwitch, bSwitch, cSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestORGate2_TwoPin(t *testing.T) {
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

	g := NewORGate2(aSwitch, bSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestORGate2_ThreePin(t *testing.T) {
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

	g := NewORGate2(aSwitch, bSwitch, cSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestNANDGate2_TwoPin(t *testing.T) {
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

	g := NewNANDGate2(aSwitch, bSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestNANDGate2_ThreePin(t *testing.T) {
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

	g := NewNANDGate2(aSwitch, bSwitch, cSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestNORGate2_TwoPin(t *testing.T) {
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

	g := NewNORGate2(aSwitch, bSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestNORGate2_ThreePin(t *testing.T) {
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

	g := NewNORGate2(aSwitch, bSwitch, cSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestXORGate2(t *testing.T) {
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

	g := NewXORGate2(aSwitch, bSwitch)

	var got bool
	g.Register(func(state bool) { got = state })

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

func TestXNORGate2(t *testing.T) {
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
	g.Register(func(state bool) { got = state })

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

func TestInverter2(t *testing.T) {
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

			i := NewInverter(pin1Battery)

			var got bool
			i.Register(func(state bool) { got = state })

			if got != tc.wantOut {
				t.Errorf("Input power was %t so wanted it inverted to %t but got %t", tc.inPowered, tc.wantOut, got)
			}
		})
	}
}

func TestHalfAdder2(t *testing.T) {
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

	h := NewHalfAdder2(aSwitch, bSwitch)

	var gotSum, gotCarry bool
	h.Sum.Register(func(state bool) { gotSum = state })
	h.Carry.Register(func(state bool) { gotCarry = state })

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

func TestFullAdder2(t *testing.T) {
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

	h := NewFullAdder2(aSwitch, bSwitch, cSwitch)

	var gotSum, gotCarry bool
	h.Sum.Register(func(state bool) { gotSum = state })
	h.Carry.Register(func(state bool) { gotCarry = state })

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

func TestEightBitAdder2_PostRunAnswer(t *testing.T) {
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

	a := NewEightBitAdder2(addend1Switches.AsBitPublishers(), addend2Switches.AsBitPublishers(), carryInSwitch)

	if a == nil {
		t.Error("Expected an adder to return due to good inputs, but got a nil one.")
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", tc.byte1, tc.byte2, tc.carryInPowered), func(t *testing.T) {

			updateSwitches(addend1Switches, tc.byte1)
			updateSwitches(addend2Switches, tc.byte2)
			carryInSwitch.Set(tc.carryInPowered)

			if got := a.AnswerAsString(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := a.CarryOutAsBool(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
	}
}

func TestEightBitAdder2_RegistrationAnswer(t *testing.T) {
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

	a := NewEightBitAdder2(addend1Switches.AsBitPublishers(), addend2Switches.AsBitPublishers(), carryInSwitch)

	for i, s := range a.Sums {
		s.Register(f[i])
	}
	a.CarryOut.Register(func(state bool) { gotCarryOut = state })

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

func TestSixteenBitAdder_PostRunAnswer(t *testing.T) {
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

	a := NewSixteenBitAdder2(addend1Switches.AsBitPublishers(), addend2Switches.AsBitPublishers(), carryInSwitch)

	if a == nil {
		t.Error("Expected an adder to return due to good inputs, but got a nil one.")
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", tc.bytes1, tc.bytes2, tc.carryInPowered), func(t *testing.T) {

			updateSwitches(addend1Switches, tc.bytes1)
			updateSwitches(addend2Switches, tc.bytes2)
			carryInSwitch.Set(tc.carryInPowered)

			if got := a.AnswerAsString(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := a.CarryOutAsBool(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
	}
}

func TestSixteenBitAdder2_RegistrationAnswer(t *testing.T) {
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

	a := NewSixteenBitAdder2(addend1Switches.AsBitPublishers(), addend2Switches.AsBitPublishers(), carryInSwitch)

	for i, s := range a.Sums {
		s.Register(f[i])
	}
	a.CarryOut.Register(func(state bool) { gotCarryOut = state })

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
			addend1BitPubs := addend1Switches.AsBitPublishers()
			addend2BitPubs := addend2Switches.AsBitPublishers()
			for i := 0; i < b.N; i++ {
				NewSixteenBitAdder2(addend1BitPubs, addend2BitPubs, carryInSwitch)
			}
		})
	}
}

func BenchmarkSixteenBitAdder_AnswerAsString(b *testing.B) {
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
		addend1BitPubs := addend1Switches.AsBitPublishers()
		addend2BitPubs := addend2Switches.AsBitPublishers()
		a := NewSixteenBitAdder2(addend1BitPubs, addend2BitPubs, carryInSwitch)
		b.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", bm.bytes1, bm.bytes2, bm.carryInPowered), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				a.AnswerAsString()
			}
		})
	}
}

func TestOnesCompliment2_AnswerAsString(t *testing.T) {

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

	getInputs := func(bits string) []bitPublisher {
		pubs := []bitPublisher{}

		for _, b := range bits {
			pubs = append(pubs, NewSwitch(b == '1'))
		}

		return pubs
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Executing complementer against %s with signal of %t", tc.bits, tc.signalIsPowered), func(t *testing.T) {
			c := NewOnesComplementer2(getInputs(tc.bits), NewSwitch(tc.signalIsPowered))

			if c == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
			}

			if got := c.AsAnswerString(); got != tc.want {
				t.Errorf(fmt.Sprintf("Wanted %s, but got %s", tc.want, got))
			}
		})
	}
}

func TestOnesCompliment2_AsBitPublishers(t *testing.T) {

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

	getInputs := func(bits string) []bitPublisher {
		pubs := []bitPublisher{}

		for _, b := range bits {
			pubs = append(pubs, NewSwitch(b == '1'))
		}

		return pubs
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Executing complementer against %s with signal of %t", tc.bits, tc.signalIsPowered), func(t *testing.T) {
			c := NewOnesComplementer2(getInputs(tc.bits), NewSwitch(tc.signalIsPowered))

			if c == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
			}

			for i, pub := range c.AsBitPublishers() {
				got := pub.(*XORGate2).isPowered
				want := tc.want[i]

				if got != want {
					t.Errorf(fmt.Sprintf("At index %d, wanted %v but got %v", i, want, got))
				}
			}
		})
	}
}
