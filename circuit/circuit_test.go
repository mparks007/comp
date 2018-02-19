package circuit

import (
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// go test
// go test -race -v (verbose)
// go test -race -cpu=1,2,4 (go max procs)
// go test -count 100
// go test -v
// go test -run TestOscillator (specific test)
// go test -run TestOscillator -count 100 -v (multi options)
// go test -run TestRelay_WithBatteries -count 50 -trace out2.txt (go tool trace out2.txt)

// setSwitches will flip the switches of a SwitchBank to match a passed in bits string
func setSwitches(switchBank *NSwitchBank, bits string) {
	for i, b := range bits {
		switchBank.Switches[i].(*Switch).Set(b == '1')
	}
}

func TestPwrsource(t *testing.T) {
	var want, got1, got2 bool
	ch1 := make(chan bool, 1)
	ch2 := make(chan bool, 1)

	pwr := &pwrSource{}
	pwr.Init()

	// two wire ups to prove both will get called
	pwr.WireUp(ch1)
	pwr.WireUp(ch2)

	want = false

	// test default state (unpowered)
	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// test power transmit
	want = true
	pwr.Transmit(want)
	<-pwr.chTransmitted

	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// test transmit loss of power
	want = false
	pwr.Transmit(want)
	<-pwr.chTransmitted

	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// test transmitting same state as last time (should skip it)
	pwr.Transmit(want)
	<-pwr.chTransmitted

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did.")
	case <-time.After(time.Millisecond * 5):
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did.")
	case <-time.After(time.Millisecond * 5):
	}
}

func TestWire_NoDelay(t *testing.T) {
	var want, got1, got2 bool
	ch1 := make(chan bool, 1)
	ch2 := make(chan bool, 1)

	wire := NewWire(0)
	defer wire.Shutdown()

	// two wire ups to prove both will get called
	wire.WireUp(ch1)
	wire.WireUp(ch2)

	want = false

	// test default state (unpowered)
	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// test power transmit
	want = true
	wire.Transmit(want)
	<-wire.chTransmitted

	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// test transmit loss of power
	want = false
	wire.Transmit(want)
	<-wire.chTransmitted

	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// test transmitting same state as last time (should skip it)
	wire.Transmit(want)
	<-wire.chTransmitted

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did.")
	case <-time.After(time.Millisecond * 5):
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did.")
	case <-time.After(time.Millisecond * 5):
	}
}

func TestWire_WithDelay(t *testing.T) {
	var want, got1, got2 bool
	var wireLen uint = 100
	ch1 := make(chan bool, 1)
	ch2 := make(chan bool, 1)

	wire := NewWire(wireLen)
	defer wire.Shutdown()

	// two wire ups to prove both will get called
	wire.WireUp(ch1)
	wire.WireUp(ch2)

	want = false

	// test default state (unpowered)
	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// test power transmit
	want = true
	start := time.Now()
	wire.Transmit(want)
	end := time.Now()
	<-wire.chTransmitted

	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// validate wire delay
	gotDuration := end.Sub(start) // + time.Millisecond*1 // adding in just a little more to avoid timing edge case
	wantDuration := time.Millisecond * time.Duration(wireLen)
	if gotDuration < wantDuration {
		t.Errorf("Wire power on transmit time should have been %v but was %v", wantDuration, gotDuration)
	}

	// test loss of power transmit
	want = false
	start = time.Now()
	wire.Transmit(want)
	end = time.Now()
	<-wire.chTransmitted

	if got1 = <-ch1; got1 != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1)
	}
	if got2 = <-ch2; got2 != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2)
	}

	// validate wire delay
	gotDuration = end.Sub(start) // + time.Millisecond*1 // adding in just a little more to avoid timing edge case
	wantDuration = time.Millisecond * time.Duration(wireLen)
	if gotDuration < wantDuration {
		t.Errorf("Wire power off transmit time should have been %v but was %v", wantDuration, gotDuration)
	}

	// test transmitting same state as last time (should skip it)
	wire.Transmit(want)
	<-wire.chTransmitted

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did.")
	case <-time.After(time.Millisecond * 5):
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did.")
	case <-time.After(time.Millisecond * 5):
	}
}

func TestRibbonCable(t *testing.T) {
	var want, got bool
	ch1 := make(chan bool, 1)
	ch2 := make(chan bool, 1)

	rib := NewRibbonCable(2, 0)
	defer rib.Shutdown()

	rib.SetInputs(NewBattery(false), NewBattery(true))

	rib.Wires[0].(*Wire).WireUp(ch1)
	rib.Wires[1].(*Wire).WireUp(ch2)

	want = false
	if got = <-ch1; got != want {
		t.Errorf("Left Switch off, wanted the wire to see power as %t but got %t", want, got)
	}

	want = true
	if got = <-ch2; got != want {
		t.Errorf("Right Switch on, wanted the wire to see power as %t but got %t", want, got)
	}
}

func TestBattery(t *testing.T) {
	var want, got bool
	ch := make(chan bool, 1)

	bat := NewBattery(true)
	bat.WireUp(ch)
	want = true

	// test default battery state (powered)
	if got = <-ch; got != want {
		t.Errorf("With a new battery, wanted the subscriber to see power as %t but got %t", want, got)
	}

	// test loss of power
	bat.Discharge()
	<-bat.chTransmitted
	want = false

	if got = <-ch; got != want {
		t.Errorf("With a discharged battery, wanted the subscriber's IsPowered to be %t but got %t", want, got)
	}

	// test re-added power
	bat.Charge()
	<-bat.chTransmitted
	want = true

	if got = <-ch; got != want {
		t.Errorf("With a charged battery, wanted the subscriber's IsPowered to be %t but got %t", want, got)
	}
}

func TestRelay_WithBatteries(t *testing.T) {
	testCases := []struct {
		aInPowered   bool
		bInPowered   bool
		wantAtOpen   bool
		wantAtClosed bool
	}{
		// {true, false, true, false},
		// {true, true, false, true},
		// {false, true, false, false},
		// {false, false, false, false},
		// {true, false, true, false},
		// {true, true, false, true},
		// {false, false, false, false},
	}

	var pin1Battery, pin2Battery *Battery
	openCh := make(chan bool, 1)
	closedCh := make(chan bool, 1)

	pin1Battery = NewBattery(true)
	pin2Battery = NewBattery(true)

	rel := NewRelay(pin1Battery, pin2Battery)
	defer rel.Shutdown()

	var gotOpenOut, gotClosedOut atomic.Value
	go func() {
		for {
			select {
			case newOpen := <-openCh:
				gotOpenOut.Store(newOpen)
			case newClosed := <-closedCh:
				gotClosedOut.Store(newClosed)
			}
		}
	}()

	rel.OpenOut.WireUp(openCh)
	rel.ClosedOut.WireUp(closedCh)

	//time.Sleep(time.Millisecond * 25)

	// if gotOpenOut.Load().(bool) != false {
	// 	t.Error("Wanted no power at the open position but got some")
	// }
	// if gotClosedOut.Load().(bool) != true {
	// 	t.Error("Wanted power at the closed position but got none")
	// }

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test#%d: Setting input A to %t and B to %t", i+1, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			if tc.aInPowered {
				pin1Battery.Charge()
			} else {
				pin1Battery.Discharge()
			}
			<-pin1Battery.chTransmitted

			if tc.bInPowered {
				pin2Battery.Charge()
			} else {
				pin2Battery.Discharge()
			}
			<-pin2Battery.chTransmitted
			//time.Sleep(time.Millisecond * 15)

			if gotOpenOut.Load().(bool) != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut.Load().(bool))
			}

			if gotClosedOut.Load().(bool) != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut.Load().(bool))
			}
		})
	}
}

func TestSwitch(t *testing.T) {
	var got, want bool
	ch := make(chan bool, 1)
	receivedCh := make(chan struct{})

	sw := NewSwitch(false)
	defer sw.Shutdown()

	go func() {
		for {
			got = <-ch
			receivedCh <- struct{}{}
		}
	}()

	sw.WireUp(ch)
	<-receivedCh

	want = false

	// test initial off switch state
	if got != want {
		t.Errorf("With an off switch, wanted the subscriber to see power as %t but got %t", want, got)
	}

	// initial turn on
	want = true
	sw.Set(want)
	<-receivedCh

	if got != want {
		t.Errorf("With an off switch turned on, wanted the subscriber to see power as %t but got %t", want, got)
	}

	// turn on again, though already on ('want' is already true from prior Set)
	sw.Set(want)

	if got != want {
		t.Errorf("With an attempt to turn on an already on switch, wanted the channel to be empty, but it wasn't")
	}

	// now off
	want = false
	sw.Set(want)
	<-receivedCh

	if got != want {
		t.Errorf("With an on switch turned off, wanted the subscriber to see power as %t but got %t", want, got)
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
				sb.Shutdown()
			}

			tc.wantError += "\"" + tc.input + "\""

			if err == nil || (err != nil && err.Error() != tc.wantError) {
				t.Errorf("Wanted error \"%s\" but got \"%v\"", tc.wantError, err.Error())
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

	received := make(chan struct{}, 1)

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting switches to %s", tc.input), func(t *testing.T) {
			sb, err := NewNSwitchBank(tc.input)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
			}

			defer sb.Shutdown()

			var got atomic.Value
			ch := make(chan bool, 1)
			go func() {
				for {
					got.Store(<-ch)
					received <- struct{}{}
				}
			}()

			for i, pwr := range sb.Switches {

				pwr.(*Switch).WireUp(ch)
				<-received

				want := tc.want[i]

				// how can I NOT have a real bool here since the <-received should have waited until the bool got stored in the go func for loop!
				if got.Load().(bool) != want {
					t.Errorf("At index %d, wanted %t but got %t", i, want, got.Load().(bool))
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
		{true, false, true, false},
		{true, true, false, true},
		{false, true, false, false},
		{false, false, false, false},
		{true, false, true, false},
		{true, true, false, true},
		{false, false, false, false},
	}

	openCh := make(chan bool, 1)
	closedCh := make(chan bool, 1)

	aSwitch := NewSwitch(true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(true)
	defer bSwitch.Shutdown()

	rel := NewRelay(aSwitch, bSwitch)
	defer rel.Shutdown()

	var gotOpenOut, gotClosedOut atomic.Value
	go func() {
		for {
			select {
			case newOpen := <-openCh:
				gotOpenOut.Store(newOpen)
			case newClosed := <-closedCh:
				gotClosedOut.Store(newClosed)
			}
		}
	}()

	rel.OpenOut.WireUp(openCh)
	rel.ClosedOut.WireUp(closedCh)

	time.Sleep(time.Millisecond * 10)

	if gotOpenOut.Load().(bool) != false {
		t.Error("Wanted no power at the open position but got some")
	}
	if gotClosedOut.Load().(bool) != true {
		t.Error("Wanted power at the closed position but got none")
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip#%d: Setting A power to %t and B power to %t", i+1, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			time.Sleep(time.Millisecond * 20)

			if gotOpenOut.Load().(bool) != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut.Load().(bool))
			}

			if gotClosedOut.Load().(bool) != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut.Load().(bool))
			}
		})
	}
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
		{false, false, true, false},
		{true, true, false, false},
		{true, false, true, false},
		{false, true, true, false},
		{true, true, true, true},
	}

	aSwitch := NewSwitch(true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(true)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(true)
	defer cSwitch.Shutdown()

	gate := NewANDGate(aSwitch, bSwitch, cSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan bool, 1)

	go func() {
		for {
			got.Store(<-ch)
		}
	}()

	gate.WireUp(ch)

	time.Sleep(time.Millisecond * 10)

	if got.Load().(bool) != true {
		t.Error("Wanted power on the gate but got none")
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip#%d: Setting A power to %t and B power to %t and C power to %t", i+1, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			time.Sleep(time.Millisecond * 20)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
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
		{false, false, false, false},
	}

	aSwitch := NewSwitch(true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(false)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(true)
	defer cSwitch.Shutdown()

	gate := NewORGate(aSwitch, bSwitch, cSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan bool, 1)

	go func() {
		for {
			got.Store(<-ch)
		}
	}()

	gate.WireUp(ch)

	time.Sleep(time.Millisecond * 10)

	if got.Load().(bool) != true {
		t.Error("Wanted power on the gate but got none")
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip#%d: Setting A power to %t and B power to %t and C power to %t", i+1, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			time.Sleep(time.Millisecond * 20)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
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
		{false, false, false, true},
	}

	aSwitch := NewSwitch(true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(false)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(true)
	defer cSwitch.Shutdown()

	gate := NewNANDGate(aSwitch, bSwitch, cSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan bool, 1)

	go func() {
		for {
			got.Store(<-ch)
		}
	}()

	gate.WireUp(ch)

	time.Sleep(time.Millisecond * 10)

	if got.Load().(bool) != true {
		t.Error("Wanted power on the gate but got none")
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip#%d: Setting A power to %t and B power to %t and C power to %t", i+1, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			time.Sleep(time.Millisecond * 20)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
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
		{false, false, false, true},
	}

	aSwitch := NewSwitch(true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(true)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(true)
	defer cSwitch.Shutdown()

	gate := NewNORGate(aSwitch, bSwitch, cSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan bool, 1)

	go func() {
		for {
			got.Store(<-ch)
		}
	}()

	gate.WireUp(ch)

	time.Sleep(time.Millisecond * 10)

	if got.Load().(bool) != false {
		t.Error("Wanted no power on the gate but got some")
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip#%d: Setting A power to %t and B power to %t and C power to %t", i+1, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			time.Sleep(time.Millisecond * 50)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
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
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(false)
	defer bSwitch.Shutdown()

	gate := NewXORGate(aSwitch, bSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan bool, 1)

	go func() {
		for {
			got.Store(<-ch)
		}
	}()

	gate.WireUp(ch)

	time.Sleep(time.Millisecond * 10)

	if got.Load().(bool) != false {
		t.Error("Wanted no power on the gate but got some")
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip#%d: Setting A power to %t and B power to %t", i+1, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			time.Sleep(time.Millisecond * 20)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
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
		{true, false, false},
		{false, true, false},
		{true, true, true},
		{false, false, true},
	}

	aSwitch := NewSwitch(false)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(false)
	defer bSwitch.Shutdown()

	gate := NewXNORGate(aSwitch, bSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan bool, 1)

	go func() {
		for {
			got.Store(<-ch)
		}
	}()

	gate.WireUp(ch)

	time.Sleep(time.Millisecond * 10)

	if got.Load().(bool) != true {
		t.Error("Wanted power on the gate but got none")
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Flip#%d: Setting A power to %t and B power to %t", i+1, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			time.Sleep(time.Millisecond * 20)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
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
		{false, true},
	}

	pin1Battery := NewBattery(true)
	inv := NewInverter(pin1Battery)
	defer inv.Shutdown()

	var got atomic.Value
	ch := make(chan bool, 1)

	go func() {
		for {
			got.Store(<-ch)
		}
	}()

	inv.WireUp(ch)

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Input as %t.", tc.inPowered), func(t *testing.T) {

			if tc.inPowered {
				pin1Battery.Charge()
			} else {
				pin1Battery.Discharge()
			}

			time.Sleep(time.Millisecond * 10)

			if got.Load().(bool) != tc.wantOut {
				t.Errorf("Input power was %t so wanted it inverted to %t but got %t", tc.inPowered, tc.wantOut, got.Load().(bool))
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
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(false)
	defer bSwitch.Shutdown()

	half := NewHalfAdder(aSwitch, bSwitch)
	defer half.Shutdown()

	var gotSum, gotCarry atomic.Value
	chSum := make(chan bool, 1)
	chCarry := make(chan bool, 1)

	go func() {
		for {
			select {
			case sum := <-chSum:
				gotSum.Store(sum)
			case carry := <-chCarry:
				gotCarry.Store(carry)
			}
		}
	}()

	half.Sum.WireUp(chSum)
	half.Carry.WireUp(chCarry)

	time.Sleep(time.Millisecond * 10)

	if gotSum.Load().(bool) != false {
		t.Error("Wanted no Sum but got one")
	}

	if gotCarry.Load().(bool) != false {
		t.Error("Wanted no Carry but got one")
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input source A to %t and source B to %t", tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			time.Sleep(time.Millisecond * 15)

			if gotSum.Load().(bool) != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, gotSum.Load().(bool))
			}

			if gotCarry.Load().(bool) != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, gotCarry.Load().(bool))
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
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(false)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(false)
	defer cSwitch.Shutdown()

	full := NewFullAdder(aSwitch, bSwitch, cSwitch)
	defer full.Shutdown()

	var gotSum, gotCarry atomic.Value
	chSum := make(chan bool, 1)
	chCarry := make(chan bool, 1)

	go func() {
		for {
			select {
			case sum := <-chSum:
				gotSum.Store(sum)
			case carry := <-chCarry:
				gotCarry.Store(carry)
			}
		}
	}()

	full.Sum.WireUp(chSum)
	full.Carry.WireUp(chCarry)

	time.Sleep(time.Millisecond * 10)

	if gotSum.Load().(bool) != false {
		t.Error("Wanted no Sum but got one")
	}

	if gotCarry.Load().(bool) != false {
		t.Error("Wanted no Carry but got one")
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input source A to %t and source B to %t with carry in of %t", tc.aInPowered, tc.bInPowered, tc.carryInPowered), func(t *testing.T) {

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.carryInPowered)

			time.Sleep(time.Millisecond * 15)

			if gotSum.Load().(bool) != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, gotSum.Load().(bool))
			}

			if gotCarry.Load().(bool) != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, gotCarry.Load().(bool))
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
			defer addend1Switches.Shutdown()

			addend2Switches, _ := NewNSwitchBank(tc.byte2)
			defer addend2Switches.Shutdown()

			addr, err := NewNBitAdder(addend1Switches.Switches, addend2Switches.Switches, nil)

			if addr != nil {
				t.Error("Did not expect an adder to be created, but got one")
				addr.Shutdown()
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

func TestNBitAdder_EightBit(t *testing.T) {
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

	// start with off switches
	addend1Switches, _ := NewNSwitchBank("00000000")
	defer addend1Switches.Shutdown()

	addend2Switches, _ := NewNSwitchBank("00000000")
	defer addend2Switches.Shutdown()

	carryInSwitch := NewSwitch(false)
	defer carryInSwitch.Shutdown()

	// create the adder based on those switches
	addr, err := NewNBitAdder(addend1Switches.Switches, addend2Switches.Switches, carryInSwitch)

	if err != nil {
		t.Errorf("Expected no error on construction, but got: %s", err.Error())
	}

	if addr == nil {
		t.Error("Expected an adder to return due to good inputs, but got a nil one.")
	}

	defer addr.Shutdown()

	// setup the Sum results bool array (default all to false to match the initial switch states above)
	var gotSums [8]atomic.Value
	for i := 0; i < len(gotSums); i++ {
		gotSums[i].Store(false)
	}

	// setup the channels for listening to channel changes (doing dynamic select-case vs. a stack of 8 channels)
	cases := make([]reflect.SelectCase, len(addr.Sums)+1) // one for each sum, BUT a +1 to hold the CarryOut channel read

	// setup a case for each sum
	for i, sum := range addr.Sums {
		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		sum.WireUp(ch)
	}

	// setup the single CarryOut result
	var gotCarryOut atomic.Value

	// add a case for the single CarryOut channel
	chCarryOut := make(chan bool, 1)
	cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(chCarryOut)}
	addr.CarryOut.WireUp(chCarryOut)

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosen, value, _ := reflect.Select(cases)

			// if know the selected case was within the range of Sums, set the matching Sums bool array element
			if chosen < len(cases)-1 {
				gotSums[chosen].Store(value.Bool())
			} else {
				gotCarryOut.Store(value.Bool())
			}
		}
	}()

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", tc.byte1, tc.byte2, tc.carryInPowered), func(t *testing.T) {

			setSwitches(addend1Switches, tc.byte1)
			setSwitches(addend2Switches, tc.byte2)
			carryInSwitch.Set(tc.carryInPowered)

			time.Sleep(time.Millisecond * 150)

			// build a string based on each sum's state
			gotAnswer := ""
			for i := 0; i < len(gotSums); i++ {
				if gotSums[i].Load().(bool) {
					gotAnswer += "1"
				} else {
					gotAnswer += "0"
				}
			}

			if gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, gotAnswer)
			}

			if gotCarryOut.Load().(bool) != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, gotCarryOut.Load().(bool))
			}
		})
	}
}

func TestNBitAdder_SixteenBit(t *testing.T) {
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

	// start with off switches
	addend1Switches, _ := NewNSwitchBank("0000000000000000")
	defer addend1Switches.Shutdown()

	addend2Switches, _ := NewNSwitchBank("0000000000000000")
	defer addend2Switches.Shutdown()

	carryInSwitch := NewSwitch(false)
	defer carryInSwitch.Shutdown()

	addr, err := NewNBitAdder(addend1Switches.Switches, addend2Switches.Switches, carryInSwitch)

	if err != nil {
		t.Errorf("Expected no error on construction, but got: %s", err.Error())
	}

	if addr == nil {
		t.Error("Expected an adder to return due to good inputs, but got a nil one.")
	}

	defer addr.Shutdown()

	// setup the Sum results bool array (default all to false to match the initial switch states above)
	var gotSums [16]atomic.Value
	for i := 0; i < len(gotSums); i++ {
		gotSums[i].Store(false)
	}

	// setup the channels for listening to channel changes (doing dynamic select-case vs. a stack of 8 channels)
	cases := make([]reflect.SelectCase, len(addr.Sums)+1) // one for each sum, BUT a +1 to hold the CarryOut channel read

	for i, sum := range addr.Sums {
		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		sum.WireUp(ch)
	}

	// setup the single CarryOut result
	var gotCarryOut atomic.Value

	// add a case for the single CarryOut channel
	chCarryOut := make(chan bool, 1)
	cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(chCarryOut)}
	addr.CarryOut.WireUp(chCarryOut)

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosen, value, _ := reflect.Select(cases)

			// if know the selected case was within the range of Sums, set the matching Sums bool array element
			if chosen < len(cases)-1 {
				gotSums[chosen].Store(value.Bool())
			} else {
				gotCarryOut.Store(value.Bool())
			}
		}
	}()

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %t", tc.bytes1, tc.bytes2, tc.carryInPowered), func(t *testing.T) {

			setSwitches(addend1Switches, tc.bytes1)
			setSwitches(addend2Switches, tc.bytes2)
			carryInSwitch.Set(tc.carryInPowered)

			time.Sleep(time.Millisecond * 400)

			// build a string based on each sum's state
			gotAnswer := ""
			for i := 0; i < len(gotSums); i++ {
				if gotSums[i].Load().(bool) {
					gotAnswer += "1"
				} else {
					gotAnswer += "0"
				}
			}

			if gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, gotAnswer)
			}

			if gotCarryOut.Load().(bool) != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, gotCarryOut.Load().(bool))
			}
		})
	}
}

/*
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
*/

func TestOnesCompliment(t *testing.T) {

	testCases := []struct {
		bits            string
		signalIsPowered bool
		wantCompliment  string
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

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Executing complementer against %s with compliment signal of %t", tc.bits, tc.signalIsPowered), func(t *testing.T) {
			bitSwitches, _ := NewNSwitchBank(tc.bits)
			defer bitSwitches.Shutdown()

			comp := NewOnesComplementer(bitSwitches.Switches, NewSwitch(tc.signalIsPowered))

			if comp == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
			}

			defer comp.Shutdown()

			// setup the Compliments results bool array (default all to false)
			gotCompliments := make([]atomic.Value, len(tc.bits))
			for i := 0; i < len(tc.bits); i++ {
				gotCompliments[i].Store(false)
			}

			// setup the channels for listening to Compliments change
			cases := make([]reflect.SelectCase, len(tc.bits))
			for i, cmp := range comp.Complements {
				ch := make(chan bool, 1)
				cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
				cmp.WireUp(ch)
			}

			go func() {
				for {
					// run the dynamic select statement to see which case index hit and the value we got off the associated channel
					chosen, value, _ := reflect.Select(cases)
					gotCompliments[chosen].Store(value.Bool())
				}
			}()

			time.Sleep(time.Millisecond * 15)

			// build a string based on each bit's state
			gotCompliment := ""
			for i := 0; i < len(gotCompliments); i++ {
				if gotCompliments[i].Load().(bool) {
					gotCompliment += "1"
				} else {
					gotCompliment += "0"
				}
			}

			if gotCompliment != tc.wantCompliment {
				t.Errorf("Wanted %s, but got %s", tc.wantCompliment, gotCompliment)
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
			minuendSwitches, _ := NewNSwitchBank(tc.byte1)
			defer minuendSwitches.Shutdown()

			subtrahendSwitches, _ := NewNSwitchBank(tc.byte2)
			defer subtrahendSwitches.Shutdown()

			sub, err := NewNBitSubtractor(minuendSwitches.Switches, subtrahendSwitches.Switches)

			if sub != nil {
				t.Error("Did not expect a Subtractor to be created, but got one")
				sub.Shutdown()
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

func TestNBitSubtractor_EightBit(t *testing.T) {
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

	// start with off switches
	minuendwitches, _ := NewNSwitchBank("00000000")
	defer minuendwitches.Shutdown()

	subtrahendSwitches, _ := NewNSwitchBank("00000000")
	defer subtrahendSwitches.Shutdown()

	sub, _ := NewNBitSubtractor(minuendwitches.Switches, subtrahendSwitches.Switches)

	if sub == nil {
		t.Error("Expected an subtractor to return due to good inputs, but gotAnswer c nil one.")
	}

	defer sub.Shutdown()

	// setup the Differences results bool array (default all to false to match the initial switch states above)
	var gotDifferences [8]atomic.Value
	for i := 0; i < len(gotDifferences); i++ {
		gotDifferences[i].Store(false)
	}

	// setup the channels for listening to Differences change
	cases := make([]reflect.SelectCase, len(sub.Differences)+1) // +1 to hold the CarryOut channel read
	for i, diff := range sub.Differences {
		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		diff.WireUp(ch)
	}

	// setup the single CarryOut result
	var gotCarryOut atomic.Value

	// add a case for the single CarryOut channel
	chCarryOut := make(chan bool, 1)
	cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(chCarryOut)}
	sub.CarryOut.WireUp(chCarryOut)

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosen, value, _ := reflect.Select(cases)

			// if know the selected case was within the range of Differences, set the matching Differences bool array element
			if chosen < len(cases)-1 {
				gotDifferences[chosen].Store(value.Bool())
			} else {
				gotCarryOut.Store(value.Bool())
			}
		}
	}()

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Subtracting %s from %s", tc.subtrahend, tc.minuend), func(t *testing.T) {

			setSwitches(minuendwitches, tc.minuend)
			setSwitches(subtrahendSwitches, tc.subtrahend)

			time.Sleep(time.Millisecond * 300)

			// build a string based on each bit's state
			gotAnswer := ""
			for i := 0; i < len(gotDifferences); i++ {
				if gotDifferences[i].Load().(bool) {
					gotAnswer += "1"
				} else {
					gotAnswer += "0"
				}
			}

			if gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, gotAnswer)
			}

			if gotCarryOut.Load().(bool) != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, gotCarryOut.Load().(bool))
			}
		})
	}
}

func TestOscillator(t *testing.T) {
	testCases := []struct {
		initState   bool
		oscHertz    int
		wantResults string
	}{
		{false, 1, "010"},
		{true, 1, "101"},
		{false, 5, "01010101010"},
		{true, 5, "10101010101"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Oscillating at %d hertz, immediate start (%t)", tc.oscHertz, tc.initState), func(t *testing.T) {

			osc := NewOscillator(tc.initState)
			defer osc.Stop()

			var gotResults atomic.Value
			ch := make(chan bool, 1)

			gotResults.Store("")
			go func() {
				for {
					result := gotResults.Load().(string)
					if <-ch {
						result += "1"
					} else {
						result += "0"
					}
					gotResults.Store(result)
				}
			}()

			osc.WireUp(ch)
			osc.Oscillate(tc.oscHertz)

			time.Sleep(time.Second * 3)

			if !strings.HasPrefix(gotResults.Load().(string), tc.wantResults) {
				t.Errorf("Wanted results of at least %s but got %s.", tc.wantResults, gotResults.Load().(string))
			}
		})
	}
}

func TestRSFlipFlop(t *testing.T) {
	testCases := []struct {
		rPinIsPowered bool
		sPinIsPowered bool
		wantQ         bool
		wantQBar      bool
	}{ // construction of the flipflop will start with a default of rPin:false, sPin:false, which causes false on both inputs of the S nor, which causes QBar on (Q off)
		{false, false, false, true}, // Un-Set should change nother
		{false, true, true, false},  // Set causes Q on (QBar off)
		{false, true, true, false},  // Set again should change nothing
		{false, false, true, false}, // Un-Set should remember prior
		{false, false, true, false}, // Un-Set again should change nothing
		{true, false, false, true},  // Reset causes Q off (QBar on)
		{true, false, false, true},  // Reset again should change nothing
		{false, false, false, true}, // Un-Reset should remember prior
		{true, false, false, true},  // Un-Reset again should change nothing
		{false, true, true, false},  // Set causes Q on (QBar off)
		{true, false, false, true},  // Reset causes Q off (QBar on)
		{false, true, true, false},  // Set causes Q on (QBar off)
		{false, false, true, false}, // Un-Set again should change nothing
	}

	testName := func(i int) string {
		var priorR bool
		var priorS bool

		if i == 0 {
			priorR = false
			priorS = false
		} else {
			priorR = testCases[i-1].rPinIsPowered
			priorS = testCases[i-1].sPinIsPowered
		}

		return fmt.Sprintf("Stage#%d: Switching from [rInIsPowered (%t) sInIsPowered (%t)] to [rInIsPowered (%t) sInIsPowered (%t)]", i+1, priorR, priorS, testCases[i].rPinIsPowered, testCases[i].sPinIsPowered)
	}

	var rPinBattery, sPinBattery *Battery
	rPinBattery = NewBattery(false)
	sPinBattery = NewBattery(false)

	chQ := make(chan bool, 1)
	chQBar := make(chan bool, 1)

	// starting with no input signals (R and S are off)
	ff := NewRSFlipFLop(rPinBattery, sPinBattery)
	defer ff.Shutdown()

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case newQ := <-chQ:
				gotQ.Store(newQ)
			case newQBar := <-chQBar:
				gotQBar.Store(newQBar)
			}
		}
	}()

	ff.QBar.WireUp(chQBar)
	ff.Q.WireUp(chQ)

	time.Sleep(time.Millisecond * 125)

	if gotQ.Load().(bool) != false {
		t.Errorf("Wanted power of %t at Q, but got %t.", false, gotQ.Load().(bool))
	}

	if gotQBar.Load().(bool) != true {
		t.Errorf("Wanted power of %t at QBar, but got %t.", true, gotQBar.Load().(bool))
	}

	for i, tc := range testCases {
		t.Run(testName(i), func(t *testing.T) {

			if tc.rPinIsPowered {
				rPinBattery.Charge()
			} else {
				rPinBattery.Discharge()
			}

			if tc.sPinIsPowered {
				sPinBattery.Charge()
			} else {
				sPinBattery.Discharge()
			}

			time.Sleep(time.Millisecond * 125)

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
}

func TestLevelTriggeredDTypeLatch(t *testing.T) {
	testCases := []struct {
		clkIn    bool
		dataIn   bool
		wantQ    bool
		wantQBar bool
	}{ // construction of the latches will start with a default of clkIn:true, dataIn:true, which causes Q on (QBar off)
		{false, false, true, false}, // clkIn off should cause no change regardless of dataIn
		{false, true, true, false},  // clkIn off should cause no change regardless of dataIn
		{true, true, true, false},   // clkIn with dataIn causes no change since same Q state as prior
		{false, false, true, false}, // clkIn off should cause no change
		{true, false, false, true},  // clkIn with no dataIn causes Q off (QBar on)
		{false, false, false, true}, // clkIn off should cause no change
		{true, false, false, true},  // clkIn again with same dataIn should cause no change
		{true, true, true, false},   // clkIn with dataIn should cause Q on (QBar off)
		{false, false, true, false}, // clkIn off should cause no change
		{true, true, true, false},   // clkIn off should cause no change since same Q state as prior
		{true, false, false, true},  // clkIn on with no dataIn causes Q off (QBar on)
		{true, true, true, false},   // clkIn on with dataIn causes Q on (QBar off)
		{true, false, false, true},  // clkIn on with no dataIn causes Q off (QBar on)
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

		return fmt.Sprintf("Stage#%d: Switching from [clkIn (%t) dataIn (%t)] to [clkIn (%t) dataIn (%t)]", i+1, priorClkIn, priorDataIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	var clkBattery, dataBattery *Battery
	clkBattery = NewBattery(true)
	dataBattery = NewBattery(true)

	chQ := make(chan bool, 1)
	chQBar := make(chan bool, 1)

	// starting with true input signals (Clk and Data are on)
	latch := NewLevelTriggeredDTypeLatch(clkBattery, dataBattery)
	defer latch.Shutdown()

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case newQ := <-chQ:
				gotQ.Store(newQ)
			case newQBar := <-chQBar:
				gotQBar.Store(newQBar)
			}
		}
	}()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	time.Sleep(time.Millisecond * 75)

	if gotQ.Load().(bool) != true {
		t.Errorf("Wanted power of %t at Q, but got %t.", true, gotQ.Load().(bool))
	}

	if gotQBar.Load().(bool) != false {
		t.Errorf("Wanted power of %t at QBar, but got %t.", false, gotQBar.Load().(bool))
	}

	for i, tc := range testCases {
		t.Run(testName(i), func(t *testing.T) {

			if tc.clkIn {
				clkBattery.Charge()
			} else {
				clkBattery.Discharge()
			}

			time.Sleep(time.Millisecond * 200)

			if tc.dataIn {
				dataBattery.Charge()
			} else {
				dataBattery.Discharge()
			}

			time.Sleep(time.Millisecond * 200)

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
}

func TestNBitLevelTriggeredDTypeLatch(t *testing.T) {
	testCases := []struct {
		input string
		want  [8]bool
	}{
		{"00000001", [8]bool{false, false, false, false, false, false, false, true}},
		{"11111111", [8]bool{true, true, true, true, true, true, true, true}},
		{"10101010", [8]bool{true, false, true, false, true, false, true, false}},
		{"10000001", [8]bool{true, false, false, false, false, false, false, true}},
	}

	latchSwitches, _ := NewNSwitchBank("00011000")
	defer latchSwitches.Shutdown()

	clkSwitch := NewSwitch(true)
	defer clkSwitch.Shutdown()

	latch := NewNBitLevelTriggeredDTypeLatch(clkSwitch, latchSwitches.Switches)
	defer latch.Shutdown()

	// for use in a dynamic select statement (a case per Q of the latch array) and bool results per case
	cases := make([]reflect.SelectCase, 8)
	got := make([]atomic.Value, 8)

	// built the case statements to deal with each Q in the latch array
	for i, q := range latch.Qs {

		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}

		q.WireUp(ch)
	}

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosenCase, caseValue, _ := reflect.Select(cases)
			got[chosenCase].Store(caseValue.Bool())
		}
	}()

	// let the above settle down before testing
	time.Sleep(time.Millisecond * 100)

	priorWant := [8]bool{false, false, false, true, true, false, false, false}
	for i := 0; i < 8; i++ {
		if got := got[i].Load().(bool); got != priorWant[i] {
			t.Errorf("Latch[%d] wanted (%t) but got (%t).\n", i, priorWant[i], got)
		}
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Stage#%d: Setting switches to %s", i+1, tc.input), func(t *testing.T) {

			// set to OFF to test that nothing will change in the latches store

			clkSwitch.Set(false)
			time.Sleep(time.Millisecond * 125)

			setSwitches(latchSwitches, tc.input) // setting switches AFTER the clk goes to off to test that nothing actually would happen to the latches

			for i := range latch.Qs {
				if got := got[i].Load().(bool); got != priorWant[i] {
					t.Errorf("Latch[%d], with clkSwitch off, wanted %t but got %t", i, priorWant[i], got)
				}
			}

			// Now set to ON to test that requested changes DID occur in the latches store

			clkSwitch.Set(true)
			time.Sleep(time.Millisecond * 200) // need to allow all the latches to settle down (transmit their new Q values)

			for i := range latch.Qs {
				if got := got[i].Load().(bool); got != tc.want[i] {
					t.Errorf("Latch[%d], with clkSwitch ON, wanted %t but got %t", i, tc.want[i], got)
				}
			}

			// now update the prior tracker bools to ensure next pass (with cklIn as OFF at the top) proves it didn't change (ie matches prior)
			for i := range latch.Qs {
				priorWant[i] = got[i].Load().(bool)
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
			defer addend1Switches.Shutdown()

			addend2Switches, _ := NewNSwitchBank(tc.byte2)
			defer addend2Switches.Shutdown()

			sel, err := NewTwoToOneSelector(nil, addend1Switches.Switches, addend2Switches.Switches)

			if sel != nil {
				t.Error("Did not expect a Selector to be created, but got one")
				sel.Shutdown()
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
		{"110", "000", false, []bool{true, true, false}},
		{"110", "000", true, []bool{false, false, false}},
		{"110", "111", true, []bool{true, true, true}},
		{"110", "111", false, []bool{true, true, false}},
	}

	// start with these switches to verify uses A intially
	aInSwitches, _ := NewNSwitchBank("111")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank("000")
	defer bInSwitches.Shutdown()

	selectBSwitch := NewSwitch(false)
	defer selectBSwitch.Shutdown()

	// for use in a dynamic select statement (a case per selector output) and bool results per case
	cases := make([]reflect.SelectCase, 3)
	got := make([]atomic.Value, 3)

	sel, _ := NewTwoToOneSelector(selectBSwitch, aInSwitches.Switches, bInSwitches.Switches)
	defer sel.Shutdown()

	// built the case statements to deal with each selector output
	for i, s := range sel.Outs {

		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}

		s.WireUp(ch)
	}

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosenCase, caseValue, _ := reflect.Select(cases)
			got[chosenCase].Store(caseValue.Bool())
		}
	}()

	// let the above settle down before testing
	time.Sleep(time.Millisecond * 75)

	want := true
	for i := 0; i < 3; i++ {
		if got := got[i].Load().(bool); got != want {
			t.Errorf("Selector Output[%d]: A(111), B(000), use B?(false).  Wanted (%t) but got (%v).\n", i, want, got)
		}
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Stage[%d]: A(%s), B(%s), use B?(%t)", i, tc.aIn, tc.bIn, tc.selectB), func(t *testing.T) {

			setSwitches(aInSwitches, tc.aIn)
			setSwitches(bInSwitches, tc.bIn)
			selectBSwitch.Set(tc.selectB)

			time.Sleep(time.Millisecond * 75)

			for i := range sel.Outs {
				if got := got[i].Load().(bool); got != tc.want[i] {
					t.Errorf("Selector Output[%d]: Wanted (%t) but got (%t).\n", i, tc.want[i], got)
				}
			}
		})
	}
}

func TestTwoToOneSelector_SelectingB_ASwitchesNoImpact(t *testing.T) {
	// start with off for A but on for B, but selecting A
	aInSwitches, _ := NewNSwitchBank("000")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank("111")
	defer bInSwitches.Shutdown()

	selectBSwitch := NewSwitch(false)
	defer selectBSwitch.Shutdown()

	// for use in a dynamic select statement (a case per selector output) and bool results per case
	cases := make([]reflect.SelectCase, 3)
	got := make([]atomic.Value, 3)

	sel, _ := NewTwoToOneSelector(selectBSwitch, aInSwitches.Switches, bInSwitches.Switches)
	defer sel.Shutdown()

	// built the case statements to deal with each selector output
	for i, s := range sel.Outs {

		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}

		s.WireUp(ch)
	}

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosenCase, caseValue, _ := reflect.Select(cases)
			got[chosenCase].Store(caseValue.Bool())
		}
	}()

	time.Sleep(time.Millisecond * 75)

	// starting with selecting A, get A's state

	for i := 0; i < 3; i++ {
		if got[i].Load().(bool) == true {
			t.Error("Expecting false on all Outs of selector but got a true")
		}
	}

	selectBSwitch.Set(true)
	time.Sleep(time.Millisecond * 75)

	// selecting B, get B's state
	for i := 0; i < 3; i++ {
		if got[i].Load().(bool) == false {
			t.Error("Expecting true on all Outs of selector but got a false")
		}
	}

	setSwitches(aInSwitches, "101")
	time.Sleep(time.Millisecond * 75)

	// still selecting B, get B's state, regardless of A's state changing
	for i := 0; i < 3; i++ {
		if got[i].Load().(bool) == false {
			t.Error("Expecting true on all Outs of selector but got a false")
		}
	}
}

func TestThreeNumberAdder_MismatchInputs(t *testing.T) {
	wantError := "Mismatched input lengths. Addend1 len: 8, Addend2 len: 4"

	aInSwitches, _ := NewNSwitchBank("00000000")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank("0000")
	defer bInSwitches.Shutdown()

	addr, err := NewThreeNumberAdder(aInSwitches.Switches, bInSwitches.Switches)

	if addr != nil {
		t.Error("Did not expect an adder back but got one.")
		addr.Shutdown()
	}

	if err != nil && err.Error() != wantError {
		t.Errorf("Wanted error %s, but got %v", wantError, err.Error())
	}
}

func TestThreeNumberAdder_TwoNumberAdd(t *testing.T) {
	testCases := []struct {
		aIn          string
		bIn          string
		wantAnswer   string
		wantCarryOut bool
	}{
		{"00000000", "00000001", "00000001", false},
		{"00000001", "00000010", "00000011", false},
		{"10000001", "10000000", "00000001", true},
		{"11111111", "11111111", "11111110", true},
	}

	aInSwitches, _ := NewNSwitchBank("00000000")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank("00000000")
	defer bInSwitches.Shutdown()

	addr, _ := NewThreeNumberAdder(aInSwitches.Switches, bInSwitches.Switches)
	defer addr.Shutdown()

	// setup the Sum results bool array (default all to false to match the initial switch states above)
	var gotSums [8]atomic.Value
	for i := 0; i < len(gotSums); i++ {
		gotSums[i].Store(false)
	}

	// setup the channels for listening to channel changes (doing dynamic select-case vs. a stack of 8 channels)
	cases := make([]reflect.SelectCase, len(addr.Sums)+1) // one for each sum, BUT a +1 to hold the CarryOut channel read

	// setup a case for each sum
	for i, sum := range addr.Sums {
		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		sum.WireUp(ch)
	}

	// setup the single CarryOut result
	var gotCarryOut atomic.Value

	// add a case for the single CarryOut channel
	chCarryOut := make(chan bool, 1)
	cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(chCarryOut)}
	addr.CarryOut.WireUp(chCarryOut)

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosen, value, _ := reflect.Select(cases)

			// if know the selected case was within the range of Sums, set the matching Sums bool array element
			if chosen < len(cases)-1 {
				gotSums[chosen].Store(value.Bool())
			} else {
				gotCarryOut.Store(value.Bool())
			}
		}
	}()

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.aIn, tc.bIn), func(t *testing.T) {

			setSwitches(aInSwitches, tc.aIn)
			setSwitches(bInSwitches, tc.bIn)

			time.Sleep(time.Millisecond * 350)

			// build a string based on each sum's state
			gotAnswer := ""
			for i := 0; i < len(gotSums); i++ {
				if gotSums[i].Load().(bool) {
					gotAnswer += "1"
				} else {
					gotAnswer += "0"
				}
			}

			if gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s but got %s", tc.wantAnswer, gotAnswer)
			}

			if gotCarryOut.Load().(bool) != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, gotCarryOut.Load().(bool))
			}
		})
	}
}

func TestThreeNumberAdder_ThreeNumberAdd(t *testing.T) {

	aInSwitches, _ := NewNSwitchBank("00000010")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank("00000001")
	defer bInSwitches.Shutdown()

	addr, _ := NewThreeNumberAdder(aInSwitches.Switches, bInSwitches.Switches)
	defer addr.Shutdown()

	// setup the Sum results bool array (default all to false to match the initial switch states above)
	var gotSums [8]atomic.Value
	for i := 0; i < len(gotSums); i++ {
		gotSums[i].Store(false)
	}

	// setup the channels for listening to channel changes (doing dynamic select-case vs. a stack of 8 channels)
	cases := make([]reflect.SelectCase, len(addr.Sums)+1) // one for each sum, BUT a +1 to hold the CarryOut channel read

	// setup a case for each sum
	for i, sum := range addr.Sums {
		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		sum.WireUp(ch)
	}

	// setup the single CarryOut result
	var gotCarryOut atomic.Value

	// add a case for the single CarryOut channel
	chCarryOut := make(chan bool, 1)
	cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(chCarryOut)}
	addr.CarryOut.WireUp(chCarryOut)

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosen, value, _ := reflect.Select(cases)

			// if know the selected case was within the range of Sums, set the matching Sums bool array element
			if chosen < len(cases)-1 {
				gotSums[chosen].Store(value.Bool())
			} else {
				gotCarryOut.Store(value.Bool())
			}
		}
	}()

	// lots to settle above before validating results
	time.Sleep(time.Millisecond * 250)

	wantAnswer := "00000011" // 2 + 1 = 3
	wantCarry := false

	// build a string based on each sum's state
	gotAnswer := ""
	for i := 0; i < len(gotSums); i++ {
		if gotSums[i].Load().(bool) {
			gotAnswer += "1"
		} else {
			gotAnswer += "0"
		}
	}

	if gotAnswer != wantAnswer {
		t.Errorf("Wanted answer %s but got %s", wantAnswer, gotAnswer)
	}

	if gotCarryOut.Load().(bool) != wantCarry {
		t.Errorf("Wanted carry %t, but got %t", wantCarry, gotCarryOut.Load().(bool))
	}

	addr.SaveToLatch.Set(true)
	time.Sleep(time.Millisecond * 100)
	addr.SaveToLatch.Set(false)

	setSwitches(aInSwitches, "00000011")

	addr.ReadFromLatch.Set(true)
	time.Sleep(time.Millisecond * 250)

	wantAnswer = "00000110" // 3 + 3 = 6
	wantCarry = false

	// build a string based on each sum's state
	gotAnswer = ""
	for i := 0; i < len(gotSums); i++ {
		if gotSums[i].Load().(bool) {
			gotAnswer += "1"
		} else {
			gotAnswer += "0"
		}
	}

	if gotAnswer != wantAnswer {
		t.Errorf("Wanted answer %s but got %s", wantAnswer, gotAnswer)
	}

	if gotCarryOut.Load().(bool) != wantCarry {
		t.Errorf("Wanted carry %t, but got %t", wantCarry, gotCarryOut.Load().(bool))
	}

	// can't add a fourth number, since once Selector is on the B side, the data from the Latch (once SaveToLatch goes true again)
	//     will send data at an unknown rate down to the adder, even if set SaveToLatch false.  Hard to get the timing just right
	//     unless I can redo all the timings in the whole system to know exactly when a component has settled (vs. relying on pauses)
	//
	//     OR....would an extra "barrier" latch between the current latch and the 2-1-Selector allow more control over the stages of the loopback? Hmmmm
}

func TestLevelTriggeredDTypeLatchWithClear(t *testing.T) {
	testCases := []struct {
		clrIn    bool
		clkIn    bool
		dataIn   bool
		wantQ    bool
		wantQBar bool
	}{ // construction of the latches will start with a default of clkIn:true, dataIn:true, which causes Q on (QBar off)
		{false, false, true, true, false},  // clrIn off, clkIn off, do nothing
		{false, true, true, true, false},   // clrIn off, clkIn on, Q = data
		{false, true, false, false, true},  // clrIn off, clkIn on, Q = data
		{false, false, true, false, true},  // clrIn off, clkIn off, do nothing
		{false, false, false, false, true}, // clrIn off, clkIn off, do nothing
		{false, true, true, true, false},   // clrIn off, clkIn on, Q = data
		{true, false, true, false, true},   // clrIn ON, Q always false, QBar will depend on if clock and data are both true
		{true, true, true, false, false},   // clrIn ON, Q always false, QBar will depend on if clock and data are both true
		{true, true, false, false, true},   // clrIn ON, Q always false, QBar will depend on if clock and data are both true
		{true, false, true, false, true},   // clrIn ON, Q always false, QBar will depend on if clock and data are both true
		{true, false, false, false, true},  // clrIn ON, Q always false, QBar will depend on if clock and data are both true
		{true, true, true, false, false},   // clrIn ON, Q always false, QBar will depend on if clock and data are both true
	}

	testName := func(i int) string {
		var priorClrIn bool
		var priorClkIn bool
		var priorDataIn bool

		if i == 0 {
			priorClrIn = false
			priorClkIn = true
			priorDataIn = true
		} else {
			priorClrIn = testCases[i-1].clrIn
			priorClkIn = testCases[i-1].clkIn
			priorDataIn = testCases[i-1].dataIn
		}

		return fmt.Sprintf("Stage#%d: Switching from [clrIn (%t), clkIn (%t), dataIn (%t)] to [clrIn (%t), clkIn (%t), dataIn (%t)]", i+1, priorClrIn, priorClkIn, priorDataIn, testCases[i].clrIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	var clrBattery, clkBattery, dataBattery *Battery
	clrBattery = NewBattery(false)
	clkBattery = NewBattery(true)
	dataBattery = NewBattery(true)

	chQ := make(chan bool, 1)
	chQBar := make(chan bool, 1)

	latch := NewLevelTriggeredDTypeLatchWithClear(clrBattery, clkBattery, dataBattery)
	defer latch.Shutdown()

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case newQ := <-chQ:
				gotQ.Store(newQ)
			case newQBar := <-chQBar:
				gotQBar.Store(newQBar)
			}
		}
	}()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	time.Sleep(time.Millisecond * 200)

	if gotQ.Load().(bool) != true {
		t.Errorf("xWanted power of %t at Q, but got %t.", true, gotQ.Load().(bool))
	}

	if gotQBar.Load().(bool) != false {
		t.Errorf("xWanted power of %t at QBar, but got %t.", false, gotQBar.Load().(bool))
	}

	for i, tc := range testCases {
		t.Run(testName(i), func(t *testing.T) {

			if tc.clrIn {
				clrBattery.Charge()
			} else {
				clrBattery.Discharge()
			}

			time.Sleep(time.Millisecond * 200)

			if tc.clkIn {
				clkBattery.Charge()
			} else {
				clkBattery.Discharge()
			}

			time.Sleep(time.Millisecond * 200)

			if tc.dataIn {
				dataBattery.Charge()
			} else {
				dataBattery.Discharge()
			}

			time.Sleep(time.Millisecond * 200)

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
}

func TestNBitLevelTriggeredDTypeLatchWithClear(t *testing.T) {
	testCases := []struct {
		input string
		want  [8]bool
	}{
		{"00000001", [8]bool{false, false, false, false, false, false, false, true}},
		{"11111111", [8]bool{true, true, true, true, true, true, true, true}},
		{"10101010", [8]bool{true, false, true, false, true, false, true, false}},
		{"10000001", [8]bool{true, false, false, false, false, false, false, true}},
	}

	dataSwitches, _ := NewNSwitchBank("00000000")
	defer dataSwitches.Shutdown()

	clrSwitch := NewSwitch(false)
	defer clrSwitch.Shutdown()

	clkSwitch := NewSwitch(false)
	defer clkSwitch.Shutdown()

	latch := NewNBitLevelTriggeredDTypeLatchWithClear(clrSwitch, clkSwitch, dataSwitches.Switches)
	defer latch.Shutdown()

	// for use in a dynamic select statement (a case per Q of the latch array) and bool results per case
	cases := make([]reflect.SelectCase, 8)
	got := make([]atomic.Value, 8)

	// built the case statements to deal with each Q in the latch array
	for i, q := range latch.Qs {

		ch := make(chan bool, 1)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}

		q.WireUp(ch)
	}

	go func() {
		for {
			// run the dynamic select statement to see which case index hit and the value we got off the associated channel
			chosenCase, caseValue, _ := reflect.Select(cases)
			got[chosenCase].Store(caseValue.Bool())
		}
	}()

	// let the above settle down before testing
	time.Sleep(time.Millisecond * 250)

	priorWant := [8]bool{false, false, false, false, false, false, false, false}
	for i := 0; i < 8; i++ {
		if got := got[i].Load().(bool); got != priorWant[i] {
			t.Errorf("Latch[%d] wanted (%t) but got (%t).\n", i, priorWant[i], got)
		}
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Stage#%d: Setting switches to %s", i+1, tc.input), func(t *testing.T) {

			// reset the Clear from the prior run (asserts at the bottom of this loop)
			clrSwitch.Set(false)

			// set to OFF to test that nothing will change in the latches store
			clkSwitch.Set(false)
			setSwitches(dataSwitches, tc.input)
			time.Sleep(time.Millisecond * 100)

			for i := 0; i < 8; i++ {
				if got := got[i].Load().(bool); got != priorWant[i] {
					t.Errorf("Latch[%d] wanted (%t) but got (%t).\n", i, priorWant[i], got)
				}
			}

			// Now set to ON to test that requested changes did occur in the latches store
			clkSwitch.Set(true)
			time.Sleep(time.Millisecond * 250)

			for i := range latch.Qs {
				if got := got[i].Load().(bool); got != tc.want[i] {
					t.Errorf("Latch[%d], with clkSwitch ON, wanted %t but got %t", i, tc.want[i], got)
				}
			}

			// now update the prior tracker bools to ensure next pass (with cklIn as OFF at the top of the loop) proves it didn't change (aka matches prior)
			for i := range latch.Qs {
				priorWant[i] = got[i].Load().(bool)
			}

			// Now Clear the latches
			clrSwitch.Set(true)
			time.Sleep(time.Millisecond * 250) // need to allow Clear some time to force all Qs off

			// clear should have set all Qs to off
			want := false
			for i := range latch.Qs {
				if got := got[i].Load().(bool); got != want {
					t.Errorf("Latch[%d], with clrSwitch ON, wanted %t but got %t", i, want, got)
				}
			}
		})
	}
}

/*
// TestNNumberAdder creates an adder loop that has no bounds so it is expected to stack overlow
//     runtime: goroutine stack exceeds 1000000000-byte limit
//     fatal error: stack overflow
func TestNNumberAdder(t *testing.T) {

	switches, _ := NewNSwitchBank("00000001")
	addr, _ := NewNNumberAdder(switches)

	addr.Clear.Set(true)
	addr.Clear.Set(false)

	want := "00000000"
	if got := addr.AsAnswerString(); got != want {
		t.Errorf("[Initial setup] Wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
	}

	want = "00000001"
	if got := addr.adder.AsAnswerString(); got != want {
		t.Errorf("[Initial setup] Wanted answer of NNumberAdder's inner-adder to be %s but got %s", want, got)
	}

	addr.Add.Set(true)
	addr.Add.Set(false)

	want = "00000001"
	if got := addr.AsAnswerString(); got != want {
		t.Errorf("After an add, wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
	}

	updateSwitches(switches, "00000010")

	want = "00000011"
	if got := addr.AsAnswerString(); got != want {
		t.Errorf("After another add, wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
	}
}

/*
func TestEdgeTriggeredDTypeLatch(t *testing.T) {
	testCases := []struct {
		clkIn    bool
		dataIn   bool
		wantQ    bool
		wantQBar bool
	}{ // construction of the latches will start with a default of clkIn:false, dataIn:false, which causes Q off (QBar on)
		{false, true, false, true},  // clkIn staying false should cause no change
		{false, false, false, true}, // clkIn staying false should cause no change
		{false, true, false, true},  // clkIn staying false should cause no change, regardless of data change
		{true, true, true, false},   // clkIn going to true, with dataIn, causes Q on (QBar off)
		{true, false, true, false},  // clkIn staying true should cause no change, regardless of data change
		{false, false, true, false}, // clkIn going to false should cause no change
		{false, true, true, false},  // clkIn staying false should cause no change, regardless of data change
		{true, false, false, true},  // clkIn going to true, with no dataIn, causes Q off (QBar on)
		{true, true, false, true},   // clkIn staying true should cause no change, regardless of data change
	}

	testName := func(i int) string {
		var priorClkIn bool
		var priorDataIn bool

		if i == 0 {
			// trues since starting with charged batteries when Newing thew Latch initially
			priorClkIn = false
			priorDataIn = false
		} else {
			priorClkIn = testCases[i-1].clkIn
			priorDataIn = testCases[i-1].dataIn
		}

		return fmt.Sprintf("Stage#%d: Switching from [clkIn (%t) dataIn (%t)] to [clkIn (%t) dataIn (%t)]", i+1, priorClkIn, priorDataIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	var clkBattery, dataBattery *Battery
	clkBattery = NewBattery(false)
	dataBattery = NewBattery(false)

	latch := NewEdgeTriggeredDTypeLatch(clkBattery, dataBattery)

	want := false
	if gotQ := latch.Q.GetIsPowered(); gotQ != want {
		t.Errorf("On contruction, wanted power of %t at Q, but got %t.", want, gotQ)
	}

	want = true
	if gotQBar := latch.QBar.GetIsPowered(); gotQBar != want {
		t.Errorf("On construction, wanted power of %t at QBar, but got %t.", want, gotQBar)
	}

	for i, tc := range testCases {
		t.Run(testName(i), func(t *testing.T) {

			if tc.dataIn {
				dataBattery.Charge()
			} else {
				dataBattery.Discharge()
			}
			if tc.clkIn {
				clkBattery.Charge()
			} else {
				clkBattery.Discharge()
			}

			if gotQ := latch.Q.GetIsPowered(); gotQ != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ)
			}

			if gotQBar := latch.QBar.GetIsPowered(); gotQBar != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar)
			}
		})
	}
}

func TestFrequencyDivider(t *testing.T) {
	var gotQBarResults string

	osc := NewOscillator(false)
	freqDiv := NewFrequencyDivider(osc)

	freqDiv.QBar.WireUp(func(state bool) {
		if state {
			gotQBarResults += "1"
		} else {
			gotQBarResults += "0"
		}
	})

	osc.Oscillate(5)

	time.Sleep(time.Second * 3)

	osc.Stop()

	want := "10101010"
	if !strings.HasPrefix(gotQBarResults, want) {
		t.Errorf("Wanted results %s but got %s.", want, gotQBarResults)
	}
}

func TestFrequencyDivider2(t *testing.T) {

	sw1 := NewSwitch(false)
	freqDiv0 := NewFrequencyDivider(sw1)
	freqDiv1 := NewFrequencyDivider(freqDiv0.QBar)

	var result0 = ""
	var result1 = ""

	if freqDiv1.Q.GetIsPowered() {
		result1 = "1"
	} else {
		result1 = "0"
	}
	if freqDiv0.Q.GetIsPowered() {
		result0 = "1"
	} else {
		result0 = "0"
	}

	fmt.Println("==Just Created==")
	fmt.Println(freqDiv0.latch.StateDump("freqDiv0"))
	fmt.Println(freqDiv1.latch.StateDump("freqDiv1"))
	fmt.Println("\n==Ripple Value==\n" + result1 + result0 + "\n")

	fmt.Println("==Wiring Up freqDiv0.Q==")
	freqDiv0.Q.WireUp(func(state bool) {
		if state {
			result0 = "1"
		} else {
			result0 = "0"
		}
		fmt.Printf("freqDiv0.Q:    %v\n", state)
		fmt.Println(freqDiv0.latch.StateDump("freqDiv0"))
	})

	fmt.Println("\n==Wiring Up freqDiv0.QBar==")
	freqDiv0.QBar.WireUp(func(state bool) {
		fmt.Printf("freqDiv0.QBar: %v\n", state)
		fmt.Println(freqDiv0.latch.StateDump("freqDiv0"))
	})

	fmt.Println("\n==Wiring Up freqDiv1.Q==")
	freqDiv1.Q.WireUp(func(state bool) {
		if state {
			result1 = "1"
		} else {
			result1 = "0"
		}
		fmt.Printf("freqDiv1.Q:    %v\n", state)
		fmt.Println(freqDiv1.latch.StateDump("freqDiv1"))
	})
	fmt.Println("\n==Wiring Up freqDiv1.QBar==")
	freqDiv1.QBar.WireUp(func(state bool) {
		fmt.Printf("freqDiv1.QBar: %v\n", state)
		fmt.Println(freqDiv1.latch.StateDump("freqDiv1"))
	})

	fmt.Println("\n==Ripple Value==\n" + result1 + result0 + "\n")

	fmt.Println("==Clock On==")
	sw1.Set(true)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")
	fmt.Println("==Clock Off==")
	sw1.Set(false)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")

	fmt.Println("==Clock On==")
	sw1.Set(true)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")
	fmt.Println("==Clock Off==")
	sw1.Set(false)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")

	fmt.Println("==Clock On==")
	sw1.Set(true)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")
	fmt.Println("==Clock Off==")
	sw1.Set(false)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")

	fmt.Println("==Clock On==")
	sw1.Set(true)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")
	fmt.Println("==Clock Off==")
	sw1.Set(false)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")

	fmt.Println("==Clock On==")
	sw1.Set(true)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")
	fmt.Println("==Clock Off==")
	sw1.Set(false)
	fmt.Println("==Ripple Value==\n" + result1 + result0 + "\n")

}
*/
