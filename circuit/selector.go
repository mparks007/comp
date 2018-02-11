package circuit

import (
	"errors"
	"fmt"
)

// sel a b  out
//  0  0 x   0
//  0  1 x   1
//  1  x 0   0
//  1  x 1   1

type TwoToOneSelector struct {
	aANDs []*ANDGate
	bANDs []*ANDGate
	Outs  []*ORGate
}

func NewTwoToOneSelector(signal pwrEmitter, aPins, bPins []pwrEmitter) (*TwoToOneSelector, error) {

	if len(aPins) != len(bPins) {
		return nil, errors.New(fmt.Sprintf("Mismatched input lengths. aPins len: %d, bPins len: %d", len(aPins), len(bPins)))
	}

	sel := &TwoToOneSelector{}

	for i := range aPins {
		sel.aANDs = append(sel.aANDs, NewANDGate(NewInverter(signal), aPins[i]))
		sel.bANDs = append(sel.bANDs, NewANDGate(signal, bPins[i]))
		sel.Outs = append(sel.Outs, NewORGate(sel.aANDs[i], sel.bANDs[i]))
	}

	return sel, nil
}
