package circuit

import (
	"fmt"
	"testing"
)

// go test
// go test -race -v (verbose)
// go test -race -cpu=1,2,4 (go max prox)
// go test -v

func TestPublication(t *testing.T) {
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

/*
func TestEightBitPublication(t *testing.T) {
	var want, got1, got2 bool

	p := &eightBitPublication{}

	p.Register(func(state bool) { got1 = state })
	p.Register(func(state bool) { got2 = state })

	p.isPowered = true
	p.Publish()
	want = true

	if got1 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 1 to be %t but got %t", want, got1))
	}

	if got2 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 2 to be %t but got %t", want, got2))
	}

	p.isPowered = false
	p.Publish()
	want = false

	if got1 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 1 to be %t but got %t", want, got1))
	}

	if got2 != want {
		t.Errorf(fmt.Sprintf("Expected subscription 2 to be %t but got %t", want, got2))
	}
}
*/

func TestBattery(t *testing.T) {
	var want, got bool

	b := &Battery{}

	b.Register(func(state bool) { got = state })

	want = true

	if got != want {
		t.Errorf(fmt.Sprintf("With a battery, wanted the subscriber to see power as %t but got %t", want, got))
	}
}

func TestSwitch_Set(t *testing.T) {
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
	s.Publish(true)
	wantState = true
	wantCount = 2

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an off switch turned on, wanted the subscriber to see power as %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an off switch turned on, wanted the subscriber to see power as %d but got %d", wantCount, gotCount))
	}

	// turn on again though already on
	s.Publish(true)
	wantCount = 2

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an attempt to turn on an already on switch, wanted the subscriber's call count to remain %d but got %d", wantCount, gotCount))
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
				got := s.IsPowered()
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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)

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
		aIn          bitPublisher
		bIn          bitPublisher
		wantAtOpen   bool
		wantAtClosed bool
	}{
		{nil, nil, false, false},
		{&Battery{}, nil, true, false},
		{nil, &Battery{}, false, false},
		{&Battery{}, &Battery{}, false, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			var gotOpenOut, gotClosedOut bool

			r := NewRelay2(tc.aIn, tc.bIn)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)
			cSwitch.Publish(tc.cInPowered)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)
			cSwitch.Publish(tc.cInPowered)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)
			cSwitch.Publish(tc.cInPowered)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)

			if got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestInverter2(t *testing.T) {
	testCases := []struct {
		in      bitPublisher
		wantOut bool
	}{
		{nil, true},
		{&Battery{}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Input as %T.", tc.in), func(t *testing.T) {

			i := NewInverter(tc.in)

			var got bool
			i.Register(func(state bool) { got = state })

			if got != tc.wantOut {
				t.Errorf("Power prior was %T so wanted it inverted to %T but got %T", tc.in, tc.wantOut, got)
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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)

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

			aSwitch.Publish(tc.aInPowered)
			bSwitch.Publish(tc.bInPowered)
			cSwitch.Publish(tc.carryInPowered)

			if gotSum != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, gotSum)
			}

			if gotCarry != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, gotCarry)
			}
		})
	}
}

func TestEightBitAdder2_GoodInputs(t *testing.T) {
	testCases := []struct {
		byte1        string
		byte2        string
		carryIn      emitter
		wantAnswer   string
		wantCarryOut bool
	}{
		{"00000000", "00000000", nil, "00000000", false},
		{"00000001", "00000000", nil, "00000001", false},
		{"00000000", "00000001", nil, "00000001", false},
		{"00000000", "00000000", &Battery{}, "00000001", false},
		{"00000001", "00000000", &Battery{}, "00000010", false},
		{"00000000", "00000001", &Battery{}, "00000010", false},
		{"10000000", "10000000", nil, "100000000", true},
		{"10000001", "10000000", nil, "100000001", true},
		{"11111111", "11111111", nil, "111111110", true},
		{"11111111", "11111111", &Battery{}, "111111111", true},
		{"01111111", "11111111", nil, "101111110", true},
		{"01111111", "11111111", &Battery{}, "101111111", true},
		{"10101010", "01010101", nil, "11111111", false},
		{"10101010", "01010101", &Battery{}, "100000000", true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %T", tc.byte1, tc.byte2, tc.carryIn), func(t *testing.T) {
			minuendBits, _ := NewEightSwitchBank(tc.byte1)
			subtrahendBits, _ := NewEightSwitchBank(tc.byte2)

			a := NewEightBitAdder2(minuendBits.AsBitPublishers(), subtrahendBits.AsBitPublishers(), tc.carryIn)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
				return // switchOn error, expecting to have a nil adder here so cannot do further tests using one
			}

			if a == nil {
				t.Error("Expected an adder to return due to good inputs, but got a nil one.")
				return // cannot continue tests if no adder to test
			}

			if got := a.AsString(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := a.carryOut.Emitting(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
	}
}
