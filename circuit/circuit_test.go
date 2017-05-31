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

func TestANDContact(t *testing.T) {
	testCases := []struct {
		sources []emitter
		want    bool
	}{
		{[]emitter{nil}, false},
		{[]emitter{&Battery{}}, true},
		{[]emitter{nil, nil}, false},
		{[]emitter{&Battery{}, nil}, false},
		{[]emitter{nil, &Battery{}}, false},
		{[]emitter{&Battery{}, &Battery{}}, true},
		{nil, false},
	}

	stringFromSources := func(sources []emitter) string {

		if sources == nil {
			return "very nil"
		}

		str := ""
		for _, s := range sources {
			str += fmt.Sprintf("%T,", s)
		}
		return str
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting sources to %s", stringFromSources(tc.sources)), func(t *testing.T) {
			p := newANDContact(tc.sources...)
			if got := p.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestXContact(t *testing.T) {
	testCases := []struct {
		sourceA emitter
		sourceB emitter
		want    bool
	}{
		{nil, nil, false},
		{&Battery{}, nil, true},
		{nil, &Battery{}, false},
		{&Battery{}, &Battery{}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting source A to %T and source B to %T", tc.sourceA, tc.sourceB), func(t *testing.T) {
			p := xContact{tc.sourceA, tc.sourceB}

			if got := p.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestRelay(t *testing.T) {
	testCases := []struct {
		aIn          emitter
		bIn          emitter
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
			r := newRelay(tc.aIn, tc.bIn)

			if got := r.openOut.Emitting(); got != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, got)
			}

			if got := r.closedOut.Emitting(); got != tc.wantAtClosed {
				t.Errorf("Wanted power at the closed position to be %t, but got %t", tc.wantAtClosed, got)
			}
		})
	}
}

func TestInverter(t *testing.T) {
	testCases := []struct {
		bIn        emitter
		wantAtOpen bool
	}{
		{nil, true},
		{&Battery{}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Input as %T.", tc.bIn), func(t *testing.T) {
			i := newInverter(tc.bIn)

			if got := i.openOut.Emitting(); got != tc.wantAtOpen {
				t.Errorf("Wanted power at the open position to be %t, but got %t", tc.wantAtOpen, got)
			}
		})
	}
}

func TestANDGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, false},
		{&Battery{}, nil, false},
		{nil, &Battery{}, false},
		{&Battery{}, &Battery{}, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			g := newANDGate(tc.aIn, tc.bIn)

			if got := g.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestORGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, false},
		{&Battery{}, nil, true},
		{nil, &Battery{}, true},
		{&Battery{}, &Battery{}, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			g := newORGate(tc.aIn, tc.bIn)

			if got := g.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestNANDGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, true},
		{&Battery{}, nil, true},
		{nil, &Battery{}, true},
		{&Battery{}, &Battery{}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			g := newNANDGate(tc.aIn, tc.bIn)

			if got := g.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestInvANDGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, true},
		{&Battery{}, nil, true},
		{nil, &Battery{}, true},
		{&Battery{}, &Battery{}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			g := newNANDGate2(tc.aIn, tc.bIn)

			if got := g.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestNORGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, true},
		{&Battery{}, nil, false},
		{nil, &Battery{}, false},
		{&Battery{}, &Battery{}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			g := newNORGate(tc.aIn, tc.bIn)

			if got := g.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestInvORGate_Emitting(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, true},
		{&Battery{}, nil, false},
		{nil, &Battery{}, false},
		{&Battery{}, &Battery{}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			g := newNORGate2(tc.aIn, tc.bIn)

			if got := g.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestXORGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, false},
		{&Battery{}, nil, true},
		{nil, &Battery{}, true},
		{&Battery{}, &Battery{}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			g := newXORGate(tc.aIn, tc.bIn)

			if got := g.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestXNORGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, true},
		{&Battery{}, nil, false},
		{nil, &Battery{}, false},
		{&Battery{}, &Battery{}, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input A to %T and B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			g := newXNORGate(tc.aIn, tc.bIn)

			if got := g.Emitting(); got != tc.want {
				t.Errorf("Wanted power %t, but got %t", tc.want, got)
			}
		})
	}
}

func TestHalfAdder(t *testing.T) {
	testCases := []struct {
		aIn       emitter
		bIn       emitter
		wantSum   bool
		wantCarry bool
	}{
		{nil, nil, false, false},
		{&Battery{}, nil, true, false},
		{nil, &Battery{}, true, false},
		{&Battery{}, &Battery{}, false, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input source A to %T and source B to %T", tc.aIn, tc.bIn), func(t *testing.T) {
			h := newHalfAdder(tc.aIn, tc.bIn)

			if got := h.sum.Emitting(); got != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, got)
			}

			if got := h.carry.Emitting(); got != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, got)
			}
		})
	}
}

func TestFullAdder(t *testing.T) {
	testCases := []struct {
		aIn       emitter
		bIn       emitter
		carryIn   emitter
		wantSum   bool
		wantCarry bool
	}{
		{nil, nil, nil, false, false},
		{&Battery{}, nil, nil, true, false},
		{&Battery{}, &Battery{}, nil, false, true},
		{&Battery{}, &Battery{}, &Battery{}, true, true},
		{nil, &Battery{}, nil, true, false},
		{nil, &Battery{}, &Battery{}, false, true},
		{nil, nil, &Battery{}, true, false},
		{&Battery{}, nil, &Battery{}, false, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting input source A to %T and source B to %T with carry in of %T", tc.aIn, tc.bIn, tc.carryIn), func(t *testing.T) {
			h := newFullAdder(tc.aIn, tc.bIn, tc.carryIn)

			if got := h.sum.Emitting(); got != tc.wantSum {
				t.Errorf("Wanted sum %t, but got %t", tc.wantSum, got)
			}

			if got := h.carry.Emitting(); got != tc.wantCarry {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarry, got)
			}
		})
	}
}

func TestEightBitAdder_BadInputs(t *testing.T) {
	testCases := []struct {
		byte1     string
		byte2     string
		wantError string
	}{
		{"0000000", "00000000", "First input not in 8-bit binary format:"},  // only 7 bits on first byte
		{"00000000", "0000000", "Second input not in 8-bit binary format:"}, // only 7 bits on second byte
		{"bad", "00000000", "First input not in 8-bit binary format:"},
		{"00000000", "bad", "Second input not in 8-bit binary format:"},
		{"", "", "First input not in 8-bit binary format:"},
		{"X00000000", "00000000", "First input not in 8-bit binary format:"},
		{"00000000", "X00000000", "Second input not in 8-bit binary format:"},
		{"00000000X", "00000000", "First input not in 8-bit binary format:"},
		{"00000000", "00000000X", "Second input not in 8-bit binary format:"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.byte1, tc.byte2), func(t *testing.T) {
			a, err := NewEightBitAdder(tc.byte1, tc.byte2, nil)

			if err != nil && !strings.HasPrefix(err.Error(), tc.wantError) {
				t.Error("Unexpected error: " + err.Error())
			}

			if a != nil {
				t.Error("Did not expect an adder to return due to bad inputs, but got one.")
			}
		})
	}
}

func TestEightBitAdder_GoodInputs(t *testing.T) {
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
			a, err := NewEightBitAdder(tc.byte1, tc.byte2, tc.carryIn)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
				return // on error, expecting to have a nil adder here so cannot do further tests using one
			}

			if a == nil {
				t.Error("Expected an adder to return due to good inputs, but got a nil one.")
				return // cannot continue tests if no adder to test
			}

			if got := a.String(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := a.carryOut.Emitting(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
	}
}

func TestSixteenBitAdder_BadInputs(t *testing.T) {
	testCases := []struct {
		bytes1    string
		bytes2    string
		wantError string
	}{
		{"000000000000000", "0000000000000000", "First input not in 16-bit binary format:"},  // only 15 bits on first byte
		{"0000000000000000", "000000000000000", "Second input not in 16-bit binary format:"}, // only 15 bits on second byte
		{"bad", "0000000000000000", "First input not in 16-bit binary format:"},
		{"0000000000000000", "bad", "Second input not in 16-bit binary format:"},
		{"", "", "First input not in 16-bit binary format:"},
		{"X0000000000000000", "0000000000000000", "First input not in 16-bit binary format:"},
		{"0000000000000000", "X0000000000000000", "Second input not in 16-bit binary format:"},
		{"0000000000000000X", "0000000000000000", "First input not in 16-bit binary format:"},
		{"0000000000000000", "0000000000000000X", "Second input not in 16-bit binary format:"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.bytes1, tc.bytes2), func(t *testing.T) {
			a, err := NewSixteenBitAdder(tc.bytes1, tc.bytes2, nil)

			if err != nil && !strings.HasPrefix(err.Error(), tc.wantError) {
				t.Error("Unexpected error: " + err.Error())
			}

			if a != nil {
				t.Error("Did not expect an adder to return due to bad inputs, but got one.")
			}
		})
	}
}

func TestSixteenBitAdder_GoodInputs(t *testing.T) {
	testCases := []struct {
		bytes1       string
		bytes2       string
		carryIn      emitter
		wantAnswer   string
		wantCarryOut bool
	}{
		{"0000000000000000", "0000000000000000", nil, "0000000000000000", false},
		{"0000000000000001", "0000000000000000", nil, "0000000000000001", false},
		{"0000000000000000", "0000000000000001", nil, "0000000000000001", false},
		{"0000000000000000", "0000000000000000", &Battery{}, "0000000000000001", false},
		{"0000000000000001", "0000000000000000", &Battery{}, "0000000000000010", false},
		{"0000000000000000", "0000000000000001", &Battery{}, "0000000000000010", false},
		{"1000000000000000", "1000000000000000", nil, "10000000000000000", true},
		{"1000000000000001", "1000000000000000", nil, "10000000000000001", true},
		{"1111111111111111", "1111111111111111", nil, "11111111111111110", true},
		{"1111111111111111", "1111111111111111", &Battery{}, "11111111111111111", true},
		{"0000000001111111", "0000000011111111", nil, "0000000101111110", false},
		{"0000000001111111", "0000000011111111", &Battery{}, "0000000101111111", false},
		{"1010101010101010", "0101010101010101", nil, "1111111111111111", false},
		{"1010101010101010", "0101010101010101", &Battery{}, "10000000000000000", true},
		{"1001110110011101", "1101011011010110", nil, "10111010001110011", true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s with carry in of %T", tc.bytes1, tc.bytes2, tc.carryIn), func(t *testing.T) {
			a, err := NewSixteenBitAdder(tc.bytes1, tc.bytes2, tc.carryIn)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
				return // on error, expecting to have a nil adder here so cannot do further tests using one
			}

			if a == nil {
				t.Error("Expected an adder to return due to good inputs, but got a nil one.")
				return // cannot continue tests if no adder to test
			}

			if got := a.String(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}

			if got := a.carryOut.Emitting(); got != tc.wantCarryOut {
				t.Errorf("Wanted carry %t, but got %t", tc.wantCarryOut, got)
			}
		})
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
		bytes1  string
		bytes2  string
		carryIn emitter
	}{
		{"0000000000000000", "0000000000000000", nil},
		{"1111111111111111", "1111111111111111", nil},
		{"0000000000000000", "0000000000000000", &Battery{}},
		{"1111111111111111", "1111111111111111", &Battery{}},
	}

	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("Adding %s to %s with carry in of %T", bm.bytes1, bm.bytes2, bm.carryIn), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NewSixteenBitAdder(bm.bytes1, bm.bytes2, bm.carryIn)
			}
		})
	}
}

func BenchmarkSixteenBitAdder_String(b *testing.B) {
	benchmarks := []struct {
		name    string
		bytes1  string
		bytes2  string
		carryIn emitter
	}{
		{"All zeros", "0000000000000000", "0000000000000000", nil},
		{"All ones", "1111111111111111", "1111111111111111", nil},
	}
	for _, bm := range benchmarks {
		a, _ := NewSixteenBitAdder(bm.bytes1, bm.bytes2, bm.carryIn)
		b.Run(fmt.Sprintf("Adding %s to %s with carry in of %T", bm.bytes1, bm.bytes2, bm.carryIn), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				a.String()
			}
		})
	}
}

func TestOnesCompliment_BadInputs(t *testing.T) {

	testCases := []struct {
		bits      string
		wantError string
	}{
		{"", "Input bits not in binary format:"},
		{"X", "Input bits not in binary format:"},
		{"X0", "Input bits not in binary format:"},
		{"0X", "Input bits not in binary format:"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Complimenting %s", tc.bits), func(t *testing.T) {
			c, err := NewOnesComplementer([]byte(tc.bits), nil)

			if err != nil && !strings.HasPrefix(err.Error(), tc.wantError) {
				t.Error("Unexpected error: " + err.Error())
			}

			if c != nil {
				t.Error("Did not expect a OnesComplementer to return due to bad inputs, but got one.")
			}
		})
	}
}

func TestOnesCompliment_GoodInputs(t *testing.T) {

	testCases := []struct {
		bits   string
		signal emitter
		want   string
	}{
		{"0", nil, "0"},
		{"0", &Battery{}, "1"},
		{"1", nil, "1"},
		{"1", &Battery{}, "0"},
		{"00000000", nil, "00000000"},
		{"00000000", &Battery{}, "11111111"},
		{"11111111", &Battery{}, "00000000"},
		{"10101010", nil, "10101010"},
		{"10101010", &Battery{}, "01010101"},
		{"1010101010101010101010101010101010101010", nil, "1010101010101010101010101010101010101010"},
		{"1010101010101010101010101010101010101010", &Battery{}, "0101010101010101010101010101010101010101"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Complimenting %s with emit of %T", tc.bits, tc.signal), func(t *testing.T) {
			c, err := NewOnesComplementer([]byte(tc.bits), tc.signal)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
				return // on error, expecting to have a nil OnesComplementer here so cannot do further tests using one
			}

			if c == nil {
				t.Error("Expected a valid OnesComplementer to return due to good inputs, but got a nil one.")
				return // cannot continue tests if no OnesComplementer to test
			}

			if got := c.Complement(); got != tc.want {
				t.Errorf(fmt.Sprintf("Wanted %s, but got %s", tc.want, got))
			}
		})
	}
}

func TestEightBitSubtracter_BadInputs(t *testing.T) {

	testCases := []struct {
		minuend    string
		subtrahend string
		wantError  string
	}{
		{"0000000", "00000000", "First input not in 8-bit binary format:"},  // only 7 bits on first byte
		{"00000000", "0000000", "Second input not in 8-bit binary format:"}, // only 7 bits on second byte
		{"bad", "00000000", "First input not in 8-bit binary format:"},
		{"00000000", "bad", "Input bits not in binary format:"},
		{"", "", "Input bits not in binary format:"},
		{"X00000000", "00000000", "First input not in 8-bit binary format:"},
		{"00000000", "X00000000", "Input bits not in binary format:"},
		{"00000000X", "00000000", "First input not in 8-bit binary format:"},
		{"00000000", "00000000X", "Input bits not in binary format:"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Subtracting %s from %s", tc.subtrahend, tc.minuend), func(t *testing.T) {
			s, err := NewEightBitSubtractor(tc.minuend, tc.subtrahend)

			if err != nil && !strings.HasPrefix(err.Error(), tc.wantError) {
				t.Error("Unexpected error: " + err.Error())
			}

			if s != nil {
				t.Error("Did not expect a subtractor to return due to bad inputs, but got one.")
			}
		})
	}
}

func TestEightBitSubtracter_GoodInputs(t *testing.T) {
	testCases := []struct {
		minuend    string
		subtrahend string
		wantAnswer string
	}{
		{"00000000", "00000000", "100000000"}, // 0 - 0 = 0
		{"00000001", "00000000", "100000001"}, // 1 - 0 = 1
		{"00000001", "00000001", "100000000"}, // 1 - 1 = 0
		{"00000011", "00000001", "100000010"}, // 3 - 1 = 2
		{"10000000", "00000001", "101111111"}, // -128 - 1 = 127 signed (or 128 - 1 = 127 unsigned)
		{"11111111", "11111111", "100000000"}, // -1 - -1 = 0 signed (or 255 - 255 = 0 unsigned)
		{"11111111", "00000001", "111111110"}, // -1 - 1 = -2 signed (or 255 - 1 = 254 unsigned)
		{"10000001", "00000001", "110000000"}, // -127 - 1 = -128 signed (or 129 - 1 = 128 unsigned)
		{"11111110", "11111011", "100000011"}, // -2 - -5 = 3 (or 254 - 251 = 3 unsigned)
		{"00000000", "00000001", "11111111"},  // 0 - 1 = -1 signed (or 255 unsigned)
		{"00000010", "00000011", "11111111"},  // 2 - 3 = -1 signed (or 255 unsigned)
		{"11111110", "11111111", "11111111"},  // -2 - -1 = -1 signed or (254 - 255 = 255 unsigned)
		{"10000001", "01111110", "100000011"}, // -127 - 126 = 3 signed or (129 - 126 = 3 unsigned)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Subtracting %s from %s", tc.subtrahend, tc.minuend), func(t *testing.T) {
			s, err := NewEightBitSubtractor(tc.minuend, tc.subtrahend)

			if err != nil {
				t.Error("Unexpected error: " + err.Error())
				return // on error, expecting to have s nil subtractor here so cannot do further tests using one
			}

			if s == nil {
				t.Error("Expected an subtractor to return due to good inputs, but got s nil one.")
				return // cannot continue tests if no subtractor to test
			}

			if got := s.String(); got != tc.wantAnswer {
				t.Errorf("Wanted answer %s, but got %s", tc.wantAnswer, got)
			}
		})
	}
}

// Fragile test due to timing of asking oscillator vs. state of oscillator at the time being asked
func TestOscillator(t *testing.T) {
	testCases := []struct {
		initState     bool
		oscHertz      int
		checkTimes    int
		wantAllTrues  bool
		wantAllFalses bool
		wantTrueFalse bool
	}{
		{false, 10, 1, false, true, false},
		{true, 10, 1, true, false, false},
		{false, 40, 5, false, false, true},
		{true, 40, 10, false, false, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Oscillating at %d hertz, immediate start (%t), checking %d times.", tc.oscHertz, tc.initState, tc.checkTimes), func(t *testing.T) {

			var results string

			o := newOscillator(tc.initState)
			o.Oscillate(tc.oscHertz)

			for i := 0; i < tc.checkTimes; i++ {
				if o.Emitting() {
					results += "T"
				} else {
					results += "F"
				}
				time.Sleep(time.Millisecond * 500)
			}
			o.Stop()

			gotAllTrues := !strings.Contains(results, "F")
			gotAllFalses := !strings.Contains(results, "T")

			if (gotAllTrues != tc.wantAllTrues) || (gotAllFalses != tc.wantAllFalses) {
				t.Errorf(fmt.Sprintf("Wanted all trues (%t), all falses (%t), and mixed (%t), but got results of %s.", tc.wantAllTrues, tc.wantAllFalses, tc.wantTrueFalse, results))
			}
		})
	}
}

func TestRSFlipFlop_Construction(t *testing.T) {
	testCases := []struct {
		rPin      emitter
		sPin      emitter
		wantError string
	}{
		{nil, nil, ""},
		{nil, &Battery{}, ""},
		{&Battery{}, nil, ""},
		{&Battery{}, &Battery{}, "Both inputs of an RS Flip-Flop cannot be powered simultaneously"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Creating as rIn (%T) and sIn (%T)", tc.rPin, tc.sPin), func(t *testing.T) {
			_, err := newRSFlipFLop(tc.rPin, tc.sPin)

			if err != nil && err.Error() != tc.wantError {
				t.Errorf(fmt.Sprintf("Wanted error %s but got %s.", tc.wantError, err.Error()))
			}
		})
	}
}

func TestRSFlipFlop_ValidateInputs(t *testing.T) {
	testCases := []struct {
		rPin      emitter
		sPin      emitter
		wantError string
	}{
		{nil, nil, ""},
		{nil, &Battery{}, ""},
		{&Battery{}, nil, ""},
		{&Battery{}, &Battery{}, "Both inputs of an RS Flip-Flop cannot be powered simultaneously"},
	}

	// starting with no input signals
	f, err := newRSFlipFLop(nil, nil)

	if err != nil {
		t.Error(fmt.Sprintf("Expecting no errors on initial creation but got %s.", err))
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Setting up as rIn (%T) and sIn (%T)", tc.rPin, tc.sPin), func(t *testing.T) {
			err := f.updateInputs(tc.rPin, tc.sPin)

			if err != nil && err.Error() != tc.wantError {
				t.Errorf(fmt.Sprintf("Wanted error %s but got %s.", tc.wantError, err.Error()))
			}
		})
	}
}

func TestRSFlipFlop_qEmitting_InputValidation(t *testing.T) {
	testCases := []struct {
		rPin      emitter
		sPin      emitter
		wantError string
	}{
		//{nil, nil, ""},
		//{nil, &Battery{}, ""},
		//{&Battery{}, nil, ""},
		{&Battery{}, &Battery{}, "Both inputs of an RS Flip-Flop cannot be powered simultaneously"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Creating as rIn (%T) and sIn (%T)", tc.rPin, tc.sPin), func(t *testing.T) {
			var rPin emitter = nil
			var sPin emitter = nil

			// starting with no input signals
			f, err := newRSFlipFLop(rPin, sPin)

			if err != nil {
				t.Error(fmt.Sprintf("Expecting no errors on initial creation but got %s.", err))
			}

			// this doesn't seem to actually change the flip-flop's inner pins (hmmmm)
			// this doesn't seem to actually change the flip-flop's inner pins (hmmmm)
			// this doesn't seem to actually change the flip-flop's inner pins (hmmmm)
			// this doesn't seem to actually change the flip-flop's inner pins (hmmmm)
			rPin = tc.rPin
			sPin = tc.sPin

			_, err = f.qEmitting()

			if err != nil && err.Error() != tc.wantError {
				t.Errorf(fmt.Sprintf("Wanted error %t on Q, but got %t.", tc.wantError, err))
			}
		})
	}
}

func TestRSFlipFlop(t *testing.T) {
	testCases := []struct {
		rPin      emitter
		sPin      emitter
		wantQ     bool
		wantQBar  bool
		wantError string
	}{
		{nil, nil, false, true, ""},
		{nil, &Battery{}, true, false, ""},
		{nil, nil, true, false, ""},
		{&Battery{}, nil, false, true, ""},
		{nil, nil, false, true, ""},
		{nil, &Battery{}, true, false, ""},
		{&Battery{}, &Battery{}, true, false, "Both inputs of an RS Flip-Flop cannot be powered simultaneously"},
	}

	testName := func(i int) string {
		var priorR emitter
		var priorS emitter

		if i == 0 {
			priorR = nil
			priorS = nil
		} else {
			priorR = testCases[i-1].rPin
			priorS = testCases[i-1].sPin
		}

		return fmt.Sprintf("Stage %d: Switching from [rIn (%T) sIn (%T)] to [rIn (%T) sIn (%T)]", i+1, priorR, priorS, testCases[i].rPin, testCases[i].sPin)
	}

	// starting with no input signals
	f, err := newRSFlipFLop(nil, nil)

	if err != nil {
		t.Error(fmt.Sprintf("Expecting no errors on initial creation but got %s.", err))
	}

	for i, tc := range testCases {
		t.Run(testName(i), func(t *testing.T) {
			err := f.updateInputs(tc.rPin, tc.sPin)

			if err != nil && err.Error() != tc.wantError {
				t.Error(fmt.Sprintf("Wanted error %s but got %s.", tc.wantError, err))
			}

			if gotQ, _ := f.qEmitting(); gotQ != tc.wantQ {
				t.Errorf(fmt.Sprintf("Wanted power of %t on Q, but got %t.", tc.wantQ, gotQ))
			}

			if gotQBar, _ := f.qBarEmitting(); gotQBar != tc.wantQBar {
				t.Errorf(fmt.Sprintf("Wanted power of %t on QBar, but got %t.", tc.wantQBar, gotQBar))
			}
		})
	}
}

/*
func TestLtDLatch(t *testing.T) {
	testCases := []struct {
		dataPin      emitter
		clkPin      emitter
		wantQ     bool
		wantQBar  bool
		wantError string
	}{
		{nil, nil, false, true, ""},
		{nil, &Battery{}, true, false, ""},
		{nil, nil, true, false, ""},
		{&Battery{}, nil, false, true, ""},
		{nil, nil, false, true, ""},
		{nil, &Battery{}, true, false, ""},
		{&Battery{}, &Battery{}, true, false, "Both inputs of an RS Flip-Flop cannot be powered simultaneously"},
	}

	testName := func(i int) string {
		var priorR emitter
		var priorS emitter

		if i == 0 {
			priorR = nil
			priorS = nil
		} else {
			priorR = testCases[i-1].rPin
			priorS = testCases[i-1].sPin
		}

		return fmt.Sprintf("Stage %d: Switching from [rIn (%T) sIn (%T)] to [rIn (%T) sIn (%T)]", i+1, priorR, priorS, testCases[i].rPin, testCases[i].sPin)
	}

	// starting with no input signals
	f, err := newRSFlipFLop(nil, nil)

	if err != nil {
		t.Error(fmt.Sprintf("Expecting no errors on initial creation but got %s.", err))
	}

	for i, tc := range testCases {
		t.Run(testName(i), func(t *testing.T) {
			err := f.updateInputs(tc.rPin, tc.sPin)

			if err != nil && err.Error() != tc.wantError {
				t.Error(fmt.Sprintf("Wanted error %s but got %s.", tc.wantError, err))
			}

			if gotQ, _ := f.qEmitting(); gotQ != tc.wantQ {
				t.Errorf(fmt.Sprintf("Wanted power of %t on Q, but got %t.", tc.wantQ, gotQ))
			}

			if gotQBar, _ := f.qBarEmitting(); gotQBar != tc.wantQBar {
				t.Errorf(fmt.Sprintf("Wanted power of %t on QBar, but got %t.", tc.wantQBar, gotQBar))
			}
		})
	}
}
*/
