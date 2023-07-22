package circuit

import (
	"flag"
	"fmt"
	"math/rand"
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
// go test -run TestRelay_WithChargeProviders -count 50 -trace out2.txt (go tool trace out2.txt)
// go test -race -cpu=1 -run TestFullAdder -count 5 -trace TestFullAdder_trace.txt > TestFullAdder_run.txt
// go test -debug (my own flag to write all the debug to the console during test run)

// go get golang.org/x/tools/cmd/cover
// go test -coverprofile cover.out (write code coverage file)
// go tool cover -html=cover.out

func init() {
	flag.BoolVar(&Debugging, "debug", false, "Enable verbose debugging to console")
}

func testName(t *testing.T, subtext string) string {
	return fmt.Sprintf("%s:%s", t.Name(), subtext)
}

func getAnswerString(gots []atomic.Value) string {
	answer := ""
	for i := 0; i < len(gots); i++ {
		if gots[i].Load() != nil && gots[i].Load().(bool) {
			answer += "1"
		} else {
			answer += "0"
		}
	}
	return answer
}
func TestChargeSource(t *testing.T) {
	cs := &chargeSource{}
	cs.Init()
	cs.Name = testName(t, "chargeSource")

	var want bool
	var got1, got2 atomic.Value
	ch1 := make(chan Charge, 1)
	ch2 := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c1 := <-ch1:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch1) Received on Channel (%v), Charge {%s}", ch1, c1.String()))
				got1.Store(c1.state)
				c1.Done()
			case c2 := <-ch2:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch2) Received on Channel (%v), Charge {%s}", ch2, c2.String()))
				got2.Store(c2.state)
				c2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	// two wire ups to prove both will get called
	cs.WireUp(ch1)
	cs.WireUp(ch2)

	// test default state (uncharged)
	want = false

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test charge transmit
	want = true
	cs.Transmit(Charge{state: want})

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test transmit loss of charge
	want = false
	cs.Transmit(Charge{state: want})

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test transmitting same state as last time (should skip it)
	cs.Transmit(Charge{state: want})

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did")
	default:
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did")
	default:
	}
}

func TestChargeProvider(t *testing.T) {
	cp := NewChargeProvider(testName(t, "ChargeProvider"), true)

	var want bool
	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	cp.WireUp(ch)

	// test default ChargeProvider state (charged)
	want = true

	if got.Load().(bool) != want {
		t.Errorf("With a new ChargeProvider, wanted the subscriber to see charge as %t but got %t", want, got.Load().(bool))
	}

	// test loss of charge
	cp.Discharge()
	want = false

	if got.Load().(bool) != want {
		t.Errorf("With a discharged ChargeProvider, wanted the subscriber's charge to be %t but got %t", want, got.Load().(bool))
	}

	// test re-added charge
	cp.Charge()
	want = true

	if got.Load().(bool) != want {
		t.Errorf("With a charged ChargeProvider, wanted the subscriber's charge to be %t but got %t", want, got.Load().(bool))
	}

	// test charging again (should skip it)
	cp.Charge()

	select {
	case <-ch:
		t.Error("Transmit of same state as prior state should have never gotten to ch, but it did")
	default:
	}
}

func TestWire(t *testing.T) {
	wire := NewWire(testName(t, "Wire"))
	defer wire.Shutdown()

	var want bool
	var got1, got2 atomic.Value
	ch1 := make(chan Charge, 1)
	ch2 := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c1 := <-ch1:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch1) Received on Channel (%v), Charge {%s}", ch1, c1.String()))
				got1.Store(c1.state)
				c1.Done()
			case c2 := <-ch2:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch2) Received on Channel (%v), Charge {%s}", ch2, c2.String()))
				got2.Store(c2.state)
				c2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	// two wire ups to prove both will get called
	wire.WireUp(ch1)
	wire.WireUp(ch2)

	// test default state (uncharged)
	want = false

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test charge transmit
	want = true
	wire.Transmit(Charge{state: want})

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test transmit loss of charge
	want = false
	wire.Transmit(Charge{state: want})

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test transmitting same state as last time (should skip it)
	wire.Transmit(Charge{state: want})

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did")
	default:
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did")
	default:
	}
}

func TestWire_WithDelay(t *testing.T) {
	var delay uint = 10

	wire := NewWire(testName(t, "Wire"))
	defer wire.Shutdown()

	wire.SetDelay(delay)

	var want bool
	var got1, got2 atomic.Value
	ch1 := make(chan Charge, 1)
	ch2 := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c1 := <-ch1:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch1) Received on Channel (%v), Charge {%s}", ch1, c1.String()))
				got1.Store(c1.state)
				c1.Done()
			case c2 := <-ch2:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch2) Received on Channel (%v), Charge {%s}", ch2, c2.String()))
				got2.Store(c2.state)
				c2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	// two wire ups to prove both will get called
	wire.WireUp(ch1)
	wire.WireUp(ch2)

	// test default state (uncharged)
	want = false

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test charge transmit
	want = true

	start := time.Now()
	wire.Transmit(Charge{state: want})
	end := time.Now()

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// validate wire delay
	gotDuration := end.Sub(start) // + time.Millisecond*1 // adding in just a little more to avoid timing edge case
	wantDuration := time.Millisecond * time.Duration(delay)
	if gotDuration < wantDuration {
		t.Errorf("Wire charge on transmit time should have been %v but was %v", wantDuration, gotDuration)
	}

	// test loss of charge transmit
	want = false

	start = time.Now()
	wire.Transmit(Charge{state: want})
	end = time.Now()

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// validate wire delay
	gotDuration = end.Sub(start) // + time.Millisecond*1 // adding in just a little more to avoid timing edge case
	wantDuration = time.Millisecond * time.Duration(delay)
	if gotDuration < wantDuration {
		t.Errorf("Wire charge loss transmit time should have been %v but was %v", wantDuration, gotDuration)
	}

	// test transmitting same state as last time (should skip it)
	wire.Transmit(Charge{state: want})

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did")
	default:
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did")
	default:
	}
}

func TestRibbonCable(t *testing.T) {
	rib := NewRibbonCable(testName(t, "RibbonCable"), 2)
	defer rib.Shutdown()

	rib.SetInputs(NewChargeProvider(testName(t, "ChargeProvider1"), false), NewChargeProvider(testName(t, "ChargeProvider2"), true))

	var want bool
	var got1, got2 atomic.Value
	ch1 := make(chan Charge, 1)
	ch2 := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c1 := <-ch1:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch1) Received on Channel (%v), Charge {%s}", ch1, c1.String()))
				got1.Store(c1.state)
				c1.Done()
			case c2 := <-ch2:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch2) Received on Channel (%v), Charge {%s}", ch2, c2.String()))
				got2.Store(c2.state)
				c2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	rib.Wires[0].(*Wire).WireUp(ch1)
	rib.Wires[1].(*Wire).WireUp(ch2)

	// the first wire in the ribbon cable had a dead ChargeProvider
	want = false

	if got1.Load().(bool) != want {
		t.Errorf("Left ChargeProvider off, wanted the wire to see charge as %t but got %t", want, got1.Load().(bool))
	}

	// the first wire in the ribbon cable had a live ChargeProvider
	want = true

	if got2.Load().(bool) != want {
		t.Errorf("Right ChargeProvider on, wanted the wire to see charge as %t but got %t", want, got2.Load().(bool))
	}
}

func TestRelay_WithChargeProviders(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
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
		{true, true, false, true}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	var pin1ChargeProvider, pin2ChargeProvider *ChargeProvider
	pin1ChargeProvider = NewChargeProvider(testName(t, "ChargeProvider1"), true)
	pin2ChargeProvider = NewChargeProvider(testName(t, "ChargeProvider2"), true)

	rel := NewRelay(testName(t, "Relay"), pin1ChargeProvider, pin2ChargeProvider)
	defer rel.Shutdown()

	var gotOpenOut, gotClosedOut atomic.Value
	chOpen := make(chan Charge, 1)
	chClosed := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case cOpen := <-chOpen:
				Debug(testName(t, "Select"), fmt.Sprintf("(chOpen) Received on Channel (%v), Charge {%s}", chOpen, cOpen.String()))
				gotOpenOut.Store(cOpen.state)
				cOpen.Done()
			case cClosed := <-chClosed:
				Debug(testName(t, "Select"), fmt.Sprintf("(chClosed) Received on Channel (%v), Charge {%s}", chClosed, cClosed.String()))
				gotClosedOut.Store(cClosed.state)
				cClosed.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	rel.OpenOut.WireUp(chOpen)
	rel.ClosedOut.WireUp(chClosed)

	if gotOpenOut.Load().(bool) {
		t.Error("Wanted no charge at the open position but got some")
	}
	if !gotClosedOut.Load().(bool) {
		t.Error("Wanted charge at the closed position but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting input A to (%t) and B to (%t)", i, tc.aInHasCharge, tc.bInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting input A to (%t) and B to (%t)", i, tc.aInHasCharge, tc.bInHasCharge))

			if tc.aInHasCharge {
				pin1ChargeProvider.Charge()
			} else {
				pin1ChargeProvider.Discharge()
			}

			if tc.bInHasCharge {
				pin2ChargeProvider.Charge()
			} else {
				pin2ChargeProvider.Discharge()
			}

			if gotOpenOut.Load().(bool) != tc.wantAtOpen {
				t.Errorf("Wanted charge at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut.Load().(bool))
			}

			if gotClosedOut.Load().(bool) != tc.wantAtClosed {
				t.Errorf("Wanted charge at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut.Load().(bool))
			}
		})
	}

	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestSwitch(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")

	sw := NewSwitch(testName(t, "Switch"), false)
	defer sw.Shutdown()

	var want bool
	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	sw.WireUp(ch)

	// test initial off switch state
	want = false

	if got.Load().(bool) != want {
		t.Errorf("With an off switch, wanted the subscriber to see charge as %t but got %t", want, got.Load().(bool))
	}

	Debug(testName(t, ""), "Start Test Cases")

	// initial turn on
	want = true
	sw.Set(want)

	if got.Load().(bool) != want {
		t.Errorf("With an off switch turned on, wanted the subscriber to see charge as %t but got %t", want, got.Load().(bool))
	}

	// turn on again, though already on ('want' is already true from prior Set)
	sw.Set(want)

	if got.Load().(bool) != want {
		t.Errorf("With an attempt to turn on an already on switch, wanted the channel to be empty, but it wasn't")
	}

	// now off
	want = false
	sw.Set(want)

	if got.Load().(bool) != want {
		t.Errorf("With an on switch turned off, wanted the subscriber to see charge as %t but got %t", want, got.Load().(bool))
	}

	Debug(testName(t, ""), "End Test Cases")
}

func TestNSwitchBank_BadInputs(t *testing.T) {
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
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting switches to (%s)", i, tc.input), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting switches to (%s)", i, tc.input))

			sb, err := NewNSwitchBank(testName(t, "NSwitchBank"), tc.input)

			if sb != nil {
				sb.Shutdown()
				t.Error("Didn't expected a Switch Bank back but got one")
			}

			tc.wantError += "\"" + tc.input + "\""

			if err != nil && err.Error() != tc.wantError {
				t.Errorf("Wanted error \"%s\" but got \"%v\"", tc.wantError, err.Error())
			}
		})
	}

	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestNSwitchBank_GoodInputs(t *testing.T) {
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
		{"1010101010101010", []bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false}},
		{"1000000000000001", []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true}},
		{"0000000000000000", []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
		{"1111111111111111", []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}},
		{"0000000000000000", []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}}, // final test ensuring we can toggle all inputs fully reversed again
	}

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Contructing NSwitchBank with switches set to (%s)", i, tc.input), func(t *testing.T) {
			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Contructing NSwitchBank with switches set to (%s)", i, tc.input))
			Debug(testName(t, ""), "Initial Setup")

			sb, err := NewNSwitchBank(testName(t, "NSwitchBank"), tc.input)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
			}
			defer sb.Shutdown()

			Debug(testName(t, ""), "Start Test Cases WireUp Per Switch")

			// will just check one switch at a time vs. trying to get some full answer in one go from the bank
			for i, sw := range sb.Switches() {

				sw.(*Switch).WireUp(ch)

				want := tc.want[i]

				if got.Load().(bool) != want {
					t.Errorf("At index %d, wanted %t but got %t", i, want, got.Load().(bool))
				}
			}
			Debug(testName(t, ""), "End Test Cases WireUp Per Switch")
		})
	}

	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestNSwitchBank_GoodInputs_SetSwitches(t *testing.T) {
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
		{"1010101010101010", []bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false}},
		{"1000000000000001", []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true}},
		{"0000000000000000", []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
		{"1111111111111111", []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}},
		{"0000000000000000", []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}}, // final test ensuring we can toggle all inputs fully reversed again
	}

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Contructing NSwitchBank with switches set to (%s)", i, tc.input), func(t *testing.T) {
			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Contructing NSwitchBank with switches set to (%s)", i, tc.input))
			Debug(testName(t, ""), "Initial Setup")

			// weak attempt to have some random 1s and 0s as the initial switch bank setup on construction
			randomInput := make([]byte, len(tc.input))
			for i := range randomInput {
				randomInput[i] = "000000111111"[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(12)]
				time.Sleep(time.Millisecond * 10)

			}

			sb, err := NewNSwitchBank(testName(t, "NSwitchBank"), string(randomInput))
			sb.SetSwitches(tc.input)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
			}
			defer sb.Shutdown()

			Debug(testName(t, ""), "Start Test Cases WireUp Per Switch")

			// will just check one switch at a time vs. trying to get some full answer in one go from the bank
			for i, sw := range sb.Switches() {

				sw.(*Switch).WireUp(ch)

				want := tc.want[i]

				if got.Load().(bool) != want {
					t.Errorf("At index %d, wanted %t but got %t", i, want, got.Load().(bool))
				}
			}
			Debug(testName(t, ""), "End Test Cases WireUp Per Switch")
		})
	}

	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestRelay_WithSwitches(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
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
		{true, true, false, true}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), true)
	defer bSwitch.Shutdown()

	rel := NewRelay(testName(t, "Relay"), aSwitch, bSwitch)
	defer rel.Shutdown()

	var gotOpenOut, gotClosedOut atomic.Value
	chOpen := make(chan Charge, 1)
	chClosed := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case cOpen := <-chOpen:
				Debug(testName(t, "Select"), fmt.Sprintf("(chOpen) Received on Channel (%v), Charge {%s}", chOpen, cOpen.String()))
				gotOpenOut.Store(cOpen.state)
				cOpen.Done()
			case cClosed := <-chClosed:
				Debug(testName(t, "Select"), fmt.Sprintf("(chClosed) Received on Channel (%v), Charge {%s}", chClosed, cClosed.String()))
				gotClosedOut.Store(cClosed.state)
				cClosed.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	rel.OpenOut.WireUp(chOpen)
	rel.ClosedOut.WireUp(chClosed)

	if gotOpenOut.Load().(bool) {
		t.Error("Wanted no charge at the open position but got some")
	}
	if !gotClosedOut.Load().(bool) {
		t.Error("Wanted charge at the closed position but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)

			if gotOpenOut.Load().(bool) != tc.wantAtOpen {
				t.Errorf("Wanted charge at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut.Load().(bool))
			}

			if gotClosedOut.Load().(bool) != tc.wantAtClosed {
				t.Errorf("Wanted charge at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestANDGate(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
		cInHasCharge bool
		want         bool
	}{
		{false, false, false, false},
		{true, false, false, false},
		{false, true, false, false},
		{false, false, true, false},
		{true, true, false, false},
		{true, false, true, false},
		{false, true, true, false},
		{true, true, true, true},
		{false, false, false, false},
		{true, true, true, true}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), true)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(testName(t, "cSwitch"), true)
	defer cSwitch.Shutdown()

	gate := NewANDGate(testName(t, "ANDGate"), aSwitch, bSwitch, cSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if !got.Load().(bool) {
		t.Error("Wanted charge on the gate but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t) and C charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.cInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t) and C charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.cInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)
			cSwitch.Set(tc.cInHasCharge)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted charge %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

// For testing behavior off looped AND Gates
// http://www.falstad.com/circuit/e-ohms.html
// $ 1 0.000005 10.20027730826997 50 5 48 5e-11
// 172 288 128 320 128 0 7 3.55 5 0 0 0.5 Voltage
// g 240 352 240 368 0 0
// w 240 320 240 352 1
// w 400 320 400 352 1
// g 400 352 400 384 0 0
// 172 416 128 368 128 0 7 3.55 5 0 0 0.5 Voltage
// w 384 144 384 160 0
// 162 240 256 240 320 2 default-led 1 0 0 0.01
// 162 400 256 400 320 2 default-led 1 0 0 0.01
// s 256 128 288 128 0 1 false
// w 256 128 256 160 0
// s 416 128 480 128 0 1 false
// w 480 128 416 160 0
// w 240 224 240 256 0
// w 240 256 352 144 0
// w 352 144 384 144 0
// w 224 160 224 96 0
// w 224 96 496 96 0
// w 496 96 496 240 0
// w 400 224 400 256 0
// w 400 256 496 256 0
// w 496 240 496 256 0
// 150 240 160 240 224 0 2 0 5
// 150 400 160 400 224 0 2 0 5

func TestLoopedANDGates(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")
	name := testName(t, "TestLoopedANDGates")

	wire1 := NewWire(fmt.Sprintf("%s-Wire1", name))
	defer wire1.Shutdown()
	wire2 := NewWire(fmt.Sprintf("%s-Wire2", name))
	defer wire2.Shutdown()

	// defaulting one input pin on gate1 to off will ensure a false output on startup
	gate1pin1ChargeProvider := NewChargeProvider("gate1pin1ChargeProvider", false)
	gate1 := NewANDGate(fmt.Sprintf("%s-ANDGate1", name), gate1pin1ChargeProvider, wire1)
	defer gate1.Shutdown()
	gate1.WireUp(wire2.Input)

	// defaulting one input pin on gate2 to off will ensure a false output on startup
	gate2pin1ChargeProvider := NewChargeProvider("gate2pin1ChargeProvider", false)
	gate2 := NewANDGate(fmt.Sprintf("%s-ANDGate2", name), gate2pin1ChargeProvider, wire2)
	defer gate2.Shutdown()
	gate2.WireUp(wire1.Input)

	var gotGate1State, gotGate2State atomic.Value
	chGate1 := make(chan Charge, 1)
	chGate2 := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case cGate1 := <-chGate1:
				Debug(testName(t, "Select"), fmt.Sprintf("(chGate1) Received on Channel (%v), Charge {%s}", chGate1, cGate1.String()))
				gotGate1State.Store(cGate1.state)
				cGate1.Done()
			case cGate2 := <-chGate2:
				Debug(testName(t, "Select"), fmt.Sprintf("(chGate2) Received on Channel (%v), Charge {%s}", chGate2, cGate2.String()))
				gotGate2State.Store(cGate2.state)
				cGate2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate1.WireUp(chGate1)
	gate2.WireUp(chGate2)

	// a set of looped together AND gates settle out as both off
	if gotGate1State.Load().(bool) {
		t.Error("Wanted no charge on Gate1's output but got one")
	}
	if gotGate2State.Load().(bool) {
		t.Error("Wanted no charge on Gate2's output but got one")
	}

	// even if charged the non-looped wire input pin on gate 1, the other is off so still false on both AND gates
	gate1pin1ChargeProvider.Charge()
	if gotGate1State.Load().(bool) {
		t.Error("Wanted no charge on Gate1's output but got one")
	}
	if gotGate2State.Load().(bool) {
		t.Error("Wanted no charge on Gate2's output but got one")
	}

	// even if charged the non-looped wire input pin on gate 2 as well, the other is off so still false on both AND gates
	gate2pin1ChargeProvider.Charge()
	if gotGate1State.Load().(bool) {
		t.Error("Wanted no charge on Gate1's output but got one")
	}

	if gotGate2State.Load().(bool) {
		t.Error("Wanted no charge on Gate2's output but got one")
	}
}

func TestORGate(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
		cInHasCharge bool
		want         bool
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
		{true, true, true, true}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), true)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(testName(t, "cSwitch"), true)
	defer cSwitch.Shutdown()

	gate := NewORGate(testName(t, "ORGate"), aSwitch, bSwitch, cSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if !got.Load().(bool) {
		t.Error("Wanted charge on the gate but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t) and C charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.cInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t) and C charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.cInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)
			cSwitch.Set(tc.cInHasCharge)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted charge %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

// For testing behavior off looped OR Gates
// http://www.falstad.com/circuit/e-ohms.html
// $ 1 0.000005 10.20027730826997 50 5 48 5e-11
// 172 288 128 320 128 0 7 3.55 5 0 0 0.5 Voltage
// g 240 352 240 368 0 0
// w 240 320 240 352 1
// w 400 320 400 352 1
// g 400 352 400 384 0 0
// 152 240 160 240 208 0 2 5 5
// 172 416 128 368 128 0 7 3.55 5 0 0 0.5 Voltage
// 152 400 160 400 224 0 2 5 5
// w 384 144 384 160 0
// 162 240 256 240 320 2 default-led 1 0 0 0.01
// 162 400 256 400 320 2 default-led 1 0 0 0.01
// s 256 128 288 128 0 1 false
// w 256 128 256 160 0
// s 416 128 480 128 0 1 false
// w 480 128 416 160 0
// w 240 208 240 256 0
// w 240 256 352 144 0
// w 352 144 384 144 0
// w 224 160 224 96 0
// w 224 96 496 96 0
// w 496 96 496 240 0
// w 400 224 400 256 0
// w 400 256 496 256 0
// w 496 240 496 256 0

func TestLoopedORGates(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")
	name := testName(t, "TestLoopedORGates")

	wire1 := NewWire(fmt.Sprintf("%s-Wire1", name))
	defer wire1.Shutdown()
	wire2 := NewWire(fmt.Sprintf("%s-Wire2", name))
	defer wire2.Shutdown()

	// defaulting one input pin on gate1 to off will ensure a false output on startup
	gate1pin1ChargeProvider := NewChargeProvider("gate1pin1ChargeProvider", false)
	gate1 := NewORGate(fmt.Sprintf("%s-ORGate1", name), gate1pin1ChargeProvider, wire1)
	defer gate1.Shutdown()
	gate1.WireUp(wire2.Input)

	// defaulting one input pin on gate2 to off will ensure a false output on startup
	gate2pin1ChargeProvider := NewChargeProvider("gate2pin1ChargeProvider", false)
	gate2 := NewORGate(fmt.Sprintf("%s-ORGate2", name), gate2pin1ChargeProvider, wire2)
	defer gate2.Shutdown()
	gate2.WireUp(wire1.Input)

	var gotGate1State, gotGate2State atomic.Value
	chGate1 := make(chan Charge, 1)
	chGate2 := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case cGate1 := <-chGate1:
				Debug(testName(t, "Select"), fmt.Sprintf("(chGate1) Received on Channel (%v), Charge {%s}", chGate1, cGate1.String()))
				gotGate1State.Store(cGate1.state)
				cGate1.Done()
			case cGate2 := <-chGate2:
				Debug(testName(t, "Select"), fmt.Sprintf("(chGate2) Received on Channel (%v), Charge {%s}", chGate2, cGate2.String()))
				gotGate2State.Store(cGate2.state)
				cGate2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate1.WireUp(chGate1)
	gate2.WireUp(chGate2)

	// a set of looped together OR gates settle out as both off (since the secondary inputs on each were not charged during setup)
	if gotGate1State.Load().(bool) {
		t.Error("Wanted no charge on Gate1's output but got one")
	}
	if gotGate2State.Load().(bool) {
		t.Error("Wanted no charge on Gate2's output but got one")
	}

	// if charged the non-looped input pin on gate 1, the gate 1 OR is true which makes the other one true too since linked
	gate1pin1ChargeProvider.Charge()
	if !gotGate1State.Load().(bool) {
		t.Error("Wanted charge on Gate1's output but got none")
	}
	if !gotGate2State.Load().(bool) {
		t.Error("Wanted charge on Gate2's output but got none")
	}

	// if charged the non-looped input pin on gate 2, the gate 2 OR is true which makes the other one true too since linked (though the other one is already true due to prior test)
	gate2pin1ChargeProvider.Charge()
	if !gotGate1State.Load().(bool) {
		t.Error("Wanted charge on Gate1's output but got none")
	}
	if !gotGate2State.Load().(bool) {
		t.Error("Wanted charge on Gate2's output but got none")
	}

	// if then discharged the non-looped input pin on gate 1, the gate 1 OR is still true since gate 2 is returning true
	gate1pin1ChargeProvider.Discharge()
	if !gotGate1State.Load().(bool) {
		t.Error("Wanted charge on Gate1's output but got none")
	}
	if !gotGate2State.Load().(bool) {
		t.Error("Wanted charge on Gate2's output but got none")
	}

	// but if finally discharged the non-looped input pin on gate 2 too, it is "remembering" and both stay on? (what really happens in a real circuit in this case? this?)
	gate2pin1ChargeProvider.Discharge()
	if !gotGate1State.Load().(bool) {
		t.Error("Wanted charge on Gate1's output but got none")
	}
	if !gotGate2State.Load().(bool) {
		t.Error("Wanted charge on Gate2's output but got none")
	}
}

func TestNANDGate(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
		cInHasCharge bool
		want         bool
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
		{true, true, true, false}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), false)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(testName(t, "cSwitch"), true)
	defer cSwitch.Shutdown()

	gate := NewNANDGate(testName(t, "NANDGate"), aSwitch, bSwitch, cSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if !got.Load().(bool) {
		t.Error("Wanted charge on the gate but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t) and C charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.cInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t) and C charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.cInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)
			cSwitch.Set(tc.cInHasCharge)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted charge %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestNORGate(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
		cInHasCharge bool
		want         bool
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
		{true, true, true, false}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), true)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), true)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(testName(t, "cSwitch"), true)
	defer cSwitch.Shutdown()

	gate := NewNORGate(testName(t, "NORGate"), aSwitch, bSwitch, cSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if got.Load().(bool) {
		t.Error("Wanted no charge on the gate but got some")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t) and C charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.cInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t) and C charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.cInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)
			cSwitch.Set(tc.cInHasCharge)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted charge %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestXORGate(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
		want         bool
	}{
		{false, false, false},
		{true, false, true},
		{false, true, true},
		{true, true, false},
		{false, false, false},
		{true, true, false}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), false)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), false)
	defer bSwitch.Shutdown()

	gate := NewXORGate(testName(t, "XORGate"), aSwitch, bSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if got.Load().(bool) {
		t.Error("Wanted no charge on the gate but got some")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted charge %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestInverter(t *testing.T) {
	testCases := []struct {
		hasCharge bool
		wantOut   bool
	}{
		{false, true},
		{true, false},
		{false, true},
		{true, false},
	}

	Debug(testName(t, ""), "Initial Setup")

	pin1ChargeProvider := NewChargeProvider(testName(t, "ChargeProvider"), true)
	inv := NewInverter(testName(t, "Inverter"), pin1ChargeProvider)
	defer inv.Shutdown()

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	inv.WireUp(ch)

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Input as (%t)", i, tc.hasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Input as (%t)", i, tc.hasCharge))

			if tc.hasCharge {
				pin1ChargeProvider.Charge()
			} else {
				pin1ChargeProvider.Discharge()
			}

			if got.Load().(bool) != tc.wantOut {
				t.Errorf("Input charge was %t so wanted it inverted to %t but got %t", tc.hasCharge, tc.wantOut, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestXNORGate(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
		want         bool
	}{
		{true, false, false},
		{false, true, false},
		{true, true, true},
		{false, false, true},
		{true, false, false}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), false)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), false)
	defer bSwitch.Shutdown()

	gate := NewXNORGate(testName(t, "XNORGate"), aSwitch, bSwitch)
	defer gate.Shutdown()

	var got atomic.Value
	ch := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
				got.Store(c.state)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if !got.Load().(bool) {
		t.Error("Wanted charge on the gate but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A charge to (%t) and B charge to (%t)", i, tc.aInHasCharge, tc.bInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted charge %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestHalfAdder(t *testing.T) {
	testCases := []struct {
		aInHasCharge bool
		bInHasCharge bool
		wantSum      bool
		wantCarry    bool
	}{

		{false, false, false, false},
		{true, false, true, false},
		{false, true, true, false},
		{true, true, false, true},
		{false, false, false, false},
		{true, true, false, true}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), false)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), false)
	defer bSwitch.Shutdown()

	half := NewHalfAdder(testName(t, "HalfAdder"), aSwitch, bSwitch)
	defer half.Shutdown()

	var gotSum, gotCarry atomic.Value
	chSum := make(chan Charge, 1)
	chCarry := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case cSum := <-chSum:
				Debug(testName(t, "Select"), fmt.Sprintf("(chSum) Received on Channel (%v), Charge {%s}", chSum, cSum.String()))
				gotSum.Store(cSum.state)
				cSum.Done()
			case cCarry := <-chCarry:
				Debug(testName(t, "Select"), fmt.Sprintf("(chCarry) Received on Channel (%v), Charge {%s}", chCarry, cCarry.String()))
				gotCarry.Store(cCarry.state)
				cCarry.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	half.Sum.WireUp(chSum)
	half.Carry.WireUp(chCarry)

	if gotSum.Load().(bool) {
		t.Error("Wanted no Sum but got one")
	}

	if gotCarry.Load().(bool) {
		t.Error("Wanted no Carry but got one")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting input source A to (%t) and source B to (%t)", i, tc.aInHasCharge, tc.bInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting input source A to (%t) and source B to (%t)", i, tc.aInHasCharge, tc.bInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)

			if gotSum.Load().(bool) != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, gotSum.Load().(bool))
			}

			if gotCarry.Load().(bool) != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, gotCarry.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestFullAdder(t *testing.T) {
	testCases := []struct {
		aInHasCharge     bool
		bInHasCharge     bool
		carryInHasCharge bool
		wantSum          bool
		wantCarry        bool
	}{
		{false, false, false, false, false},
		{true, false, false, true, false},
		{true, true, false, false, true},
		{true, true, true, true, true},
		{false, true, false, true, false},
		{false, true, true, false, true},
		{false, false, true, true, false},
		{true, false, true, false, true},
		{true, true, true, true, true},
		{false, false, false, false, false},
		{true, true, true, true, true}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	aSwitch := NewSwitch(testName(t, "aSwitch"), false)
	defer aSwitch.Shutdown()

	bSwitch := NewSwitch(testName(t, "bSwitch"), false)
	defer bSwitch.Shutdown()

	cSwitch := NewSwitch(testName(t, "cSwitch"), false)
	defer cSwitch.Shutdown()

	full := NewFullAdder(testName(t, "FullAdder"), aSwitch, bSwitch, cSwitch)
	defer full.Shutdown()

	var gotSum, gotCarry atomic.Value
	chSum := make(chan Charge, 1)
	chCarry := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case cSum := <-chSum:
				Debug(testName(t, "Select"), fmt.Sprintf("(chSum) Received on Channel (%v), Charge {%s}", chSum, cSum.String()))
				gotSum.Store(cSum.state)
				cSum.Done()
			case cCarry := <-chCarry:
				Debug(testName(t, "Select"), fmt.Sprintf("(chCarry) Received on Channel (%v), Charge {%s}", chCarry, cCarry.String()))
				gotCarry.Store(cCarry.state)
				cCarry.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	full.Sum.WireUp(chSum)
	full.Carry.WireUp(chCarry)

	if gotSum.Load().(bool) {
		t.Error("Wanted no Sum but got one")
	}

	if gotCarry.Load().(bool) {
		t.Error("Wanted no Carry but got one")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting input source A to (%t) and source B to (%t) with carry in of (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.carryInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting input source A to (%t) and source B to (%t) with carry in of (%t)", i, tc.aInHasCharge, tc.bInHasCharge, tc.carryInHasCharge))

			aSwitch.Set(tc.aInHasCharge)
			bSwitch.Set(tc.bInHasCharge)
			cSwitch.Set(tc.carryInHasCharge)

			if gotSum.Load().(bool) != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, gotSum.Load().(bool))
			}

			if gotCarry.Load().(bool) != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, gotCarry.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Adding (%s) to (%s)", i, tc.byte1, tc.byte2), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Adding (%s) to (%s)", i, tc.byte1, tc.byte2))

			addend1Switches, _ := NewNSwitchBank(testName(t, "addend1Switches"), tc.byte1)
			defer addend1Switches.Shutdown()

			addend2Switches, _ := NewNSwitchBank(testName(t, "addend2Switches"), tc.byte2)
			defer addend2Switches.Shutdown()

			addr, err := NewNBitAdder(testName(t, "NBitAdder"), addend1Switches.Switches(), addend2Switches.Switches(), nil)

			if addr != nil {
				addr.Shutdown()
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
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestNBitAdder_EightBit(t *testing.T) {
	testCases := []struct {
		byte1            string
		byte2            string
		carryInHasCharge bool
		wantAnswer       string
		wantCarryOut     bool
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

	Debug(testName(t, ""), "Initial Setup")

	// start with off switches
	addend1Switches, _ := NewNSwitchBank(testName(t, "addend1Switches"), "00000000")
	defer addend1Switches.Shutdown()

	addend2Switches, _ := NewNSwitchBank(testName(t, "addend2Switches"), "00000000")
	defer addend2Switches.Shutdown()

	carryInSwitch := NewSwitch(testName(t, "carryInSwitch"), false)
	defer carryInSwitch.Shutdown()

	// create the adder based on those switches
	addr, err := NewNBitAdder(testName(t, "NBitAdder"), addend1Switches.Switches(), addend2Switches.Switches(), carryInSwitch)

	if err != nil {
		t.Errorf("Expected no error on construction, but got: %s", err.Error())
	}

	if addr == nil {
		t.Error("Expected an adder to return due to good inputs, but got a nil one")
	} else {
		defer addr.Shutdown()
	}

	var gots [9]atomic.Value // 0-7 for sums, 8 for carryout
	var chStates []chan Charge
	var chStops []chan bool
	for i := 0; i < 9; i++ {
		gots[i].Store(false)
		chStates = append(chStates, make(chan Charge, 1))
		chStops = append(chStops, make(chan bool, 1))
		go func(chState chan Charge, chStop chan bool, i int) {
			for {
				select {
				case c := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
					gots[i].Store(c.state)
					c.Done()
				case <-chStop:
					return
				}
			}
		}(chStates[i], chStops[i], i)
	}
	defer func() {
		for i := 0; i < 9; i++ {
			chStops[i] <- true
		}
	}()

	for i, sum := range addr.Sums {
		sum.WireUp(chStates[i])
	}

	addr.CarryOut.WireUp(chStates[8])

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Adding (%s) to (%s) with carry in of (%t)", i, tc.byte1, tc.byte2, tc.carryInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Adding (%s) to (%s) with carry in of (%t)", i, tc.byte1, tc.byte2, tc.carryInHasCharge))

			addend1Switches.SetSwitches(tc.byte1)
			addend2Switches.SetSwitches(tc.byte2)
			carryInSwitch.Set(tc.carryInHasCharge)

			if gotAnswer := getAnswerString(gots[:8]); gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, gotAnswer)
			}

			if gots[8].Load().(bool) != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, gots[8].Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestNBitAdder_SixteenBit(t *testing.T) {
	testCases := []struct {
		bytes1           string
		bytes2           string
		carryInHasCharge bool
		wantAnswer       string
		wantCarryOut     bool
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

	Debug(testName(t, ""), "Initial Setup")

	// start with off switches
	addend1Switches, _ := NewNSwitchBank(testName(t, "addend1Switches"), "0000000000000000")
	defer addend1Switches.Shutdown()

	addend2Switches, _ := NewNSwitchBank(testName(t, "addend2Switches"), "0000000000000000")
	defer addend2Switches.Shutdown()

	carryInSwitch := NewSwitch(testName(t, "carryInSwitch"), false)
	defer carryInSwitch.Shutdown()

	// create the adder based on those switches
	addr, err := NewNBitAdder(testName(t, "NBitAdder"), addend1Switches.Switches(), addend2Switches.Switches(), carryInSwitch)

	if err != nil {
		t.Errorf("Expected no error on construction, but got: %s", err.Error())
	}

	if addr == nil {
		t.Error("Expected an adder to return due to good inputs, but got a nil one")
	} else {
		defer addr.Shutdown()
	}

	var gots [17]atomic.Value // 0-15 for sums, 16 for carryout
	var chStates []chan Charge
	var chStops []chan bool
	for i := 0; i < 17; i++ {
		gots[i].Store(false)
		chStates = append(chStates, make(chan Charge, 1))
		chStops = append(chStops, make(chan bool, 1))
		go func(chState chan Charge, chStop chan bool, i int) {
			for {
				select {
				case c := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
					gots[i].Store(c.state)
					c.Done()
				case <-chStop:
					return
				}
			}
		}(chStates[i], chStops[i], i)
	}
	defer func() {
		for i := 0; i < 17; i++ {
			chStops[i] <- true
		}
	}()

	for i, sum := range addr.Sums {
		sum.WireUp(chStates[i])
	}

	addr.CarryOut.WireUp(chStates[16])

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Adding (%s) to (%s) with carry in of (%t)", i, tc.bytes1, tc.bytes2, tc.carryInHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Adding (%s) to (%s) with carry in of (%t)", i, tc.bytes1, tc.bytes2, tc.carryInHasCharge))

			addend1Switches.SetSwitches(tc.bytes1)
			addend2Switches.SetSwitches(tc.bytes2)
			carryInSwitch.Set(tc.carryInHasCharge)

			if gotAnswer := getAnswerString(gots[:16]); gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, gotAnswer)
			}

			if gots[16].Load().(bool) != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, gots[16].Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestOnesCompliment(t *testing.T) {

	testCases := []struct {
		bits            string
		signalHasCharge bool
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

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Executing complementer against (%s) with compliment signal of (%t)", i, tc.bits, tc.signalHasCharge), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Executing complementer against (%s) with compliment signal of (%t)", i, tc.bits, tc.signalHasCharge))

			bitSwitches, _ := NewNSwitchBank(testName(t, "bitSwitches"), tc.bits)
			defer bitSwitches.Shutdown()

			signalSwitch := NewSwitch(testName(t, "signalSwitch"), tc.signalHasCharge)
			defer signalSwitch.Shutdown()

			comp := NewOnesComplementer(testName(t, "OnesComplementer"), bitSwitches.Switches(), signalSwitch)

			if comp == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one")
			} else {
				defer comp.Shutdown()
			}

			gotCompliments := make([]atomic.Value, len(tc.bits))
			var chStates []chan Charge
			var chStops []chan bool
			for i := 0; i < len(tc.bits); i++ {
				gotCompliments[i].Store(false)
				chStates = append(chStates, make(chan Charge, 1))
				chStops = append(chStops, make(chan bool, 1))
				go func(chState chan Charge, chStop chan bool, index int) {
					for {
						select {
						case c := <-chState:
							Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
							gotCompliments[index].Store(c.state)
							c.Done()
						case <-chStop:
							return
						}
					}
				}(chStates[i], chStops[i], i)
			}
			defer func() {
				for i := 0; i < len(tc.bits); i++ {
					chStops[i] <- true
				}
			}()

			for i, c := range comp.Complements {
				c.WireUp(chStates[i])
			}

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
	Debug(testName(t, ""), "End Test Cases Loop")
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

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Adding (%s) to (%s)", i, tc.byte1, tc.byte2), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Adding (%s) to (%s)", i, tc.byte1, tc.byte2))

			minuendSwitches, _ := NewNSwitchBank(testName(t, "minuendSwitches"), tc.byte1)
			defer minuendSwitches.Shutdown()

			subtrahendSwitches, _ := NewNSwitchBank(testName(t, "subtrahendSwitches"), tc.byte2)
			defer subtrahendSwitches.Shutdown()

			sub, err := NewNBitSubtractor(testName(t, "NBitSubtractor"), minuendSwitches.Switches(), subtrahendSwitches.Switches())

			if sub != nil {
				sub.Shutdown()
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
	Debug(testName(t, ""), "End Test Cases Loop")
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

	Debug(testName(t, ""), "Initial Setup")

	// start with off switches
	minuendwitches, _ := NewNSwitchBank(testName(t, "minuendSwitches"), "00000000")
	defer minuendwitches.Shutdown()

	subtrahendSwitches, _ := NewNSwitchBank(testName(t, "subtrahendSwitches"), "00000000")
	defer subtrahendSwitches.Shutdown()

	sub, err := NewNBitSubtractor(testName(t, "NBitSubtractor"), minuendwitches.Switches(), subtrahendSwitches.Switches())

	if err != nil {
		t.Errorf("Expected no error on construction, but got: %s", err.Error())
	}

	if sub == nil {
		t.Error("Expected a subtractor to return due to good inputs, but got a nil one")
	} else {
		defer sub.Shutdown()
	}

	var gots [9]atomic.Value // 0-7 for diffs, 8 for carryout
	var chStates []chan Charge
	var chStops []chan bool
	for i := 0; i < 9; i++ {
		gots[i].Store(false)
		chStates = append(chStates, make(chan Charge, 1))
		chStops = append(chStops, make(chan bool, 1))
		go func(chState chan Charge, chStop chan bool, index int) {
			for {
				select {
				case c := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
					gots[index].Store(c.state)
					c.Done()
				case <-chStop:
					return
				}
			}
		}(chStates[i], chStops[i], i)
	}
	defer func() {
		for i := 0; i < 9; i++ {
			chStops[i] <- true
		}
	}()

	for i, dif := range sub.Differences {
		dif.WireUp(chStates[i])
	}

	sub.CarryOut.WireUp(chStates[8])

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Subtracting %s from %s", i, tc.subtrahend, tc.minuend), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Subtracting %s from %s", i, tc.subtrahend, tc.minuend))

			minuendwitches.SetSwitches(tc.minuend)
			subtrahendSwitches.SetSwitches(tc.subtrahend)

			if gotAnswer := getAnswerString(gots[:8]); gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, gotAnswer)
			}

			if gots[8].Load().(bool) != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, gots[8].Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Oscillating at (%d) hertz, immediate start (%t)", i, tc.oscHertz, tc.initState), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Oscillating at (%d) hertz, immediate start (%t)", i, tc.oscHertz, tc.initState))

			osc := NewOscillator(testName(t, "Oscillator"), tc.initState)
			defer osc.Stop()

			var gotResults atomic.Value
			ch := make(chan Charge, 1)

			gotResults.Store("")
			chStop := make(chan bool, 1)
			go func() {
				for {
					result := gotResults.Load().(string)
					select {
					case c := <-ch:
						Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
						if c.state {
							result += "1"
						} else {
							result += "0"
						}
						gotResults.Store(result)
						c.Done()
					case <-chStop:
						return
					}
				}
			}()
			defer func() { chStop <- true }()

			osc.WireUp(ch)
			osc.Oscillate(tc.oscHertz)

			// give the oscillator some time to....oscillate
			time.Sleep(time.Second * 3)

			if !strings.HasPrefix(gotResults.Load().(string), tc.wantResults) {
				t.Errorf("Wanted results of at least %s but got %s", tc.wantResults, gotResults.Load().(string))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

// func TestOscillator2(t *testing.T) {
// 	testCases := []struct {
// 		initState   bool
// 		oscHertz    int
// 		wantResults string
// 	}{
// 		{false, 1, "010"},
// 		// {true, 1, "101"},
// 		// {false, 5, "01010101010"},
// 		// {true, 5, "10101010101"},
// 	}

// 	Debug(testName(t, ""), "Start Test Cases Loop")

// 	for i, tc := range testCases {
// 		t.Run(fmt.Sprintf("testCases[%d]: Oscillating at (%d) hertz, immediate start (%t)", i, tc.oscHertz, tc.initState), func(t *testing.T) {

// 			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Oscillating at (%d) hertz, immediate start (%t)", i, tc.oscHertz, tc.initState))

// 			osc := NewOscillator2(testName(t, "Oscillator"), tc.initState)
// 			defer osc.Stop()

// 			var gotResults atomic.Value
// 			ch := make(chan Charge, 1)

// 			gotResults.Store("")
// 			chStop := make(chan bool, 1)
// 			go func() {
// 				for {
// 					result := gotResults.Load().(string)
// 					select {
// 					case c := <-ch:
// 						Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", ch, c.String()))
// 						if c.state {
// 							result += "1"
// 						} else {
// 							result += "0"
// 						}
// 						gotResults.Store(result)
// 						c.Done()
// 					case <-chStop:
// 						return
// 					}
// 				}
// 			}()
// 			defer func() { chStop <- true }()

// 			osc.WireUp(ch)
// 			osc.Oscillate(tc.oscHertz)

// 			// give the oscillator some time to....oscillate
// 			//time.Sleep(time.Second * 3)

// 			if !strings.HasPrefix(gotResults.Load().(string), tc.wantResults) {
// 				t.Errorf("Wanted results of at least %s but got %s", tc.wantResults, gotResults.Load().(string))
// 			}
// 		})
// 	}
// 	Debug(testName(t, ""), "End Test Cases Loop")
// }

func TestRSFlipFlop(t *testing.T) {
	testCases := []struct {
		rPinHasCharge bool
		sPinHasCharge bool
		wantQ         bool
		wantQBar      bool
	}{ // construction of the flipflop will start with a default of rPin:false, sPin:false, which causes false on both inputs of the S nor, which causes QBar on (Q off)
		{false, false, false, true}, // Un-Set should change nothing
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

	testNameDetail := func(i int) string {
		var priorR bool
		var priorS bool

		if i == 0 {
			priorR = false
			priorS = false
		} else {
			priorR = testCases[i-1].rPinHasCharge
			priorS = testCases[i-1].sPinHasCharge
		}

		return fmt.Sprintf("testCases[%d]: Switching from [rInHasCharge (%t), sInHasCharge (%t)] to [rInHasCharge (%t), sInHasCharge (%t)]", i, priorR, priorS, testCases[i].rPinHasCharge, testCases[i].sPinHasCharge)
	}

	Debug(testName(t, ""), "Initial Setup")

	var rPinChargeProvider, sPinChargeProvider *ChargeProvider
	rPinChargeProvider = NewChargeProvider(testName(t, "rChargeProvider"), false)
	sPinChargeProvider = NewChargeProvider(testName(t, "sChargeProvider"), false)

	// starting with no input signals (R and S are off)
	ff := NewRSFlipFlop(testName(t, "RSFlipFlop"), rPinChargeProvider, sPinChargeProvider)
	defer ff.Shutdown()

	var gotQ, gotQBar atomic.Value
	chQ := make(chan Charge, 1)
	chQBar := make(chan Charge, 1)
	chStopQ := make(chan bool, 1)
	chStopQBar := make(chan bool, 1)
	go func() {
		for {
			select {
			case cQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Charge {%s}", chQBar, cQBar.String()))
				gotQBar.Store(cQBar.state)
				cQBar.Done()
			case <-chStopQBar:
				return
			}
		}
	}()
	defer func() { chStopQBar <- true }()
	go func() {
		for {
			select {
			case cQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Charge {%s}", chQ, cQ.String()))
				gotQ.Store(cQ.state)
				cQ.Done()
			case <-chStopQ:
				return
			}
		}
	}()
	defer func() { chStopQ <- true }()

	ff.QBar.WireUp(chQBar)
	ff.Q.WireUp(chQ)

	if gotQ.Load().(bool) {
		t.Error("Wanted no charge at Q, but got charge")
	}

	if !gotQBar.Load().(bool) {
		t.Errorf("Wanted charge at QBar, but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	var summary string
	for i, tc := range testCases {
		summary = testNameDetail(i)
		t.Run(summary, func(t *testing.T) {

			Debug(testName(t, ""), summary)

			if tc.rPinHasCharge {
				rPinChargeProvider.Charge()
			} else {
				rPinChargeProvider.Discharge()
			}

			if tc.sPinHasCharge {
				sPinChargeProvider.Charge()
			} else {
				sPinChargeProvider.Discharge()
			}

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted charge of %t at Q, but got %t", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted charge of %t at QBar, but got %t", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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

	testNameDetail := func(i int) string {
		var priorClkIn bool
		var priorDataIn bool

		if i == 0 {
			// trues since starting with charged ChargeProviders when Newing the Latch initially
			priorClkIn = true
			priorDataIn = true
		} else {
			priorClkIn = testCases[i-1].clkIn
			priorDataIn = testCases[i-1].dataIn
		}

		return fmt.Sprintf("testCases[%d]: Switching from [clkIn (%t) dataIn (%t)] to [clkIn (%t) dataIn (%t)]", i, priorClkIn, priorDataIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	Debug(testName(t, ""), "Initial Setup")

	var clkChargeProvider, dataChargeProvider *ChargeProvider
	clkChargeProvider = NewChargeProvider(testName(t, "clkChargeProvider"), true)
	dataChargeProvider = NewChargeProvider(testName(t, "dataChargeProvider"), true)

	// starting with true input signals (Clk and Data are on)
	latch := NewLevelTriggeredDTypeLatch(testName(t, "LevelTriggeredDTypeLatch"), clkChargeProvider, dataChargeProvider)
	defer latch.Shutdown()

	chQ := make(chan Charge, 1)
	chQBar := make(chan Charge, 1)
	chStop := make(chan bool, 1)

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case cQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Charge {%s}", chQBar, cQBar.String()))
				gotQBar.Store(cQBar.state)
				cQBar.Done()
			case cQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Charge {%s}", chQ, cQ.String()))
				gotQ.Store(cQ.state)
				cQ.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	if !gotQ.Load().(bool) {
		t.Error("Wanted charge at Q, but got none")
	}

	if gotQBar.Load().(bool) {
		t.Error("Wanted no charge at QBar, but got charge")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	var summary string
	for i, tc := range testCases {
		summary = testNameDetail(i)
		t.Run(summary, func(t *testing.T) {

			Debug(testName(t, ""), summary)

			if tc.clkIn {
				clkChargeProvider.Charge()
			} else {
				clkChargeProvider.Discharge()
			}
			if tc.dataIn {
				dataChargeProvider.Charge()
			} else {
				dataChargeProvider.Discharge()
			}

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted charge of %t at Q, but got %t", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted charge of %t at QBar, but got %t", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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

	Debug(testName(t, ""), "Initial Setup")

	latchSwitches, _ := NewNSwitchBank(testName(t, "latchSwitches"), "00011000")
	defer latchSwitches.Shutdown()

	clkSwitch := NewSwitch(testName(t, "clkSwitch"), true)
	defer clkSwitch.Shutdown()

	latch := NewNBitLevelTriggeredDTypeLatch(testName(t, "NBitLevelTriggeredDTypeLatch"), clkSwitch, latchSwitches.Switches())
	defer latch.Shutdown()

	// build listen/transmit funcs to deal with each latch's Q
	gots := make([]atomic.Value, len(latch.Qs))
	var chStates []chan Charge
	var chStops []chan bool
	for i, q := range latch.Qs {

		chStates = append(chStates, make(chan Charge, 1))
		chStops = append(chStops, make(chan bool, 1))

		go func(chState chan Charge, chStop chan bool, index int) {
			for {
				select {
				case c := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Latches[%d]) Received on Channel (%v), Charge {%s}", index, chState, c.String()))
					gots[index].Store(c.state)
					c.Done()
				case <-chStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(Latches[%d]) Stopped", index))
					return
				}
			}
		}(chStates[i], chStops[i], i)

		q.WireUp(chStates[i])
	}
	defer func() {
		for i := 0; i < len(latch.Qs); i++ {
			chStops[i] <- true
		}
	}()

	priorWant := [8]bool{false, false, false, true, true, false, false, false}
	for i := 0; i < 8; i++ {
		if got := gots[i].Load().(bool); got != priorWant[i] {
			t.Errorf("Latches[%d] wanted (%t) but got (%t).\n", i, priorWant[i], got)
		}
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting switches to %s", i, tc.input), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting switches to %s", i, tc.input))

			// set to OFF to test that nothing will change in the latches store
			clkSwitch.Set(false)

			latchSwitches.SetSwitches(tc.input) // setting switches AFTER the clk goes to off to test that nothing actually would happen to the latches

			for i := range latch.Qs {
				if got := gots[i].Load().(bool); got != priorWant[i] {
					t.Errorf("Latches[%d], with clkSwitch off, wanted %t but got %t", i, priorWant[i], got)
				}
			}

			// Now set to ON to test that requested changes DID occur in the latches store
			clkSwitch.Set(true)

			for i := range latch.Qs {
				if got := gots[i].Load().(bool); got != tc.want[i] {
					t.Errorf("Latches[%d], with clkSwitch ON, wanted %t but got %t", i, tc.want[i], got)
				}
			}

			// now update the prior tracker bools to ensure next pass (with cklIn as OFF at the top) proves it didn't change (ie matches prior)
			for i := range latch.Qs {
				priorWant[i] = gots[i].Load().(bool)
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Adding (%s) to (%s)", i, tc.byte1, tc.byte2), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Adding (%s) to (%s)", i, tc.byte1, tc.byte2))

			addend1Switches, _ := NewNSwitchBank(testName(t, "addend1Switches"), tc.byte1)
			defer addend1Switches.Shutdown()

			addend2Switches, _ := NewNSwitchBank(testName(t, "addend2Switches"), tc.byte2)
			defer addend2Switches.Shutdown()

			sel, err := NewTwoToOneSelector(testName(t, "TwoToOneSelector"), nil, addend1Switches.Switches(), addend2Switches.Switches())

			if sel != nil {
				sel.Shutdown()
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
	Debug(testName(t, ""), "End Test Cases Loop")
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

	Debug(testName(t, ""), "Initial Setup")

	// start with these switches to verify uses A intially
	aInSwitches, _ := NewNSwitchBank(testName(t, "aInSwitches"), "111")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank(testName(t, "bInSwitches"), "000")
	defer bInSwitches.Shutdown()

	selectBSwitch := NewSwitch(testName(t, "selectBSwitch"), false)
	defer selectBSwitch.Shutdown()

	sel, _ := NewTwoToOneSelector(testName(t, "TwoToOneSelector"), selectBSwitch, aInSwitches.Switches(), bInSwitches.Switches())
	defer sel.Shutdown()

	// build listen/transmit funcs to deal with each of the selector's outputs
	gots := make([]atomic.Value, len(sel.Outs))
	var chStates []chan Charge
	var chStops []chan bool
	for i, o := range sel.Outs {

		chStates = append(chStates, make(chan Charge, 1))
		chStops = append(chStops, make(chan bool, 1))

		go func(chState chan Charge, chStop chan bool, index int) {
			for {
				select {
				case c := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("(SelectorOuts[%d]) Received on Channel (%v), Charge {%s}", index, chState, c.String()))
					gots[index].Store(c.state)
					c.Done()
				case <-chStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(SelectorOuts[%d]) Stopped", index))
					return
				}
			}
		}(chStates[i], chStops[i], i)

		o.WireUp(chStates[i])
	}
	defer func() {
		for i := 0; i < len(sel.Outs); i++ {
			chStops[i] <- true
		}
	}()

	want := true
	for i := 0; i < 3; i++ {
		if got := gots[i].Load().(bool); got != want {
			t.Errorf("Selector Output[%d]: A(111), B(000), use B?(false).  Wanted (%t) but got (%v).\n", i, want, got)
		}
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: A(%s), B(%s), use B?(%t)", i, tc.aIn, tc.bIn, tc.selectB), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: A(%s), B(%s), use B?(%t)", i, tc.aIn, tc.bIn, tc.selectB))

			aInSwitches.SetSwitches(tc.aIn)
			bInSwitches.SetSwitches(tc.bIn)
			selectBSwitch.Set(tc.selectB)

			for i := range sel.Outs {
				if got := gots[i].Load().(bool); got != tc.want[i] {
					t.Errorf("Selector Output[%d]: Wanted (%t) but got (%t).\n", i, tc.want[i], got)
				}
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestTwoToOneSelector_SelectingB_ASwitchesNoImpact(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")

	// start with off for A but on for B, but selecting A
	aInSwitches, _ := NewNSwitchBank(testName(t, "aInSwitches"), "000")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank(testName(t, "bInSwitches"), "111")
	defer bInSwitches.Shutdown()

	selectBSwitch := NewSwitch(testName(t, "selectBSwitch"), false)
	defer selectBSwitch.Shutdown()

	sel, _ := NewTwoToOneSelector(testName(t, "TwoToOneSelector"), selectBSwitch, aInSwitches.Switches(), bInSwitches.Switches())
	defer sel.Shutdown()

	// build listen/transmit funcs to deal with each of the selector's outputs
	gots := make([]atomic.Value, len(sel.Outs))
	var chStates []chan Charge
	var chStops []chan bool
	for i, o := range sel.Outs {

		chStates = append(chStates, make(chan Charge, 1))
		chStops = append(chStops, make(chan bool, 1))

		go func(chState chan Charge, chStop chan bool, index int) {
			for {
				select {
				case c := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("(SelectorOuts[%d]) Received on Channel (%v), Charge {%s}", index, chState, c.String()))
					gots[index].Store(c.state)
					c.Done()
				case <-chStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(SelectorOuts[%d]) Stopped", index))
					return
				}
			}
		}(chStates[i], chStops[i], i)

		o.WireUp(chStates[i])
	}
	defer func() {
		for i := 0; i < len(sel.Outs); i++ {
			chStops[i] <- true
		}
	}()

	Debug(testName(t, ""), "Start Test Cases")

	// starting with selecting A, get A's state

	for i := 0; i < 3; i++ {
		if gots[i].Load().(bool) {
			t.Error("Expecting false on all Outs of selector but got a true")
		}
	}

	selectBSwitch.Set(true)

	// selecting B, get B's state
	for i := 0; i < 3; i++ {
		if !gots[i].Load().(bool) {
			t.Error("Expecting true on all Outs of selector but got a false")
		}
	}

	aInSwitches.SetSwitches("101")

	// still selecting B, get B's state, regardless of A's state changing
	for i := 0; i < 3; i++ {
		if !gots[i].Load().(bool) {
			t.Error("Expecting true on all Outs of selector but got a false")
		}
	}
	Debug(testName(t, ""), "End Test Cases")
}

func TestThreeNumberAdder_MismatchInputs(t *testing.T) {
	wantError := "Mismatched input lengths. Addend1 len: 8, Addend2 len: 4"

	aInSwitches, _ := NewNSwitchBank(testName(t, "aInSwitches"), "00000000")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank(testName(t, "bInSwitches"), "0000")
	defer bInSwitches.Shutdown()

	addr, err := NewThreeNumberAdder(testName(t, "ThreeNumberAdder"), aInSwitches.Switches(), bInSwitches.Switches())

	if addr != nil {
		addr.Shutdown()
		t.Error("Did not expect an adder back but got one")
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

	Debug(testName(t, ""), "Initial Setup")

	aInSwitches, _ := NewNSwitchBank(testName(t, "aInSwitches"), "00000000")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank(testName(t, "bInSwitches"), "00000000")
	defer bInSwitches.Shutdown()

	addr, _ := NewThreeNumberAdder(testName(t, "ThreeNumberAdder"), aInSwitches.Switches(), bInSwitches.Switches())
	defer addr.Shutdown()

	// build listen/transmit funcs to deal with each of the adder's sum outputs
	var gotSums [8]atomic.Value
	var chSumStates []chan Charge
	var chSumStops []chan bool
	for i, s := range addr.Sums {

		chSumStates = append(chSumStates, make(chan Charge, 1))
		chSumStops = append(chSumStops, make(chan bool, 1))
		gotSums[i].Store(false)

		go func(chSumState chan Charge, chSumStop chan bool, index int) {
			for {
				select {
				case c := <-chSumState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Received on Channel (%v), Charge {%s}", index, chSumState, c.String()))
					gotSums[index].Store(c.state)
					c.Done()
				case <-chSumStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Stopped", index))
					return
				}
			}
		}(chSumStates[i], chSumStops[i], i)

		s.WireUp(chSumStates[i])
	}
	defer func() {
		for i := 0; i < len(addr.Sums); i++ {
			chSumStops[i] <- true
		}
	}()

	// build listen/transmit func to deal with the adder's carryout
	var gotCarryOut atomic.Value
	chCarryOutState := make(chan Charge, 1)
	chCarryOutStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-chCarryOutState:
				Debug(testName(t, "Select"), fmt.Sprintf("(CarryOut) Received on Channel (%v), Charge {%s}", chCarryOutState, c.String()))
				gotCarryOut.Store(c.state)
				c.Done()
			case <-chCarryOutStop:
				Debug(testName(t, "Select"), "(CarryOut) Stopped")
				return
			}
		}
	}()
	defer func() { chCarryOutStop <- true }()

	addr.CarryOut.WireUp(chCarryOutState)

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Adding (%s) to (%s)", i, tc.aIn, tc.bIn), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Adding (%s) to (%s)", i, tc.aIn, tc.bIn))

			aInSwitches.SetSwitches(tc.aIn)
			bInSwitches.SetSwitches(tc.bIn)

			if gotAnswer := getAnswerString(gotSums[:]); gotAnswer != tc.wantAnswer {
				t.Errorf("Wanted answer %s but got %s", tc.wantAnswer, gotAnswer)
			}

			if gotCarryOut.Load().(bool) != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, gotCarryOut.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestThreeNumberAdder_ThreeNumberAdd(t *testing.T) {

	aInSwitches, _ := NewNSwitchBank(testName(t, "aInSwitches"), "00000010")
	defer aInSwitches.Shutdown()

	bInSwitches, _ := NewNSwitchBank(testName(t, "bInSwitches"), "00000001")
	defer bInSwitches.Shutdown()

	addr, _ := NewThreeNumberAdder(testName(t, "ThreeNumberAdder"), aInSwitches.Switches(), bInSwitches.Switches())
	defer addr.Shutdown()

	// build listen/transmit funcs to deal with each of the adder's sum outputs
	var gotSums [8]atomic.Value
	var chSumStates []chan Charge
	var chSumStops []chan bool
	for i, s := range addr.Sums {

		chSumStates = append(chSumStates, make(chan Charge, 1))
		chSumStops = append(chSumStops, make(chan bool, 1))
		gotSums[i].Store(false)

		go func(chSumState chan Charge, chSumStop chan bool, index int) {
			for {
				select {
				case c := <-chSumState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Received on Channel (%v), Charge {%s}", index, chSumState, c.String()))
					gotSums[index].Store(c.state)
					c.Done()
				case <-chSumStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Stopped", index))
					return
				}
			}
		}(chSumStates[i], chSumStops[i], i)

		s.WireUp(chSumStates[i])
	}
	defer func() {
		for i := 0; i < len(addr.Sums); i++ {
			chSumStops[i] <- true
		}
	}()

	// build listen/transmit func to deal with the adder's carryout
	var gotCarryOut atomic.Value
	chCarryOutState := make(chan Charge, 1)
	chCarryOutStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-chCarryOutState:
				Debug(testName(t, "Select"), fmt.Sprintf("(CarryOut) Received on Channel (%v), Charge {%s}", chCarryOutState, c.String()))
				gotCarryOut.Store(c.state)
				c.Done()
			case <-chCarryOutStop:
				Debug(testName(t, "Select"), "(CarryOut) Stopped")
				return
			}
		}
	}()
	defer func() { chCarryOutStop <- true }()

	addr.CarryOut.WireUp(chCarryOutState)

	Debug(testName(t, ""), "Start Test Cases")

	wantAnswer := "00000011" // 2 + 1 = 3
	wantCarry := false

	if gotAnswer := getAnswerString(gotSums[:]); gotAnswer != wantAnswer {
		t.Errorf("Wanted answer %s but got %s", wantAnswer, gotAnswer)
	}

	if gotCarryOut.Load().(bool) != wantCarry {
		t.Errorf("Wanted carry %t, but got %t", wantCarry, gotCarryOut.Load().(bool))
	}

	addr.SaveToLatch.Set(true)
	addr.SaveToLatch.Set(false)

	aInSwitches.SetSwitches("00000011")

	addr.ReadFromLatch.Set(true)

	wantAnswer = "00000110" // 3 + 3 = 6
	wantCarry = false

	if gotAnswer := getAnswerString(gotSums[:]); gotAnswer != wantAnswer {
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
	Debug(testName(t, ""), "End Test Cases")
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

	Debug(testName(t, ""), "Initial Setup")

	testNameDetail := func(i int) string {
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

		return fmt.Sprintf("testCases[%d]: Switching from [clrIn (%t), clkIn (%t), dataIn (%t)] to [clrIn (%t), clkIn (%t), dataIn (%t)]", i, priorClrIn, priorClkIn, priorDataIn, testCases[i].clrIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	var clrChargeProvider, clkChargeProvider, dataChargeProvider *ChargeProvider
	clrChargeProvider = NewChargeProvider(testName(t, "clrChargeProvider"), false)
	clkChargeProvider = NewChargeProvider(testName(t, "clkChargeProvider"), true)
	dataChargeProvider = NewChargeProvider(testName(t, "dataChargeProvider"), true)

	latch := NewLevelTriggeredDTypeLatchWithClear(testName(t, "LevelTriggeredDTypeLatchWithClear"), clrChargeProvider, clkChargeProvider, dataChargeProvider)
	defer latch.Shutdown()

	chQ := make(chan Charge, 1)
	chQBar := make(chan Charge, 1)
	chStop := make(chan bool, 1)

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case cQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Charge {%s}", chQBar, cQBar.String()))
				gotQBar.Store(cQBar.state)
				cQBar.Done()
			case cQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Charge {%s}", chQ, cQ.String()))
				gotQ.Store(cQ.state)
				cQ.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	if !gotQ.Load().(bool) {
		t.Error("Wanted charge at Q, but got none")
	}

	if gotQBar.Load().(bool) {
		t.Error("Wanted no charge at QBar, but got charge")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(testNameDetail(i), func(t *testing.T) {

			Debug(testName(t, ""), testNameDetail(i))

			if tc.clrIn {
				clrChargeProvider.Charge()
			} else {
				clrChargeProvider.Discharge()
			}
			if tc.clkIn {
				clkChargeProvider.Charge()
			} else {
				clkChargeProvider.Discharge()
			}
			if tc.dataIn {
				dataChargeProvider.Charge()
			} else {
				dataChargeProvider.Discharge()
			}

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted charge of %t at Q, but got %t", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted charge of %t at QBar, but got %t", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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

	Debug(testName(t, ""), "Initial Setup")

	dataSwitches, _ := NewNSwitchBank(testName(t, "dataSwitches"), "00000000")
	defer dataSwitches.Shutdown()

	clrSwitch := NewSwitch(testName(t, "clrSwitch"), false)
	defer clrSwitch.Shutdown()

	clkSwitch := NewSwitch(testName(t, "clkSwitch"), false)
	defer clkSwitch.Shutdown()

	latch := NewNBitLevelTriggeredDTypeLatchWithClear(testName(t, "NBitLevelTriggeredDTypeLatchWithClear"), clrSwitch, clkSwitch, dataSwitches.Switches())
	defer latch.Shutdown()

	// build listen/transmit funcs to deal with each latch's Q
	gots := make([]atomic.Value, len(latch.Qs))
	var chStates []chan Charge
	var chStops []chan bool
	for i, q := range latch.Qs {

		chStates = append(chStates, make(chan Charge, 1))
		chStops = append(chStops, make(chan bool, 1))

		go func(chState chan Charge, chStop chan bool, index int) {
			for {
				select {
				case c := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Latches[%d]) Received on Channel (%v), Charge {%s}", index, chState, c.String()))
					gots[index].Store(c.state)
					c.Done()
				case <-chStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(Latches[%d]) Stopped", index))
					return
				}
			}
		}(chStates[i], chStops[i], i)

		q.WireUp(chStates[i])
	}
	defer func() {
		for i := 0; i < len(latch.Qs); i++ {
			chStops[i] <- true
		}
	}()

	priorWant := [8]bool{false, false, false, false, false, false, false, false}
	for i := 0; i < 8; i++ {
		if got := gots[i].Load().(bool); got != priorWant[i] {
			t.Errorf("Latch[%d] wanted (%t) but got (%t).\n", i, priorWant[i], got)
		}
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting switches to %s", i, tc.input), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting switches to %s", i, tc.input))

			// reset the Clear from the prior run (asserts at the bottom of this loop)
			clrSwitch.Set(false)

			// set to OFF to test that nothing will change in the latches store
			clkSwitch.Set(false)
			dataSwitches.SetSwitches(tc.input)

			for i := 0; i < 8; i++ {
				if got := gots[i].Load().(bool); got != priorWant[i] {
					t.Errorf("Latch[%d] wanted (%t) but got (%t).\n", i, priorWant[i], got)
				}
			}

			// Now set to ON to test that requested changes did occur in the latches store
			clkSwitch.Set(true)

			for i := range latch.Qs {
				if got := gots[i].Load().(bool); got != tc.want[i] {
					t.Errorf("Latch[%d], with clkSwitch ON, wanted %t but got %t", i, tc.want[i], got)
				}
			}

			// now update the prior tracker bools to ensure next pass (with cklIn as OFF at the top of the loop) proves it didn't change (aka matches prior)
			for i := range latch.Qs {
				priorWant[i] = gots[i].Load().(bool)
			}

			// Now Clear the latches
			clrSwitch.Set(true)

			// clear should have set all Qs to off
			want := false
			for i := range latch.Qs {
				if got := gots[i].Load().(bool); got != want {
					t.Errorf("Latch[%d], with clrSwitch ON, wanted %t but got %t", i, want, got)
				}
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

// // TODO: TestNNumberAdder creates an adder loop that has no bounds so it is expected to stack overlow
// func TestNNumberAdder(t *testing.T) {

// 	Debug(testName(t, ""), "Initial Setup")

// 	switches, _ := NewNSwitchBank(testName(t, "switches"), "1")
// 	defer switches.Shutdown()

// 	addr, _ := NewNNumberAdder(testName(t, "NNumberAdder"), switches.Switches())
// 	defer addr.Shutdown()

// 	// build listen/transmit funcs to deal with each of the NNumberAdder's Q outputs (the latest inner-adder's Sum sent through)
// 	var gotSums [1]atomic.Value
// 	var chSumStates []chan Charge
// 	var chSumStops []chan bool
// 	for i, s := range addr.Sums {

// 		chSumStates = append(chSumStates, make(chan Charge, 1))
// 		chSumStops = append(chSumStops, make(chan bool, 1))
// 		gotSums[i].Store(false)

// 		go func(chSumState chan Charge, chSumStop chan bool, index int) {
// 			for {
// 				select {
// 				case c := <-chSumState:
// 					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Received on Channel (%v), Charge {%s}", index, chSumState, c.String()))
// 					gotSums[index].Store(c.state)
// 					c.Done()
// 				case <-chSumStop:
// 					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Stopped", index))
// 					return
// 				}
// 			}
// 		}(chSumStates[i], chSumStops[i], i)

// 		s.WireUp(chSumStates[i])
// 	}
// 	defer func() {
// 		for i := 0; i < len(addr.Sums); i++ {
// 			chSumStops[i] <- true
// 		}
// 	}()

// 	// build listen/transmit funcs to deal with each of the adder's INTERNAL adder's sum outputs
// 	var gotInnerSums [1]atomic.Value
// 	var chInnerSumStates []chan Charge
// 	var chInnerSumStops []chan bool
// 	for i, s := range addr.adder.Sums {

// 		chInnerSumStates = append(chInnerSumStates, make(chan Charge, 1))
// 		chInnerSumStops = append(chInnerSumStops, make(chan bool, 1))
// 		gotInnerSums[i].Store(false)

// 		go func(chInnerSumState chan Charge, chInnerSumStop chan bool, index int) {
// 			for {
// 				select {
// 				case c := <-chInnerSumState:
// 					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Received on Channel (%v), Charge {%s}", index, chInnerSumState, c.String()))
// 					gotInnerSums[index].Store(c.state)
// 					c.Done()
// 				case <-chInnerSumStop:
// 					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Stopped", index))
// 					return
// 				}
// 			}
// 		}(chInnerSumStates[i], chInnerSumStops[i], i)

// 		s.WireUp(chInnerSumStates[i])
// 	}
// 	defer func() {
// 		for i := 0; i < len(addr.adder.Sums); i++ {
// 			chInnerSumStops[i] <- true
// 		}
// 	}()

// 	Debug(testName(t, ""), "Start Test Cases")

// 	// totally putting fake setup here to test some things.  this test needs to be reworked to trap the expected states once I get it to work
// 	// totally putting fake setup here to test some things.  this test needs to be reworked to trap the expected states once I get it to work
// 	// totally putting fake setup here to test some things.  this test needs to be reworked to trap the expected states once I get it to work

// 	// can't really prove that Clear clears anything since the internal Add switch defaults to false so the latch doesn't get the adders answer in the first place
// 	addr.Clear.Set(true)
// 	addr.Clear.Set(false)

// 	// regardless of sending non-zero into the NBitAdder's input switches, the fact the Add switch isn't on would prevent any 1s from getting into the latch, so expect no 1s
// 	want := "0"

// 	if got := getAnswerString(gotSums[:]); got != want {
// 		t.Errorf("[Initial setup] Wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
// 	}

// 	// however, the internal adder aspect of the NBitAdder should have state based on the initial switches
// 	want = "1"
// 	if got := getAnswerString(gotInnerSums[:]); got != want {
// 		t.Errorf("[Initial setup] Wanted answer of NNumberAdder's inner-adder to be %s but got %s", want, got)
// 	}

// 	// this deadlocks once true.  booooooooo
// 	addr.Add.Set(true)
// 	//addr.Add.Set(false)

// 	// !!! would want 1 here if the Add Set worked, yes?  allowing the latch to get fed the latest inner-adder's state
// 	want = "0"
// 	if got := getAnswerString(gotSums[:]); got != want {
// 		t.Errorf("After an add, wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
// 	}

// 	switches.SetSwitches("0")

// 	// the internal adder aspect of the NBitAdder should have state based on the switches each time
// 	want = "0"
// 	if got := getAnswerString(gotInnerSums[:]); got != want {
// 		t.Errorf("[Initial setup] Wanted answer of NNumberAdder's inner-adder to be %s but got %s", want, got)
// 	}

// 	// want = "00000011"
// 	// if got := getAnswerString(gotSums[:]); got != want {
// 	// 	t.Errorf("After another add, wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
// 	// }

// 	Debug(testName(t, ""), "End Test Cases")
// }

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

	testNameDetail := func(i int) string {
		var priorClkIn bool
		var priorDataIn bool

		if i == 0 {
			priorClkIn = false
			priorDataIn = false
		} else {
			priorClkIn = testCases[i-1].clkIn
			priorDataIn = testCases[i-1].dataIn
		}

		return fmt.Sprintf("testCases[%d]: Switching from [clkIn (%t) dataIn (%t)] to [clkIn (%t) dataIn (%t)]", i, priorClkIn, priorDataIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	Debug(testName(t, ""), "Initial Setup")

	var clkChargeProvider, dataChargeProvider *ChargeProvider
	clkChargeProvider = NewChargeProvider(testName(t, "clkChargeProvider"), false)
	dataChargeProvider = NewChargeProvider(testName(t, "dataChargeProvider"), false)

	latch := NewEdgeTriggeredDTypeLatch(testName(t, "EdgeTriggeredDTypeLatch"), clkChargeProvider, dataChargeProvider)
	defer latch.Shutdown()

	chQ := make(chan Charge, 1)
	chQBar := make(chan Charge, 1)
	chStop := make(chan bool, 1)

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case cQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Charge {%s}", chQBar, cQBar.String()))
				gotQBar.Store(cQBar.state)
				cQBar.Done()
			case cQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Charge {%s}", chQ, cQ.String()))
				gotQ.Store(cQ.state)
				cQ.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	if gotQ.Load().(bool) {
		t.Error("Wanted no charge at Q, but got charge")
	}

	if !gotQBar.Load().(bool) {
		t.Error("Wanted charge at QBar, but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	var summary string
	for i, tc := range testCases {
		summary = testNameDetail(i)
		t.Run(summary, func(t *testing.T) {

			Debug(testName(t, ""), summary)

			if tc.dataIn {
				dataChargeProvider.Charge()
			} else {
				dataChargeProvider.Discharge()
			}
			if tc.clkIn {
				clkChargeProvider.Charge()
			} else {
				clkChargeProvider.Discharge()
			}

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted charge of %t at Q, but got %t", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted charge of %t at QBar, but got %t", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestFrequencyDivider(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")

	osc := NewOscillator(testName(t, "Oscillator"), false)
	freqDiv := NewFrequencyDivider(testName(t, "FrequencyDivider"), osc)
	defer freqDiv.Shutdown()

	var gotOscResults atomic.Value
	chOsc := make(chan Charge, 1)
	gotOscResults.Store("")
	chStopOsc := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-chOsc:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", chOsc, c.String()))
				result := gotOscResults.Load().(string)
				if c.state {
					result += "1"
				} else {
					result += "0"
				}
				gotOscResults.Store(result)
				c.Done()
			case <-chStopOsc:
				return
			}
		}
	}()
	defer func() { chStopOsc <- true }()

	osc.WireUp(chOsc)

	var gotDivResults atomic.Value
	chDiv := make(chan Charge, 1)
	gotDivResults.Store("")
	chStopDiv := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-chDiv:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", chDiv, c.String()))
				result := gotDivResults.Load().(string)
				if c.state {
					result += "1"
				} else {
					result += "0"
				}
				gotDivResults.Store(result)
				c.Done()
			case <-chStopDiv:
				return
			}
		}
	}()
	defer func() { chStopDiv <- true }()

	freqDiv.Q.WireUp(chDiv)

	wantOsc := "0"
	if !strings.HasPrefix(gotOscResults.Load().(string), wantOsc) {
		t.Errorf("Wanted oscillator results %s but got %s", wantOsc, gotOscResults.Load().(string))
	}

	wantDiv := "0"
	if !strings.HasPrefix(gotDivResults.Load().(string), wantDiv) {
		t.Errorf("Wanted divider results %s but got %s", wantDiv, gotDivResults.Load().(string))
	}

	Debug(testName(t, ""), "Start Test Case")

	osc.Oscillate(2) // 2 times a second

	time.Sleep(time.Second * 5) // for 5 seconds, should give me at least 8 oscillations and 4 divider  (though a tad nondeterministic and could fail)

	osc.Stop()

	wantOsc = "01010101"
	if !strings.HasPrefix(gotOscResults.Load().(string), wantOsc) {
		t.Errorf("Wanted oscillator results of at least %s but got %s", wantOsc, gotOscResults.Load().(string))
	}

	wantDiv = "0101"
	if !strings.HasPrefix(gotDivResults.Load().(string), wantDiv) {
		t.Errorf("Wanted divider results %s but got %s", wantDiv, gotDivResults.Load().(string))
	}

	Debug(testName(t, ""), "End Test Case")
}

// TODO: Troubles with this beast
func TestNBitRippleCounter_EightBit(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")

	osc := NewOscillator(testName(t, "Oscillator"), false)
	ripple := NewNBitRippleCounter(testName(t, "NBitRippleCounter"), osc, 8)
	defer ripple.Shutdown()

	var gotOscResults atomic.Value
	gotOscResults.Store("")

	// build listen/transmit func to deal with the oscillator
	chOsc := make(chan Charge, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case c := <-chOsc:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Charge {%s}", chOsc, c.String()))
				result := gotOscResults.Load().(string)
				if c.state {
					result += "1"
				} else {
					result += "0"
				}
				gotOscResults.Store(result)
				c.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	osc.WireUp(chOsc)

	// build listen/transmit funcs to deal with each of the RippleCounter's Q outputs
	var gotRippleQs [8]atomic.Value
	var gotDivResults [8]atomic.Value
	var chRippleStates []chan Charge
	var chRippleStops []chan bool

	for i, q := range ripple.Qs {

		chRippleStates = append(chRippleStates, make(chan Charge, 1))
		chRippleStops = append(chRippleStops, make(chan bool, 1))
		gotRippleQs[i].Store(false)
		gotDivResults[i].Store("")

		go func(chRippleState chan Charge, chRippleStop chan bool, index int) {
			for {
				select {
				case c := <-chRippleState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Qs[%d]) Received on Channel (%v), Charge {%s}", index, chRippleState, c.String()))
					result := gotDivResults[index].Load().(string)
					if c.state {
						result += "1"
					} else {
						result += "0"
					}
					gotDivResults[index].Store(result)
					gotRippleQs[i].Store(c.state)
					Debug(testName(t, "Latest Answer"), fmt.Sprintf("(Qs[%d]) Caused {%s}", index, getAnswerString(gotRippleQs[:])))
					c.Done()
				case <-chRippleStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(Qs[%d]) Stopped", index))
					return
				}
			}
		}(chRippleStates[i], chRippleStops[i], i)

		q.WireUp(chRippleStates[i])
	}
	defer func() {
		for i := 0; i < len(ripple.Qs); i++ {
			chRippleStops[i] <- true
		}
	}()

	wantOsc := "0"
	if !strings.HasPrefix(gotOscResults.Load().(string), wantOsc) {
		t.Errorf("Wanted oscillator results %s but got %s", wantOsc, gotOscResults.Load().(string))
	}

	wantCounterAnswer := "10101010"
	if gotAnswer := getAnswerString(gotRippleQs[:]); gotAnswer != wantCounterAnswer {
		t.Errorf("Wanted counter results %s but got %s", wantCounterAnswer, gotAnswer)
	}

	Debug(testName(t, ""), "Start Test Case")

	osc.Oscillate(2) // 2 times a second

	time.Sleep(time.Second * 4) // for 4 seconds, should give me 8 oscillations and something or other final answer from all the counter's Q states

	osc.Stop()

	wantOsc = "01010101"
	if !strings.HasPrefix(gotOscResults.Load().(string), wantOsc) {
		t.Errorf("Wanted oscillator results of at least %s but got %s", wantOsc, gotOscResults.Load().(string))
	}

	wantCounterAnswer = "10101110"
	if gotAnswer := getAnswerString(gotRippleQs[:]); gotAnswer != wantCounterAnswer {
		t.Errorf("Wanted counter results %s but got %s", wantCounterAnswer, gotAnswer)
	}

	for i := 0; i < len(gotDivResults); i++ {
		Debug(testName(t, "Inner-Div-Results"), fmt.Sprintf("(Divs[%d]) %s", i, gotDivResults[i].Load().(string)))

	}
	Debug(testName(t, ""), "End Test Case")
}

// pre clr d clk   q  !q
//
//	1   0  X  X    1  0   preset makes data and clock not matter, forces Q
//	0   1  X  X    0  1   clear makes data and clock not matter, forces QBar
//	0   0  1  ^    1  0   should take the data value since clock was raised (transitioned to 1)
//	0   0  0  ^    0  1	  should take the data value since clock was raised (transitioned to 1)
//	0   0  X  0    q  !q  data doesn't matter, no clock raised (transition to 1) to trigger a store-it action

// TODO: Not sure why this isn't working yet.  Clear should have set Q false, QBar true.  :(
func TestEdgeTriggeredDTypeLatchWithPresetAndClear(t *testing.T) {
	testCases := []struct {
		presetIn bool
		clearIn  bool
		dataIn   bool
		clkIn    bool
		wantQ    bool
		wantQBar bool
	}{ // construction of the latches will start with a default of presetIn:true, clearIn:false, dataIn:false, clkIn:false, which causes Q on (QBar off)
		{false, true, false, false, false, true}, // clear makes data and clock not matter, forces QBar
		// {false, false, true, true, true, false},   // should take the data value since clock was raised (transitioned to 1)
		// {false, false, false, false, true, false}, // Q/QBar should not change regardless of data false since lowered clock
		// {false, false, false, true, false, true},  // should take the data value since clock was raised (transitioned to 1)
		// {false, false, true, true, true, false},   // should take the data value since clock was raised (transitioned to 1)
	}

	testNameDetail := func(i int) string {
		var priorPresetIn bool
		var priorClearIn bool
		var priorClkIn bool
		var priorDataIn bool

		if i == 0 {
			// starting with some defaults, where preset should make Q charged
			priorPresetIn = true
			priorClearIn = false
			priorClkIn = false
			priorDataIn = false
		} else {
			priorPresetIn = testCases[i-1].presetIn
			priorClearIn = testCases[i-1].clearIn
			priorClkIn = testCases[i-1].clkIn
			priorDataIn = testCases[i-1].dataIn
		}

		return fmt.Sprintf("testCases[%d]: Switching from [presetIn (%t) clearIn (%t) clkIn (%t) dataIn (%t)] to [presetIn (%t) clearIn (%t) clkIn (%t) dataIn (%t)]", i, priorPresetIn, priorClearIn, priorClkIn, priorDataIn, testCases[i].presetIn, testCases[i].clearIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	Debug(testName(t, ""), "Initial Setup")

	var presetChargeProvider, clearChargeProvider, clkChargeProvider, dataChargeProvider *ChargeProvider
	presetChargeProvider = NewChargeProvider(testName(t, "presetChargeProvider"), true)
	clearChargeProvider = NewChargeProvider(testName(t, "clearChargeProvider"), false)
	clkChargeProvider = NewChargeProvider(testName(t, "clkChargeProvider"), false)
	dataChargeProvider = NewChargeProvider(testName(t, "dataChargeProvider"), false)

	latch := NewEdgeTriggeredDTypeLatchWithPresetAndClear(testName(t, "EdgeTriggeredDTypeLatchWithPresetAndClear"), presetChargeProvider, clearChargeProvider, clkChargeProvider, dataChargeProvider)
	defer latch.Shutdown()

	chQ := make(chan Charge, 1)
	chQBar := make(chan Charge, 1)
	chStop := make(chan bool, 1)

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case cQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Charge {%s}", chQBar, cQBar.String()))
				gotQBar.Store(cQBar.state)
				cQBar.Done()
			case cQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Charge {%s}", chQ, cQ.String()))
				gotQ.Store(cQ.state)
				cQ.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	if !gotQ.Load().(bool) {
		t.Error("Wanted charge at Q but got none")
	}

	if gotQBar.Load().(bool) {
		t.Error("Wanted no charge at QBar but got charge")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	var summary string
	for i, tc := range testCases {
		summary = testNameDetail(i)
		t.Run(summary, func(t *testing.T) {

			Debug(testName(t, ""), summary)

			if tc.presetIn {
				presetChargeProvider.Charge()
			} else {
				presetChargeProvider.Discharge()
			}
			if tc.clearIn {
				clearChargeProvider.Charge()
			} else {
				clearChargeProvider.Discharge()
			}
			if tc.dataIn {
				dataChargeProvider.Charge()
			} else {
				dataChargeProvider.Discharge()
			}
			if tc.clkIn {
				clkChargeProvider.Charge()
			} else {
				clkChargeProvider.Discharge()
			}

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted charge of %t at Q, but got %t", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted charge of %t at QBar, but got %t", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}
