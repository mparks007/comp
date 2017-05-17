package circuit

import (
	"fmt"
	"strings"
	"testing"
)

func TestANDContact(t *testing.T) {
	testCases := []struct {
		sources []emitter
		want    bool
	}{
		{[]emitter{nil}, false},
		{[]emitter{&battery{}}, true},
		{[]emitter{nil, nil}, false},
		{[]emitter{&battery{}, nil}, false},
		{[]emitter{nil, &battery{}}, false},
		{[]emitter{&battery{}, &battery{}}, true},
	}

	stringFromSources := func(sources []emitter) string {
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
		{&battery{}, nil, true},
		{nil, &battery{}, false},
		{&battery{}, &battery{}, false},
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
		{&battery{}, nil, true, false},
		{nil, &battery{}, false, false},
		{&battery{}, &battery{}, false, true},
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

func TestANDGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, false},
		{&battery{}, nil, false},
		{nil, &battery{}, false},
		{&battery{}, &battery{}, true},
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
		{&battery{}, nil, true},
		{nil, &battery{}, true},
		{&battery{}, &battery{}, true},
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
		{&battery{}, nil, true},
		{nil, &battery{}, true},
		{&battery{}, &battery{}, false},
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

func TestNORGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, true},
		{&battery{}, nil, false},
		{nil, &battery{}, false},
		{&battery{}, &battery{}, false},
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

func TestXORGate(t *testing.T) {
	testCases := []struct {
		aIn  emitter
		bIn  emitter
		want bool
	}{
		{nil, nil, false},
		{&battery{}, nil, true},
		{nil, &battery{}, true},
		{&battery{}, &battery{}, false},
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

func TestHalfAdder(t *testing.T) {
	testCases := []struct {
		aIn       emitter
		bIn       emitter
		wantSum   bool
		wantCarry bool
	}{
		{nil, nil, false, false},
		{&battery{}, nil, true, false},
		{nil, &battery{}, true, false},
		{&battery{}, &battery{}, false, true},
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
		{&battery{}, nil, nil, true, false},
		{&battery{}, &battery{}, nil, false, true},
		{&battery{}, &battery{}, &battery{}, true, true},
		{nil, &battery{}, nil, true, false},
		{nil, &battery{}, &battery{}, false, true},
		{nil, nil, &battery{}, true, false},
		{&battery{}, nil, &battery{}, false, true},
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
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.byte1, tc.byte2), func(t *testing.T) {
			a, err := NewEightBitAdder(tc.byte1, tc.byte2, nil)

			if err != nil && !strings.HasPrefix(err.Error(), tc.wantError) {
				t.Error("Unexpected error: " + err.Error())
				return // expecting to have a nil adder here so cannot do further tests using one
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
		{"00000001", "00000000", &battery{}, "00000010", false},
		{"00000000", "00000001", &battery{}, "00000010", false},
		{"10000000", "10000000", nil, "100000000", true},
		{"10000001", "10000000", nil, "100000001", true},
		{"11111111", "11111111", nil, "111111110", true},
		{"11111111", "11111111", &battery{}, "111111111", true},
		{"01111111", "11111111", nil, "101111110", true},
		{"01111111", "11111111", &battery{}, "101111111", true},
		{"10101010", "01010101", nil, "11111111", false},
		{"10101010", "01010101", &battery{}, "100000000", true},
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
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Adding %s to %s", tc.bytes1, tc.bytes2), func(t *testing.T) {
			a, err := NewSixteenBitAdder(tc.bytes1, tc.bytes2, nil)

			if err != nil && !strings.HasPrefix(err.Error(), tc.wantError) {
				t.Error("Unexpected error: " + err.Error())
				return // expecting to have a nil adder here so cannot do further tests using one
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
		{"0000000000000001", "0000000000000000", &battery{}, "0000000000000010", false},
		{"0000000000000000", "0000000000000001", &battery{}, "0000000000000010", false},
		{"1000000000000000", "1000000000000000", nil, "10000000000000000", true},
		{"1000000000000001", "1000000000000000", nil, "10000000000000001", true},
		{"1111111111111111", "1111111111111111", nil, "11111111111111110", true},
		{"1111111111111111", "1111111111111111", &battery{}, "11111111111111111", true},
		{"0000000001111111", "0000000011111111", nil, "0000000101111110", false},
		{"0000000001111111", "0000000011111111", &battery{}, "0000000101111111", false},
		{"1010101010101010", "0101010101010101", nil, "1111111111111111", false},
		{"1010101010101010", "0101010101010101", &battery{}, "10000000000000000", true},
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
