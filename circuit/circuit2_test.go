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

	p := &publication{}

	p.Register(func(state bool) { got1 = state })
	p.Register(func(state bool) { got2 = state })

	p.state = true
	p.Publish()
	want = true

	if got1 != want {
		t.Errorf(fmt.Sprintf("Exected subscription 1 to be %t but got %t", want, got1))
	}

	if got2 != want {
		t.Errorf(fmt.Sprintf("Exected subscription 2 to be %t but got %t", want, got2))
	}

	p.state = false
	p.Publish()
	want = false

	if got1 != want {
		t.Errorf(fmt.Sprintf("Exected subscription 1 to be %t but got %t", want, got1))
	}

	if got2 != want {
		t.Errorf(fmt.Sprintf("Exected subscription 2 to be %t but got %t", want, got2))
	}
}

func TestBattery(t *testing.T) {
	var want, got bool

	b := &Battery{}

	b.Register(func(state bool) { got = state })

	b.Charge()
	want = true

	if got != want {
		t.Errorf(fmt.Sprintf("With a charged battery, wanted the subscriber's state to be %t but got %t", want, got))
	}

	b.Discharge()
	want = false

	if got != want {
		t.Errorf(fmt.Sprintf("With a discharged battery, wanted the subscriber's state to be %t but got %t", want, got))
	}
}

func TestSwitch_TurnOn(t *testing.T) {
	var wantState, gotState bool
	var wantCount, gotCount int

	// start with switch being off
	s := NewSwitch(false)

	// register callback (will trigger immediate call to push state at time of registration)
	s.Register(func(state bool) {
		gotState = state
		gotCount += 1
	})

	// initial turn on
	s.TurnOn()
	wantState = true
	wantCount = 2

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an off switch turned on, wanted the subscriber's state to be %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an off switch turned on, wanted the subscriber's call count to be %d but got %d", wantCount, gotCount))
	}

	// turn on again though already on
	s.TurnOn()
	wantCount = 2

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an attempt to turn on an already on switch, wanted the subscriber's call count to remain %d but got %d", wantCount, gotCount))
	}
}

func TestSwitch_TurnOff(t *testing.T) {
	var wantState, gotState bool
	var wantCount, gotCount int

	// start with switch being on
	s := NewSwitch(true)

	// register callback (will trigger immediate call to push state at time of registration)
	s.Register(func(state bool) {
		gotState = state
		gotCount += 1
	})

	// initial turn off
	s.TurnOff()
	wantState = false
	wantCount = 2

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an on switch turned off, wanted the subscriber's state to be %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an on switch turned off, wanted the subscriber's call count to be %d but got %d", wantCount, gotCount))
	}

	// turn off again though already on
	s.TurnOff()
	wantCount = 2

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an attempt to turn off an already off switch, wanted the subscriber's call count to remain %d but got %d", wantCount, gotCount))
	}
}

func TestSwitch_Toggle(t *testing.T) {
	var wantState, gotState bool
	var wantCount, gotCount int

	// start with switch being off
	s := NewSwitch(false)

	// register callback (will trigger immediate call to push state at time of registration)
	s.Register(func(state bool) {
		gotState = state
		gotCount += 1
	})

	// initial toggle on
	s.Toggle()
	wantState = true
	wantCount = 2

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an off switch toggled, wanted the subscriber's state to be %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an off switch toggled, wanted the subscriber's call count to be %d but got %d", wantCount, gotCount))
	}

	// now toggle off
	s.Toggle()
	wantState = false
	wantCount = 3

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an on switch toggled, wanted the subscriber's state to be %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("When toggling an on switch again, wanted the subscriber's call count to increment to %d but got %d", wantCount, gotCount))
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

			if tc.aInPowered {
				aSwitch.TurnOn()
			} else {
				aSwitch.TurnOff()
			}

			if tc.bInPowered {
				bSwitch.TurnOn()
			} else {
				bSwitch.TurnOff()
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

func TestRelay2_WithBatteries(t *testing.T) {
	testCases := []struct {
		aIn          publisher
		bIn          publisher
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

			if tc.aInPowered {
				aSwitch.TurnOn()
			} else {
				aSwitch.TurnOff()
			}

			if tc.bInPowered {
				bSwitch.TurnOn()
			} else {
				bSwitch.TurnOff()
			}

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

			if tc.aInPowered {
				aSwitch.TurnOn()
			} else {
				aSwitch.TurnOff()
			}

			if tc.bInPowered {
				bSwitch.TurnOn()
			} else {
				bSwitch.TurnOff()
			}

			if tc.cInPowered {
				cSwitch.TurnOn()
			} else {
				cSwitch.TurnOff()
			}

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

			if tc.aInPowered {
				aSwitch.TurnOn()
			} else {
				aSwitch.TurnOff()
			}

			if tc.bInPowered {
				bSwitch.TurnOn()
			} else {
				bSwitch.TurnOff()
			}

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

			if tc.aInPowered {
				aSwitch.TurnOn()
			} else {
				aSwitch.TurnOff()
			}

			if tc.bInPowered {
				bSwitch.TurnOn()
			} else {
				bSwitch.TurnOff()
			}

			if tc.cInPowered {
				cSwitch.TurnOn()
			} else {
				cSwitch.TurnOff()
			}

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

			if tc.aInPowered {
				aSwitch.TurnOn()
			} else {
				aSwitch.TurnOff()
			}

			if tc.bInPowered {
				bSwitch.TurnOn()
			} else {
				bSwitch.TurnOff()
			}

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

			if tc.aInPowered {
				aSwitch.TurnOn()
			} else {
				aSwitch.TurnOff()
			}

			if tc.bInPowered {
				bSwitch.TurnOn()
			} else {
				bSwitch.TurnOff()
			}

			if tc.cInPowered {
				cSwitch.TurnOn()
			} else {
				cSwitch.TurnOff()
			}

			if got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}
