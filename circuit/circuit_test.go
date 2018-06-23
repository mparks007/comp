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
// go test -run TestRelay_WithBatteries -count 50 -trace out2.txt (go tool trace out2.txt)
// go test -race -cpu=1 -run TestFullAdder -count 5 -trace TestFullAdder_trace.txt > TestFullAdder_run.txt
// go test -debug (my own flag to write all the debug to the console during test run)

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
func TestPwrsource(t *testing.T) {
	pwr := &pwrSource{}
	pwr.Init()
	pwr.Name = testName(t, "pwrSource")

	var want bool
	var got1, got2 atomic.Value
	ch1 := make(chan Electron, 1)
	ch2 := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e1 := <-ch1:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch1) Received on Channel (%v), Electron {%s}", ch1, e1.String()))
				got1.Store(e1.powerState)
				e1.Done()
			case e2 := <-ch2:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch2) Received on Channel (%v), Electron {%s}", ch2, e2.String()))
				got2.Store(e2.powerState)
				e2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	// two wire ups to prove both will get called
	pwr.WireUp(ch1)
	pwr.WireUp(ch2)

	// test default state (unpowered)
	want = false

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test power transmit
	want = true
	pwr.Transmit(Electron{powerState: want})

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test transmit loss of power
	want = false
	pwr.Transmit(Electron{powerState: want})

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test transmitting same state as last time (should skip it)
	pwr.Transmit(Electron{powerState: want})

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did.")
	default:
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did.")
	default:
	}
}

func TestWire_NoDelay(t *testing.T) {
	wire := NewWire(testName(t, "Wire"), 0)
	defer wire.Shutdown()

	var want bool
	var got1, got2 atomic.Value
	ch1 := make(chan Electron, 1)
	ch2 := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e1 := <-ch1:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch1) Received on Channel (%v), Electron {%s}", ch1, e1.String()))
				got1.Store(e1.powerState)
				e1.Done()
			case e2 := <-ch2:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch2) Received on Channel (%v), Electron {%s}", ch2, e2.String()))
				got2.Store(e2.powerState)
				e2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	// two wire ups to prove both will get called
	wire.WireUp(ch1)
	wire.WireUp(ch2)

	// test default state (unpowered)
	want = false

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test power transmit
	want = true
	wire.Transmit(Electron{powerState: want})

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test transmit loss of power
	want = false
	wire.Transmit(Electron{powerState: want})

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test transmitting same state as last time (should skip it)
	wire.Transmit(Electron{powerState: want})

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did.")
	default:
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did.")
	default:
	}
}

func TestWire_WithDelay(t *testing.T) {
	var wireLen uint = 10

	wire := NewWire(testName(t, "Wire"), wireLen)
	defer wire.Shutdown()

	var want bool
	var got1, got2 atomic.Value
	ch1 := make(chan Electron, 1)
	ch2 := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e1 := <-ch1:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch1) Received on Channel (%v), Electron {%s}", ch1, e1.String()))
				got1.Store(e1.powerState)
				e1.Done()
			case e2 := <-ch2:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch2) Received on Channel (%v), Electron {%s}", ch2, e2.String()))
				got2.Store(e2.powerState)
				e2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	// two wire ups to prove both will get called
	wire.WireUp(ch1)
	wire.WireUp(ch2)

	// test default state (unpowered)
	want = false

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// test power transmit
	want = true

	start := time.Now()
	wire.Transmit(Electron{powerState: want})
	end := time.Now()

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
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
	wire.Transmit(Electron{powerState: want})
	end = time.Now()

	if got1.Load().(bool) != want {
		t.Errorf("Expected channel 1 to be %t but got %t", want, got1.Load().(bool))
	}
	if got2.Load().(bool) != want {
		t.Errorf("Expected channel 2 to be %t but got %t", want, got2.Load().(bool))
	}

	// validate wire delay
	gotDuration = end.Sub(start) // + time.Millisecond*1 // adding in just a little more to avoid timing edge case
	wantDuration = time.Millisecond * time.Duration(wireLen)
	if gotDuration < wantDuration {
		t.Errorf("Wire power off transmit time should have been %v but was %v", wantDuration, gotDuration)
	}

	// test transmitting same state as last time (should skip it)
	wire.Transmit(Electron{powerState: want})

	select {
	case <-ch1:
		t.Error("Transmit of same state as prior state should have never gotten to ch1, but it did.")
	default:
	}

	select {
	case <-ch2:
		t.Error("Transmit of same state as prior state should have never gotten to ch2, but it did.")
	default:
	}
}

func TestRibbonCable(t *testing.T) {
	rib := NewRibbonCable(testName(t, "RibbonCable"), 2, 0)
	defer rib.Shutdown()

	rib.SetInputs(NewBattery(testName(t, "Battery1"), false), NewBattery(testName(t, "Battery2"), true))

	var want bool
	var got1, got2 atomic.Value
	ch1 := make(chan Electron, 1)
	ch2 := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e1 := <-ch1:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch1) Received on Channel (%v), Electron {%s}", ch1, e1.String()))
				got1.Store(e1.powerState)
				e1.Done()
			case e2 := <-ch2:
				Debug(testName(t, "Select"), fmt.Sprintf("(ch2) Received on Channel (%v), Electron {%s}", ch2, e2.String()))
				got2.Store(e2.powerState)
				e2.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	rib.Wires[0].(*Wire).WireUp(ch1)
	rib.Wires[1].(*Wire).WireUp(ch2)

	// the first wire in the ribbon cable had a dead battery
	want = false

	if got1.Load().(bool) != want {
		t.Errorf("Left Switch off, wanted the wire to see power as %t but got %t", want, got1.Load().(bool))
	}

	// the first wire in the ribbon cable had a live battery
	want = true

	if got2.Load().(bool) != want {
		t.Errorf("Right Switch on, wanted the wire to see power as %t but got %t", want, got2.Load().(bool))
	}
}

func TestBattery(t *testing.T) {
	bat := NewBattery(testName(t, "Battery"), true)

	var want bool
	var got atomic.Value
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	bat.WireUp(ch)

	// test default battery state (powered)
	want = true

	if got.Load().(bool) != want {
		t.Errorf("With a new battery, wanted the subscriber to see power as %t but got %t", want, got.Load().(bool))
	}

	// test loss of power
	bat.Discharge()
	want = false

	if got.Load().(bool) != want {
		t.Errorf("With a discharged battery, wanted the subscriber's IsPowered to be %t but got %t", want, got.Load().(bool))
	}

	// test re-added power
	bat.Charge()
	want = true

	if got.Load().(bool) != want {
		t.Errorf("With a charged battery, wanted the subscriber's IsPowered to be %t but got %t", want, got.Load().(bool))
	}

	// test charging again (should skip it)
	bat.Charge()

	select {
	case <-ch:
		t.Error("Transmit of same state as prior state should have never gotten to ch, but it did.")
	default:
	}
}

func TestRelay_WithBatteries(t *testing.T) {
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
		{true, true, false, true}, // final test ensuring we can toggle all inputs fully reversed again
	}

	Debug(testName(t, ""), "Initial Setup")

	var pin1Battery, pin2Battery *Battery
	pin1Battery = NewBattery(testName(t, "Battery1"), true)
	pin2Battery = NewBattery(testName(t, "Battery2"), true)

	rel := NewRelay(testName(t, "Relay"), pin1Battery, pin2Battery)
	defer rel.Shutdown()

	var gotOpenOut, gotClosedOut atomic.Value
	chOpen := make(chan Electron, 1)
	chClosed := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case eOpen := <-chOpen:
				Debug(testName(t, "Select"), fmt.Sprintf("(chOpen) Received on Channel (%v), Electron {%s}", chOpen, eOpen.String()))
				gotOpenOut.Store(eOpen.powerState)
				eOpen.Done()
			case eClosed := <-chClosed:
				Debug(testName(t, "Select"), fmt.Sprintf("(chClosed) Received on Channel (%v), Electron {%s}", chClosed, eClosed.String()))
				gotClosedOut.Store(eClosed.powerState)
				eClosed.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	rel.OpenOut.WireUp(chOpen)
	rel.ClosedOut.WireUp(chClosed)

	if gotOpenOut.Load().(bool) != false {
		t.Error("Wanted no power at the open position but got some")
	}
	if gotClosedOut.Load().(bool) != true {
		t.Error("Wanted power at the closed position but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting input A to (%t) and B to (%t)", i, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting input A to (%t) and B to (%t)", i, tc.aInPowered, tc.bInPowered))

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

			if gotOpenOut.Load().(bool) != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut.Load().(bool))
			}

			if gotClosedOut.Load().(bool) != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut.Load().(bool))
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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
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
		t.Errorf("With an off switch, wanted the subscriber to see power as %t but got %t", want, got.Load().(bool))
	}

	Debug(testName(t, ""), "Start Test Cases")

	// initial turn on
	want = true
	sw.Set(want)

	if got.Load().(bool) != want {
		t.Errorf("With an off switch turned on, wanted the subscriber to see power as %t but got %t", want, got.Load().(bool))
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
		t.Errorf("With an on switch turned off, wanted the subscriber to see power as %t but got %t", want, got.Load().(bool))
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
		{"", "Input not in binary format: "},
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting switches to (%s)", i, tc.input), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting switches to (%s)", i, tc.input))

			sb, err := NewNSwitchBank(testName(t, "NSwitchBank"), tc.input)

			if sb != nil {
				sb.Shutdown()
				t.Error("Didn't expected a Switch Bank back but got one.")
			}

			tc.wantError += "\"" + tc.input + "\""

			if err == nil || (err != nil && err.Error() != tc.wantError) {
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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
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
			for i, pwr := range sb.Switches() {

				pwr.(*Switch).WireUp(ch)

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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
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
			}

			sb, err := NewNSwitchBank(testName(t, "NSwitchBank"), string(randomInput))
			sb.SetSwitches(tc.input)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
			}
			defer sb.Shutdown()

			Debug(testName(t, ""), "Start Test Cases WireUp Per Switch")

			// will just check one switch at a time vs. trying to get some full answer in one go from the bank
			for i, pwr := range sb.Switches() {

				pwr.(*Switch).WireUp(ch)

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
	chOpen := make(chan Electron, 1)
	chClosed := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case eOpen := <-chOpen:
				Debug(testName(t, "Select"), fmt.Sprintf("(chOpen) Received on Channel (%v), Electron {%s}", chOpen, eOpen.String()))
				gotOpenOut.Store(eOpen.powerState)
				eOpen.Done()
			case eClosed := <-chClosed:
				Debug(testName(t, "Select"), fmt.Sprintf("(chClosed) Received on Channel (%v), Electron {%s}", chClosed, eClosed.String()))
				gotClosedOut.Store(eClosed.powerState)
				eClosed.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	rel.OpenOut.WireUp(chOpen)
	rel.ClosedOut.WireUp(chClosed)

	if gotOpenOut.Load().(bool) != false {
		t.Error("Wanted no power at the open position but got some")
	}
	if gotClosedOut.Load().(bool) != true {
		t.Error("Wanted power at the closed position but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t)", i, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t)", i, tc.aInPowered, tc.bInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			if gotOpenOut.Load().(bool) != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, gotOpenOut.Load().(bool))
			}

			if gotClosedOut.Load().(bool) != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, gotClosedOut.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if got.Load().(bool) != true {
		t.Error("Wanted power on the gate but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t) and C power to (%t)", i, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t) and C power to (%t)", i, tc.aInPowered, tc.bInPowered, tc.cInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if got.Load().(bool) != true {
		t.Error("Wanted power on the gate but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t) and C power to (%t)", i, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t) and C power to (%t)", i, tc.aInPowered, tc.bInPowered, tc.cInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if got.Load().(bool) != true {
		t.Error("Wanted power on the gate but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t) and C power to (%t)", i, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t) and C power to (%t)", i, tc.aInPowered, tc.bInPowered, tc.cInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if got.Load().(bool) != false {
		t.Error("Wanted no power on the gate but got some")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t) and C power to (%t)", i, tc.aInPowered, tc.bInPowered, tc.cInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t) and C power to (%t)", i, tc.aInPowered, tc.bInPowered, tc.cInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.cInPowered)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if got.Load().(bool) != false {
		t.Error("Wanted no power on the gate but got some")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t)", i, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t)", i, tc.aInPowered, tc.bInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestInverter(t *testing.T) {
	testCases := []struct {
		inPowered bool
		wantOut   bool
	}{
		{false, true},
		{true, false},
		{false, true},
		{true, false},
	}

	Debug(testName(t, ""), "Initial Setup")

	pin1Battery := NewBattery(testName(t, "Battery"), true)
	inv := NewInverter(testName(t, "Inverter"), pin1Battery)
	defer inv.Shutdown()

	var got atomic.Value
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	inv.WireUp(ch)

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Input as (%t)", i, tc.inPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Input as (%t)", i, tc.inPowered))

			if tc.inPowered {
				pin1Battery.Charge()
			} else {
				pin1Battery.Discharge()
			}

			if got.Load().(bool) != tc.wantOut {
				t.Errorf("Input power was %t so wanted it inverted to %t but got %t", tc.inPowered, tc.wantOut, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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
	ch := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-ch:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
				got.Store(e.powerState)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	gate.WireUp(ch)

	if got.Load().(bool) != true {
		t.Error("Wanted power on the gate but got none")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t)", i, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting A power to (%t) and B power to (%t)", i, tc.aInPowered, tc.bInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

			if got.Load().(bool) != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
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
	chSum := make(chan Electron, 1)
	chCarry := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case eSum := <-chSum:
				Debug(testName(t, "Select"), fmt.Sprintf("(chSum) Received on Channel (%v), Electron {%s}", chSum, eSum.String()))
				gotSum.Store(eSum.powerState)
				eSum.Done()
			case eCarry := <-chCarry:
				Debug(testName(t, "Select"), fmt.Sprintf("(chCarry) Received on Channel (%v), Electron {%s}", chCarry, eCarry.String()))
				gotCarry.Store(eCarry.powerState)
				eCarry.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	half.Sum.WireUp(chSum)
	half.Carry.WireUp(chCarry)

	if gotSum.Load().(bool) != false {
		t.Error("Wanted no Sum but got one")
	}

	if gotCarry.Load().(bool) != false {
		t.Error("Wanted no Carry but got one")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting input source A to (%t) and source B to (%t)", i, tc.aInPowered, tc.bInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting input source A to (%t) and source B to (%t)", i, tc.aInPowered, tc.bInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)

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

func TestLoopedORGate(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")
	name := testName(t, "Test")

	wireQOut := NewWire(fmt.Sprintf("%s-QOutWire", name), 0)
	defer wireQOut.Shutdown()
	wireQBarOut := NewWire(fmt.Sprintf("%s-QBarOutWire", name), 0)
	defer wireQBarOut.Shutdown()

	sPinBattery := NewBattery("sPinBattery", false)
	QBar := NewORGate(fmt.Sprintf("%s-QBarORGate", name), sPinBattery, wireQOut)
	defer QBar.Shutdown()
	QBar.WireUp(wireQBarOut.Input)

	rPinBattery := NewBattery("rPinBattery", false)
	Q := NewORGate(fmt.Sprintf("%s-QORGate", name), rPinBattery, wireQBarOut)
	defer Q.Shutdown()
	Q.WireUp(wireQOut.Input)

	Debug(testName(t, ""), "Set S")
	sPinBattery.Charge()
	Debug(testName(t, ""), "UnSet S")
	sPinBattery.Discharge()

	Debug(testName(t, ""), "Set R")
	rPinBattery.Charge()
	Debug(testName(t, ""), "UnSet R")
	rPinBattery.Discharge()

	Debug(testName(t, ""), "Set S")
	sPinBattery.Charge()
	Debug(testName(t, ""), "Set R")
	rPinBattery.Charge()

	Debug(testName(t, ""), "UnSet S")
	sPinBattery.Discharge()
	Debug(testName(t, ""), "UnSet R")
	rPinBattery.Discharge()
}

func TestLoopedANDGate(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")
	name := testName(t, "Test")

	wireQOut := NewWire(fmt.Sprintf("%s-QOutWire", name), 0)
	defer wireQOut.Shutdown()
	wireQBarOut := NewWire(fmt.Sprintf("%s-QBarOutWire", name), 0)
	defer wireQBarOut.Shutdown()

	sPinBattery := NewBattery("sPinBattery", false)
	QBar := NewANDGate(fmt.Sprintf("%s-QBarANDGate", name), sPinBattery, wireQOut)
	defer QBar.Shutdown()
	QBar.WireUp(wireQBarOut.Input)

	rPinBattery := NewBattery("rPinBattery", false)
	Q := NewANDGate(fmt.Sprintf("%s-QANDGate", name), rPinBattery, wireQBarOut)
	defer Q.Shutdown()
	Q.WireUp(wireQOut.Input)

	Debug(testName(t, ""), "Set S")
	sPinBattery.Charge()
	Debug(testName(t, ""), "UnSet S")
	sPinBattery.Discharge()

	Debug(testName(t, ""), "Set R")
	rPinBattery.Charge()
	Debug(testName(t, ""), "UnSet R")
	rPinBattery.Discharge()

	Debug(testName(t, ""), "Set S")
	sPinBattery.Charge()
	Debug(testName(t, ""), "Set R")
	rPinBattery.Charge()

	Debug(testName(t, ""), "UnSet S")
	sPinBattery.Discharge()
	Debug(testName(t, ""), "UnSet R")
	rPinBattery.Discharge()
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
	chSum := make(chan Electron, 1)
	chCarry := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case eSum := <-chSum:
				Debug(testName(t, "Select"), fmt.Sprintf("(chSum) Received on Channel (%v), Electron {%s}", chSum, eSum.String()))
				gotSum.Store(eSum.powerState)
				eSum.Done()
			case eCarry := <-chCarry:
				Debug(testName(t, "Select"), fmt.Sprintf("(chCarry) Received on Channel (%v), Electron {%s}", chCarry, eCarry.String()))
				gotCarry.Store(eCarry.powerState)
				eCarry.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	full.Sum.WireUp(chSum)
	full.Carry.WireUp(chCarry)

	if gotSum.Load().(bool) != false {
		t.Error("Wanted no Sum but got one")
	}

	if gotCarry.Load().(bool) != false {
		t.Error("Wanted no Carry but got one")
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Setting input source A to (%t) and source B to (%t) with carry in of (%t)", i, tc.aInPowered, tc.bInPowered, tc.carryInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Setting input source A to (%t) and source B to (%t) with carry in of (%t)", i, tc.aInPowered, tc.bInPowered, tc.carryInPowered))

			aSwitch.Set(tc.aInPowered)
			bSwitch.Set(tc.bInPowered)
			cSwitch.Set(tc.carryInPowered)

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
		t.Error("Expected an adder to return due to good inputs, but got a nil one.")
	} else {
		defer addr.Shutdown()
	}

	var gots [9]atomic.Value // 0-7 for sums, 8 for carryout
	var chStates []chan Electron
	var chStops []chan bool
	for i := 0; i < 9; i++ {
		gots[i].Store(false)
		chStates = append(chStates, make(chan Electron, 1))
		chStops = append(chStops, make(chan bool, 1))
		go func(chState chan Electron, chStop chan bool, i int) {
			for {
				select {
				case e := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", chState, e.String()))
					gots[i].Store(e.powerState)
					e.Done()
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
		t.Run(fmt.Sprintf("testCases[%d]: Adding (%s) to (%s) with carry in of (%t)", i, tc.byte1, tc.byte2, tc.carryInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Adding (%s) to (%s) with carry in of (%t)", i, tc.byte1, tc.byte2, tc.carryInPowered))

			addend1Switches.SetSwitches(tc.byte1)
			addend2Switches.SetSwitches(tc.byte2)
			carryInSwitch.Set(tc.carryInPowered)

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
		t.Error("Expected an adder to return due to good inputs, but got a nil one.")
	} else {
		defer addr.Shutdown()
	}

	var gots [17]atomic.Value // 0-15 for sums, 16 for carryout
	var chStates []chan Electron
	var chStops []chan bool
	for i := 0; i < 17; i++ {
		gots[i].Store(false)
		chStates = append(chStates, make(chan Electron, 1))
		chStops = append(chStops, make(chan bool, 1))
		go func(chState chan Electron, chStop chan bool, i int) {
			for {
				select {
				case e := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", chState, e.String()))
					gots[i].Store(e.powerState)
					e.Done()
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
		t.Run(fmt.Sprintf("testCases[%d]: Adding (%s) to (%s) with carry in of (%t)", i, tc.bytes1, tc.bytes2, tc.carryInPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Adding (%s) to (%s) with carry in of (%t)", i, tc.bytes1, tc.bytes2, tc.carryInPowered))

			addend1Switches.SetSwitches(tc.bytes1)
			addend2Switches.SetSwitches(tc.bytes2)
			carryInSwitch.Set(tc.carryInPowered)

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

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]: Executing complementer against (%s) with compliment signal of (%t)", i, tc.bits, tc.signalIsPowered), func(t *testing.T) {

			Debug(testName(t, ""), fmt.Sprintf("testCases[%d]: Executing complementer against (%s) with compliment signal of (%t)", i, tc.bits, tc.signalIsPowered))

			bitSwitches, _ := NewNSwitchBank(testName(t, "bitSwitches"), tc.bits)
			defer bitSwitches.Shutdown()

			signalSwitch := NewSwitch(testName(t, "signalSwitch"), tc.signalIsPowered)
			defer signalSwitch.Shutdown()

			comp := NewOnesComplementer(testName(t, "OnesComplementer"), bitSwitches.Switches(), signalSwitch)

			if comp == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
			} else {
				defer comp.Shutdown()
			}

			gotCompliments := make([]atomic.Value, len(tc.bits))
			var chStates []chan Electron
			var chStops []chan bool
			for i := 0; i < len(tc.bits); i++ {
				gotCompliments[i].Store(false)
				chStates = append(chStates, make(chan Electron, 1))
				chStops = append(chStops, make(chan bool, 1))
				go func(chState chan Electron, chStop chan bool, index int) {
					for {
						select {
						case e := <-chState:
							Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", chState, e.String()))
							gotCompliments[index].Store(e.powerState)
							e.Done()
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
		t.Error("Expected a subtractor to return due to good inputs, but got a nil one.")
	} else {
		defer sub.Shutdown()
	}

	var gots [9]atomic.Value // 0-7 for diffs, 8 for carryout
	var chStates []chan Electron
	var chStops []chan bool
	for i := 0; i < 9; i++ {
		gots[i].Store(false)
		chStates = append(chStates, make(chan Electron, 1))
		chStops = append(chStops, make(chan bool, 1))
		go func(chState chan Electron, chStop chan bool, index int) {
			for {
				select {
				case e := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", chState, e.String()))
					gots[index].Store(e.powerState)
					e.Done()
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
			ch := make(chan Electron, 1)

			gotResults.Store("")
			chStop := make(chan bool, 1)
			go func() {
				for {
					result := gotResults.Load().(string)
					select {
					case e := <-ch:
						Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", ch, e.String()))
						if e.powerState {
							result += "1"
						} else {
							result += "0"
						}
						gotResults.Store(result)
						e.Done()
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
				t.Errorf("Wanted results of at least %s but got %s.", tc.wantResults, gotResults.Load().(string))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}

func TestRSFlipFlop(t *testing.T) {
	testCases := []struct {
		rPinIsPowered bool
		sPinIsPowered bool
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
			priorR = testCases[i-1].rPinIsPowered
			priorS = testCases[i-1].sPinIsPowered
		}

		return fmt.Sprintf("testCases[%d]: Switching from [rInIsPowered (%t), sInIsPowered (%t)] to [rInIsPowered (%t), sInIsPowered (%t)]", i, priorR, priorS, testCases[i].rPinIsPowered, testCases[i].sPinIsPowered)
	}

	Debug(testName(t, ""), "Initial Setup")

	var rPinBattery, sPinBattery *Battery
	rPinBattery = NewBattery(testName(t, "rBattery"), false)
	sPinBattery = NewBattery(testName(t, "sBattery"), false)

	// starting with no input signals (R and S are off)
	ff := NewRSFlipFlop(testName(t, "RSFlipFlop"), rPinBattery, sPinBattery)
	defer ff.Shutdown()

	var gotQ, gotQBar atomic.Value
	chQ := make(chan Electron, 1)
	chQBar := make(chan Electron, 1)
	chStopQ := make(chan bool, 1)
	chStopQBar := make(chan bool, 1)
	go func() {
		for {
			select {
			case eQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Electron {%s}", chQBar, eQBar.String()))
				gotQBar.Store(eQBar.powerState)
				eQBar.Done()
			case <-chStopQBar:
				return
			}
		}
	}()
	defer func() { chStopQBar <- true }()
	go func() {
		for {
			select {
			case eQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Electron {%s}", chQ, eQ.String()))
				gotQ.Store(eQ.powerState)
				eQ.Done()
			case <-chStopQ:
				return
			}
		}
	}()
	defer func() { chStopQ <- true }()

	ff.QBar.WireUp(chQBar)
	ff.Q.WireUp(chQ)

	if gotQ.Load().(bool) != false {
		t.Errorf("Wanted power of %t at Q, but got %t.", false, gotQ.Load().(bool))
	}

	if gotQBar.Load().(bool) != true {
		t.Errorf("Wanted power of %t at QBar, but got %t.", true, gotQBar.Load().(bool))
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	var summary string
	for i, tc := range testCases {
		summary = testNameDetail(i)
		t.Run(summary, func(t *testing.T) {

			Debug(testName(t, ""), summary)

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

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar.Load().(bool))
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
			// trues since starting with charged batteries when Newing thew Latch initially
			priorClkIn = true
			priorDataIn = true
		} else {
			priorClkIn = testCases[i-1].clkIn
			priorDataIn = testCases[i-1].dataIn
		}

		return fmt.Sprintf("testCases[%d]: Switching from [clkIn (%t) dataIn (%t)] to [clkIn (%t) dataIn (%t)]", i, priorClkIn, priorDataIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	Debug(testName(t, ""), "Initial Setup")

	var clkBattery, dataBattery *Battery
	clkBattery = NewBattery(testName(t, "clkBattery"), true)
	dataBattery = NewBattery(testName(t, "dataBattery"), true)

	// starting with true input signals (Clk and Data are on)
	latch := NewLevelTriggeredDTypeLatch(testName(t, "LevelTriggeredDTypeLatch"), clkBattery, dataBattery)
	defer latch.Shutdown()

	chQ := make(chan Electron, 1)
	chQBar := make(chan Electron, 1)
	chStop := make(chan bool, 1)

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case eQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Electron {%s}", chQBar, eQBar.String()))
				gotQBar.Store(eQBar.powerState)
				eQBar.Done()
			case eQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Electron {%s}", chQ, eQ.String()))
				gotQ.Store(eQ.powerState)
				eQ.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	if gotQ.Load().(bool) != true {
		t.Errorf("Wanted power of %t at Q, but got %t.", true, gotQ.Load().(bool))
	}

	if gotQBar.Load().(bool) != false {
		t.Errorf("Wanted power of %t at QBar, but got %t.", false, gotQBar.Load().(bool))
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	var summary string
	for i, tc := range testCases {
		summary = testNameDetail(i)
		t.Run(summary, func(t *testing.T) {

			Debug(testName(t, ""), summary)

			if tc.clkIn {
				clkBattery.Charge()
			} else {
				clkBattery.Discharge()
			}
			if tc.dataIn {
				dataBattery.Charge()
			} else {
				dataBattery.Discharge()
			}

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar.Load().(bool))
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
	var chStates []chan Electron
	var chStops []chan bool
	for i, q := range latch.Qs {

		chStates = append(chStates, make(chan Electron, 1))
		chStops = append(chStops, make(chan bool, 1))

		go func(chState chan Electron, chStop chan bool, index int) {
			for {
				select {
				case e := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Latches[%d]) Received on Channel (%v), Electron {%s}", index, chState, e.String()))
					gots[index].Store(e.powerState)
					e.Done()
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
	var chStates []chan Electron
	var chStops []chan bool
	for i, o := range sel.Outs {

		chStates = append(chStates, make(chan Electron, 1))
		chStops = append(chStops, make(chan bool, 1))

		go func(chState chan Electron, chStop chan bool, index int) {
			for {
				select {
				case e := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("(SelectorOuts[%d]) Received on Channel (%v), Electron {%s}", index, chState, e.String()))
					gots[index].Store(e.powerState)
					e.Done()
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
	var chStates []chan Electron
	var chStops []chan bool
	for i, o := range sel.Outs {

		chStates = append(chStates, make(chan Electron, 1))
		chStops = append(chStops, make(chan bool, 1))

		go func(chState chan Electron, chStop chan bool, index int) {
			for {
				select {
				case e := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("(SelectorOuts[%d]) Received on Channel (%v), Electron {%s}", index, chState, e.String()))
					gots[index].Store(e.powerState)
					e.Done()
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
		if gots[i].Load().(bool) == true {
			t.Error("Expecting false on all Outs of selector but got a true")
		}
	}

	selectBSwitch.Set(true)

	// selecting B, get B's state
	for i := 0; i < 3; i++ {
		if gots[i].Load().(bool) == false {
			t.Error("Expecting true on all Outs of selector but got a false")
		}
	}

	aInSwitches.SetSwitches("101")

	// still selecting B, get B's state, regardless of A's state changing
	for i := 0; i < 3; i++ {
		if gots[i].Load().(bool) == false {
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
		t.Error("Did not expect an adder back but got one.")
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
	var chSumStates []chan Electron
	var chSumStops []chan bool
	for i, s := range addr.Sums {

		chSumStates = append(chSumStates, make(chan Electron, 1))
		chSumStops = append(chSumStops, make(chan bool, 1))
		gotSums[i].Store(false)

		go func(chSumState chan Electron, chSumStop chan bool, index int) {
			for {
				select {
				case e := <-chSumState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Received on Channel (%v), Electron {%s}", index, chSumState, e.String()))
					gotSums[index].Store(e.powerState)
					e.Done()
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
	chCarryOutState := make(chan Electron, 1)
	chCarryOutStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-chCarryOutState:
				Debug(testName(t, "Select"), fmt.Sprintf("(CarryOut) Received on Channel (%v), Electron {%s}", chCarryOutState, e.String()))
				gotCarryOut.Store(e.powerState)
				e.Done()
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
	var chSumStates []chan Electron
	var chSumStops []chan bool
	for i, s := range addr.Sums {

		chSumStates = append(chSumStates, make(chan Electron, 1))
		chSumStops = append(chSumStops, make(chan bool, 1))
		gotSums[i].Store(false)

		go func(chSumState chan Electron, chSumStop chan bool, index int) {
			for {
				select {
				case e := <-chSumState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Received on Channel (%v), Electron {%s}", index, chSumState, e.String()))
					gotSums[index].Store(e.powerState)
					e.Done()
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
	chCarryOutState := make(chan Electron, 1)
	chCarryOutStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-chCarryOutState:
				Debug(testName(t, "Select"), fmt.Sprintf("(CarryOut) Received on Channel (%v), Electron {%s}", chCarryOutState, e.String()))
				gotCarryOut.Store(e.powerState)
				e.Done()
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

	var clrBattery, clkBattery, dataBattery *Battery
	clrBattery = NewBattery(testName(t, "clrBattery"), false)
	clkBattery = NewBattery(testName(t, "clkBattery"), true)
	dataBattery = NewBattery(testName(t, "dataBattery"), true)

	latch := NewLevelTriggeredDTypeLatchWithClear(testName(t, "LevelTriggeredDTypeLatchWithClear"), clrBattery, clkBattery, dataBattery)
	defer latch.Shutdown()

	chQ := make(chan Electron, 1)
	chQBar := make(chan Electron, 1)
	chStop := make(chan bool, 1)

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case eQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Electron {%s}", chQBar, eQBar.String()))
				gotQBar.Store(eQBar.powerState)
				eQBar.Done()
			case eQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Electron {%s}", chQ, eQ.String()))
				gotQ.Store(eQ.powerState)
				eQ.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	if gotQ.Load().(bool) != true {
		t.Errorf("Wanted power of %t at Q, but got %t.", true, gotQ.Load().(bool))
	}

	if gotQBar.Load().(bool) != false {
		t.Errorf("Wanted power of %t at QBar, but got %t.", false, gotQBar.Load().(bool))
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	for i, tc := range testCases {
		t.Run(testNameDetail(i), func(t *testing.T) {

			Debug(testName(t, ""), testNameDetail(i))

			if tc.clrIn {
				clrBattery.Charge()
			} else {
				clrBattery.Discharge()
			}
			if tc.clkIn {
				clkBattery.Charge()
			} else {
				clkBattery.Discharge()
			}
			if tc.dataIn {
				dataBattery.Charge()
			} else {
				dataBattery.Discharge()
			}

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar.Load().(bool))
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

	clrSwitch := NewSwitch(testName(t, "clrBattery"), false)
	defer clrSwitch.Shutdown()

	clkSwitch := NewSwitch(testName(t, "clkSwitch"), false)
	defer clkSwitch.Shutdown()

	latch := NewNBitLevelTriggeredDTypeLatchWithClear(testName(t, "NBitLevelTriggeredDTypeLatchWithClear"), clrSwitch, clkSwitch, dataSwitches.Switches())
	defer latch.Shutdown()

	// build listen/transmit funcs to deal with each latch's Q
	gots := make([]atomic.Value, len(latch.Qs))
	var chStates []chan Electron
	var chStops []chan bool
	for i, q := range latch.Qs {

		chStates = append(chStates, make(chan Electron, 1))
		chStops = append(chStops, make(chan bool, 1))

		go func(chState chan Electron, chStop chan bool, index int) {
			for {
				select {
				case e := <-chState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Latches[%d]) Received on Channel (%v), Electron {%s}", index, chState, e.String()))
					gots[index].Store(e.powerState)
					e.Done()
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

/*
// TestNNumberAdder creates an adder loop that has no bounds so it is expected to stack overlow
//     runtime: goroutine stack exceeds 1000000000-byte limit
//     fatal error: stack overflow
func TestNNumberAdder(t *testing.T) {

	Debug(testName(t, ""), "Initial Setup")

	switches, _ := NewNSwitchBank(testName(t, "switches"), "1")
	defer switches.Shutdown()

	addr, _ := NewNNumberAdder(testName(t, "NNumberAdder"), switches.Switches())
	defer addr.Shutdown()

	// build listen/transmit funcs to deal with each of the NNumberAdder's Q outputs (the latest inner-adder's Sum sent through)
	var gotSums [1]atomic.Value
	var chSumStates []chan Electron
	var chSumStops []chan bool
	for i, s := range addr.Sums {

		chSumStates = append(chSumStates, make(chan Electron, 1))
		chSumStops = append(chSumStops, make(chan bool, 1))
		gotSums[i].Store(false)

		go func(chSumState chan Electron, chSumStop chan bool, index int) {
			for {
				select {
				case e := <-chSumState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Received on Channel (%v), Electron {%s}", index, chSumState, e.String()))
					gotSums[index].Store(e.powerState)
					e.Done()
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

	// build listen/transmit funcs to deal with each of the adder's INTERNAL adder's sum outputs
	var gotInnerSums [1]atomic.Value
	var chInnerSumStates []chan Electron
	var chInnerSumStops []chan bool
	for i, s := range addr.adder.Sums {

		chInnerSumStates = append(chInnerSumStates, make(chan Electron, 1))
		chInnerSumStops = append(chInnerSumStops, make(chan bool, 1))
		gotInnerSums[i].Store(false)

		go func(chInnerSumState chan Electron, chInnerSumStop chan bool, index int) {
			for {
				select {
				case e := <-chInnerSumState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Received on Channel (%v), Electron {%s}", index, chInnerSumState, e.String()))
					gotInnerSums[index].Store(e.powerState)
					e.Done()
				case <-chInnerSumStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(Sums[%d]) Stopped", index))
					return
				}
			}
		}(chInnerSumStates[i], chInnerSumStops[i], i)

		s.WireUp(chInnerSumStates[i])
	}
	defer func() {
		for i := 0; i < len(addr.adder.Sums); i++ {
			chInnerSumStops[i] <- true
		}
	}()

	Debug(testName(t, ""), "Start Test Cases")

	// totally putting fake setup here to test some things.  this test needs to be reworked to trap the expected states once I get it to work
	// totally putting fake setup here to test some things.  this test needs to be reworked to trap the expected states once I get it to work
	// totally putting fake setup here to test some things.  this test needs to be reworked to trap the expected states once I get it to work

	// can't really prove that Clear clears anything since the internal Add switch defaults to false so the latch doesn't get the adders answer in the first place
	addr.Clear.Set(true)
	addr.Clear.Set(false)

	// regardless of sending non-zero into the NBitAdder's input switches, the fact the Add switch isn't on would prevent any 1s from getting into the latch, so expect no 1s
	want := "0"

	if got := getAnswerString(gotSums[:]); got != want {
		t.Errorf("[Initial setup] Wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
	}

	// however, the internal adder aspect of the NBitAdder should have state based on the initial switches
	want = "1"
	if got := getAnswerString(gotInnerSums[:]); got != want {
		t.Errorf("[Initial setup] Wanted answer of NNumberAdder's inner-adder to be %s but got %s", want, got)
	}

	// this deadlocks once true.  booooooooo
	//	addr.Add.Set(true)
	addr.Add.Set(false)

	// !!! would want 1 here if the Add Set worked, yes?  allowing the latch to get fed the latest inner-adder's state
	want = "0"
	if got := getAnswerString(gotSums[:]); got != want {
		t.Errorf("After an add, wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
	}

	switches.SetSwitches("0")

	// the internal adder aspect of the NBitAdder should have state based on the switches each time
	want = "0"
	if got := getAnswerString(gotInnerSums[:]); got != want {
		t.Errorf("[Initial setup] Wanted answer of NNumberAdder's inner-adder to be %s but got %s", want, got)
	}

	// want = "00000011"
	// if got := getAnswerString(gotSums[:]); got != want {
	// 	t.Errorf("After another add, wanted answer of NNumberAdder (the latch output) to be %s but got %s", want, got)
	// }

	Debug(testName(t, ""), "End Test Cases")
}
*/
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
			// trues since starting with charged batteries when Newing thew Latch initially
			priorClkIn = false
			priorDataIn = false
		} else {
			priorClkIn = testCases[i-1].clkIn
			priorDataIn = testCases[i-1].dataIn
		}

		return fmt.Sprintf("testCases[%d]: Switching from [clkIn (%t) dataIn (%t)] to [clkIn (%t) dataIn (%t)]", i, priorClkIn, priorDataIn, testCases[i].clkIn, testCases[i].dataIn)
	}

	Debug(testName(t, ""), "Initial Setup")

	var clkBattery, dataBattery *Battery
	clkBattery = NewBattery(testName(t, "clkBattery"), false)
	dataBattery = NewBattery(testName(t, "dataBattery"), false)

	latch := NewEdgeTriggeredDTypeLatch(testName(t, "EdgeTriggeredDTypeLatch"), clkBattery, dataBattery)
	defer latch.Shutdown()

	chQ := make(chan Electron, 1)
	chQBar := make(chan Electron, 1)
	chStop := make(chan bool, 1)

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case eQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Electron {%s}", chQBar, eQBar.String()))
				gotQBar.Store(eQBar.powerState)
				eQBar.Done()
			case eQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Electron {%s}", chQ, eQ.String()))
				gotQ.Store(eQ.powerState)
				eQ.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	if gotQ.Load().(bool) != false {
		t.Errorf("Wanted power of %t at Q, but got %t.", false, gotQ.Load().(bool))
	}

	if gotQBar.Load().(bool) != true {
		t.Errorf("Wanted power of %t at QBar, but got %t.", true, gotQBar.Load().(bool))
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	var summary string
	for i, tc := range testCases {
		summary = testNameDetail(i)
		t.Run(summary, func(t *testing.T) {

			Debug(testName(t, ""), summary)

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

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar.Load().(bool))
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
	chOsc := make(chan Electron, 1)
	gotOscResults.Store("")
	chStopOsc := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-chOsc:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", chOsc, e.String()))
				result := gotOscResults.Load().(string)
				if e.powerState {
					result += "1"
				} else {
					result += "0"
				}
				gotOscResults.Store(result)
				e.Done()
			case <-chStopOsc:
				return
			}
		}
	}()
	defer func() { chStopOsc <- true }()

	osc.WireUp(chOsc)

	var gotDivResults atomic.Value
	chDiv := make(chan Electron, 1)
	gotDivResults.Store("")
	chStopDiv := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-chDiv:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", chDiv, e.String()))
				result := gotDivResults.Load().(string)
				if e.powerState {
					result += "1"
				} else {
					result += "0"
				}
				gotDivResults.Store(result)
				e.Done()
			case <-chStopDiv:
				return
			}
		}
	}()
	defer func() { chStopDiv <- true }()

	freqDiv.Q.WireUp(chDiv)

	wantOsc := "0"
	if !strings.HasPrefix(gotOscResults.Load().(string), wantOsc) {
		t.Errorf("Wanted oscillator results %s but got %s.", wantOsc, gotOscResults.Load().(string))
	}

	wantDiv := "0"
	if !strings.HasPrefix(gotDivResults.Load().(string), wantDiv) {
		t.Errorf("Wanted divider results %s but got %s.", wantDiv, gotDivResults.Load().(string))
	}

	Debug(testName(t, ""), "Start Test Case")

	osc.Oscillate(2) // 2 times a second

	time.Sleep(time.Second * 4) // for 4 seconds, should give me 8 oscillations and 4 divider pulses

	osc.Stop()

	wantOsc = "01010101"
	if !strings.HasPrefix(gotOscResults.Load().(string), wantOsc) {
		t.Errorf("Wanted oscillator results of at least %s but got %s.", wantOsc, gotOscResults.Load().(string))
	}

	wantDiv = "0101"
	if !strings.HasPrefix(gotDivResults.Load().(string), wantDiv) {
		t.Errorf("Wanted divider results %s but got %s.", wantDiv, gotDivResults.Load().(string))
	}

	Debug(testName(t, ""), "End Test Case")
}
/*
func TestNBitRippleCounter_EightBit(t *testing.T) {
	Debug(testName(t, ""), "Initial Setup")

	osc := NewOscillator(testName(t, "Oscillator"), false)
	counter := NewNBitRippleCounter(testName(t, "NBitRippleCounter"), osc, 8)
	defer counter.Shutdown()

	var gotOscResults atomic.Value
	gotOscResults.Store("")

	// build listen/transmit func to deal with the oscillator
	chOsc := make(chan Electron, 1)
	chStop := make(chan bool, 1)
	go func() {
		for {
			select {
			case e := <-chOsc:
				Debug(testName(t, "Select"), fmt.Sprintf("Received on Channel (%v), Electron {%s}", chOsc, e.String()))
				result := gotOscResults.Load().(string)
				if e.powerState {
					result += "1"
				} else {
					result += "0"
				}
				gotOscResults.Store(result)
				e.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	osc.WireUp(chOsc)

	// build listen/transmit funcs to deal with each of the Counter's Q outputs
	var gotCounterQs [8]atomic.Value
	var gotDivResults [8]atomic.Value
	var chCounterStates []chan Electron
	var chCounterStops []chan bool

	for i, q := range counter.Qs {

		chCounterStates = append(chCounterStates, make(chan Electron, 1))
		chCounterStops = append(chCounterStops, make(chan bool, 1))
		gotCounterQs[i].Store(false)
		gotDivResults[i].Store("")

		go func(chCounterState chan Electron, chCounterStop chan bool, index int) {
			for {
				select {
				case e := <-chCounterState:
					Debug(testName(t, "Select"), fmt.Sprintf("(Qs[%d]) Received on Channel (%v), Electron {%s}", index, chCounterState, e.String()))
					result := gotDivResults[index].Load().(string)
					if e.powerState {
						result += "1"
					} else {
						result += "0"
					}
					gotDivResults[index].Store(result)
					gotCounterQs[i].Store(e.powerState)
					Debug(testName(t, "Latest Answer"), fmt.Sprintf("(Qs[%d]) Caused {%s}", index, getAnswerString(gotCounterQs[:])))
					e.Done()
				case <-chCounterStop:
					Debug(testName(t, "Select"), fmt.Sprintf("(Qs[%d]) Stopped", index))
					return
				}
			}
		}(chCounterStates[i], chCounterStops[i], i)

		q.WireUp(chCounterStates[i])
	}
	defer func() {
		for i := 0; i < len(counter.Qs); i++ {
			chCounterStops[i] <- true
		}
	}()

	wantOsc := "0"
	if !strings.HasPrefix(gotOscResults.Load().(string), wantOsc) {
		t.Errorf("Wanted oscillator results %s but got %s.", wantOsc, gotOscResults.Load().(string))
	}

	wantCounterAnswer := "10101010"
	if gotAnswer := getAnswerString(gotCounterQs[:]); gotAnswer != wantCounterAnswer {
		t.Errorf("Wanted counter results %s but got %s.", wantCounterAnswer, gotAnswer)
	}

	Debug(testName(t, ""), "Start Test Case")

	osc.Oscillate(2) // 2 times a second

	time.Sleep(time.Second * 4) // for 4 seconds, should give me 8 oscillations and something or other final answer from all the counter's Q states

	osc.Stop()

	wantOsc = "01010101"
	if !strings.HasPrefix(gotOscResults.Load().(string), wantOsc) {
		t.Errorf("Wanted oscillator results of at least %s but got %s.", wantOsc, gotOscResults.Load().(string))
	}

	wantCounterAnswer = "10101110"
	if gotAnswer := getAnswerString(gotCounterQs[:]); gotAnswer != wantCounterAnswer {
		t.Errorf("Wanted counter results %s but got %s.", wantCounterAnswer, gotAnswer)
	}

	for i := 0; i < len(gotDivResults); i++ {
		Debug(testName(t, "Inner-Div-Results"), fmt.Sprintf("(Divs[%d]) %s", i, gotDivResults[i].Load().(string)))

	}
	Debug(testName(t, ""), "End Test Case")
}
*/
func TestEdgeTriggeredDTypeLatchWithPresetAndClear(t *testing.T) {
	testCases := []struct {
		presetIn bool
		clearIn  bool
		clkIn    bool
		dataIn   bool
		wantQ    bool
		wantQBar bool
	}{ // construction of the latches will start with a default of clkIn:false, dataIn:false, which causes Q off (QBar on)
		// {false, false, false, true, false, true},  // preset and clear OFF, clkIn staying false should cause no change
		// {false, false, false, false, false, true}, // preset and clear OFF, clkIn staying false should cause no change
		// {false, false, false, true, false, true},  // preset and clear OFF, clkIn staying false should cause no change, regardless of data change
		// {false, false, true, true, true, false},   // preset and clear OFF, clkIn going to true, with dataIn, causes Q on (QBar off)
		// {false, false, true, false, true, false},  // preset and clear OFF, clkIn staying true should cause no change, regardless of data change
		// {false, false, false, false, true, false}, // preset and clear OFF, clkIn going to false should cause no change
		// {false, false, false, true, true, false},  // preset and clear OFF, clkIn staying false should cause no change, regardless of data change
		// {false, false, true, false, false, true},  // preset and clear OFF, clkIn going to true, with no dataIn, causes Q off (QBar on)
		// {false, false, true, true, false, true},   // preset and clear OFF, clkIn staying true should cause no change, regardless of data change
	}

	testNameDetail := func(i int) string {
		var priorPresetIn bool
		var priorClearIn bool
		var priorClkIn bool
		var priorDataIn bool

		if i == 0 {
			// trues since starting with charged batteries when Newing thew Latch initially
			priorPresetIn = false
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

	var presetBattery, clearBattery, clkBattery, dataBattery *Battery
	presetBattery = NewBattery(testName(t, "presetBattery"), false)
	clearBattery = NewBattery(testName(t, "clearBattery"), false)
	clkBattery = NewBattery(testName(t, "clkBattery"), false)
	dataBattery = NewBattery(testName(t, "dataBattery"), false)

	latch := NewEdgeTriggeredDTypeLatchWithPresetAndClear(testName(t, "EdgeTriggeredDTypeLatchWithPresetAndClear"), presetBattery, clearBattery, clkBattery, dataBattery)
	defer latch.Shutdown()

	chQ := make(chan Electron, 1)
	chQBar := make(chan Electron, 1)
	chStop := make(chan bool, 1)

	var gotQ, gotQBar atomic.Value
	go func() {
		for {
			select {
			case eQBar := <-chQBar:
				Debug(testName(t, "Select"), fmt.Sprintf("(QBar) Received on Channel (%v), Electron {%s}", chQBar, eQBar.String()))
				gotQBar.Store(eQBar.powerState)
				eQBar.Done()
			case eQ := <-chQ:
				Debug(testName(t, "Select"), fmt.Sprintf("(Q) Received on Channel (%v), Electron {%s}", chQ, eQ.String()))
				gotQ.Store(eQ.powerState)
				eQ.Done()
			case <-chStop:
				return
			}
		}
	}()
	defer func() { chStop <- true }()

	latch.QBar.WireUp(chQBar)
	latch.Q.WireUp(chQ)

	if gotQ.Load().(bool) != false {
		t.Errorf("Wanted power of %t at Q, but got %t.", false, gotQ.Load().(bool))
	}

	if gotQBar.Load().(bool) != true {
		t.Errorf("Wanted power of %t at QBar, but got %t.", true, gotQBar.Load().(bool))
	}

	Debug(testName(t, ""), "Start Test Cases Loop")

	var summary string
	for i, tc := range testCases {
		summary = testNameDetail(i)
		t.Run(summary, func(t *testing.T) {

			Debug(testName(t, ""), summary)

			if tc.presetIn {
				presetBattery.Charge()
			} else {
				presetBattery.Discharge()
			}
			if tc.clearIn {
				clearBattery.Charge()
			} else {
				clearBattery.Discharge()
			}
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

			if gotQ.Load().(bool) != tc.wantQ {
				t.Errorf("Wanted power of %t at Q, but got %t.", tc.wantQ, gotQ.Load().(bool))
			}

			if gotQBar.Load().(bool) != tc.wantQBar {
				t.Errorf("Wanted power of %t at QBar, but got %t.", tc.wantQBar, gotQBar.Load().(bool))
			}
		})
	}
	Debug(testName(t, ""), "End Test Cases Loop")
}
