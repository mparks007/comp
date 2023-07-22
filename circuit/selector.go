package circuit

import "fmt"

// TwoToOneSelector is a circuit which takes two sets of input pins and exposes a switch to decide with of those input sets will be sent to the output
//
// Truth Table
// sel a  b out
//
//	0  0  x  0
//	0  1  x  1
//	1  x  0  0
//	1  x  1  1
type TwoToOneSelector struct {
	aANDs     []*ANDGate
	bANDs     []*ANDGate
	inverters []*Inverter
	Outs      []chargeEmitter
}

// NewTwoToOneSelector will return an 2-to-1 Selector component whose output will depend on the state of the selector signal input pin
//
//	With selector off, the first set of pins will be the output.  If on, the second set is the output.
func NewTwoToOneSelector(name string, signal chargeEmitter, aPins, bPins []chargeEmitter) (*TwoToOneSelector, error) {

	if len(aPins) != len(bPins) {
		return nil, fmt.Errorf("Mismatched input lengths. aPins len: %d, bPins len: %d", len(aPins), len(bPins))
	}

	sel := &TwoToOneSelector{}

	for i := range aPins {
		sel.inverters = append(sel.inverters, NewInverter(fmt.Sprintf("%s-Inverters[%d]", name, i), signal)) // having to make inverters as named objects so they can be Shutdown later (vs. just feeding NewInverter(signal) into NewANDGate())
		sel.aANDs = append(sel.aANDs, NewANDGate(fmt.Sprintf("%s-aANDGates[%d]", name, i), sel.inverters[i], aPins[i]))
		sel.bANDs = append(sel.bANDs, NewANDGate(fmt.Sprintf("%s-bANDGates[%d]", name, i), signal, bPins[i]))
		sel.Outs = append(sel.Outs, NewORGate(fmt.Sprintf("%s-Outs[%d]", name, i), sel.aANDs[i], sel.bANDs[i]))
	}

	return sel, nil
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (s *TwoToOneSelector) Shutdown() {
	for i := range s.aANDs {
		s.Outs[i].(*ORGate).Shutdown()
		s.bANDs[i].Shutdown()
		s.aANDs[i].Shutdown()
		s.inverters[i].Shutdown()
	}
}
