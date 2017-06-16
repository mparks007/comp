package circuit

import (
	"errors"
	"fmt"
)

type TwoToOneSelector struct {
	aANDs []*ANDGate
	bANDs []*ANDGate
	Outs  []*ORGate
}

func NewTwoToOneSelector(aPins, bPins []pwrEmitter, signal pwrEmitter) (*TwoToOneSelector, error) {

	if len(aPins) != len(bPins) {
		return nil, errors.New(fmt.Sprintf("Mismatched input lengths. aPins len: %d, bPins len: %d", len(aPins), len(bPins)))
	}

	sel := &TwoToOneSelector{}

	for i := range aPins {
		sel.aANDs = append(sel.aANDs, NewANDGate(aPins[i], NewInverter(signal)))
		sel.bANDs = append(sel.bANDs, NewANDGate(bPins[i], signal))
		sel.Outs = append(sel.Outs, NewORGate(sel.aANDs[i], sel.bANDs[i]))
	}

	return sel, nil
}
