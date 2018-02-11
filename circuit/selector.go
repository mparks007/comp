package circuit

//"errors"
//"fmt"

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

	// // allowing bPins to be nil since we might want to plug in those pins later
	// if (len(aPins) != len(bPins)) && (len(bPins) != 0) {
	// 	return nil, errors.New(fmt.Sprintf("Mismatched input lengths. aPins len: %d, bPins len: %d", len(aPins), len(bPins)))
	// }

	sel := &TwoToOneSelector{}

	for i := range aPins {
		sel.aANDs = append(sel.aANDs, NewANDGate(NewInverter(signal), aPins[i]))

		// if wanting to plug in bPins later, just set to nil for now (but at least make the AND gate!)
		if len(bPins) == 0 {
			sel.bANDs = append(sel.bANDs, NewANDGate(signal, nil))
		} else {
			sel.bANDs = append(sel.bANDs, NewANDGate(signal, bPins[i]))
		}

		sel.Outs = append(sel.Outs, NewORGate(sel.aANDs[i], sel.bANDs[i]))
	}

	return sel, nil
}
/*
func (s *TwoToOneSelector) UpdateBPins(bPins []pwrEmitter) {

	// TODO: validate bPins is same length as the aANDs

	for i, bPin := range bPins {
		s.bANDs[i].UpdatePin(2, 2, bPin)
	}
}
*/
