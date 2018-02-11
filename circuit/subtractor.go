package circuit

import (
	"fmt"
)

type NBitSubtractor struct {
	adder       *NBitAdder
	comp        *OnesComplementer
	Differences []*XORGate
	CarryOut    *ORGate
}

func NewNBitSubtractor(minuendPins, subtrahendPins []pwrEmitter) (*NBitSubtractor, error) {

	if len(minuendPins) != len(subtrahendPins) {
		return nil, fmt.Errorf("Mismatched input lengths.  Minuend len: %d, Subtrahend len: %d", len(minuendPins), len(subtrahendPins))
	}

	sub := &NBitSubtractor{}
	sub.comp = NewOnesComplementer(subtrahendPins, NewBattery())                 // the Battery ensures the compliment is "On"
	sub.adder, _ = NewNBitAdder(minuendPins, sub.comp.Complements, NewBattery()) // the added Battery is the "+1" to make the "twos compliment"

	// use some better fields for easier external access
	sub.Differences = sub.adder.Sums
	sub.CarryOut = sub.adder.CarryOut

	return sub, nil
}
