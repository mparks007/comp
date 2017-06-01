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

	p.Subscribe(func(state bool) { got1 = state })
	p.Subscribe(func(state bool) { got2 = state })

	p.Publish(true)
	want = true

	if got1 != want {
		t.Errorf(fmt.Sprintf("Exected subscription 1 to be %t but got %t", want, got1))
	}

	if got2 != want {
		t.Errorf(fmt.Sprintf("Exected subscription 2 to be %t but got %t", want, got2))
	}

	p.Publish(false)
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

	b.Subscribe(func(state bool) { got = state })

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

	// setup a subscription to the switch's events
	s.Subscribe(func(state bool) {
		gotState = state
		gotCount += 1
	})

	// initial turn on
	s.TurnOn()
	wantState = true
	wantCount = 1

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an off switch turned on, wanted the subscriber's state to be %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an off switch turned on, wanted the subscriber's call count to be %d but got %d", wantCount, gotCount))
	}

	// turn on again though already on
	s.TurnOn()
	wantCount = 1

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an attempt to turn on an already on switch, wanted the subscriber's call count to remain %d but got %d", wantCount, gotCount))
	}
}

func TestSwitch_TurnOff(t *testing.T) {
	var wantState, gotState bool
	var wantCount, gotCount int

	// start with switch being on
	s := NewSwitch(true)

	// setup a subscription to the switch's events
	s.Subscribe(func(state bool) {
		gotState = state
		gotCount += 1
	})

	// initial turn off
	s.TurnOff()
	wantState = false
	wantCount = 1

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an on switch turned off, wanted the subscriber's state to be %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an on switch turned off, wanted the subscriber's call count to be %d but got %d", wantCount, gotCount))
	}

	// turn off again though already on
	s.TurnOff()
	wantCount = 1

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an attempt to turn off an already off switch, wanted the subscriber's call count to remain %d but got %d", wantCount, gotCount))
	}
}

func TestSwitch_Toggle(t *testing.T) {
	var wantState, gotState bool
	var wantCount, gotCount int

	// start with switch being off
	s := NewSwitch(false)

	// setup a subscription to the switch's events
	s.Subscribe(func(state bool) {
		gotState = state
		gotCount += 1
	})

	// initial toggle on
	s.Toggle()
	wantState = true
	wantCount = 1

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an off switch toggled, wanted the subscriber's state to be %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("With an off switch toggled, wanted the subscriber's call count to be %d but got %d", wantCount, gotCount))
	}

	// now toggle off
	s.Toggle()
	wantState = false
	wantCount = 2

	if gotState != wantState {
		t.Errorf(fmt.Sprintf("With an on switch toggled, wanted the subscriber's state to be %t but got %t", wantState, gotState))
	}

	if gotCount != wantCount {
		t.Errorf(fmt.Sprintf("When toggling an on switch again, wanted the subscriber's call count to increment to %d but got %d", wantCount, gotCount))
	}
}

func TestRelay2(t *testing.T) {
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
		t.Run(fmt.Sprintf("Setting A power to %t and B power to %t", tc.aInPowered, tc.bInPowered), func(t *testing.T) {
			var gotOpenOut, gotClosedOut bool

			aSwitch := NewSwitch(!tc.aInPowered)
			bSwitch := NewSwitch(!tc.bInPowered)

			r := NewRelay2(aSwitch, bSwitch)

			r.OpenOut.Subscribe(func(state bool) { gotOpenOut = state })
			r.ClosedOut.Subscribe(func(state bool) { gotClosedOut = state })

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
