package circuit

import "fmt"

// NBitSubtractor allows for the determination of the difference between to binary numbers
type NBitSubtractor struct {
	adder       *NBitAdder
	comp        *OnesComplementer
	Differences []chargeEmitter
	CarryOut    *ORGate
}

// NewNBitSubtractor returns an NBitSubtractor which will return the difference between the values of the two sets of input pins (minuend, subtrahend)
func NewNBitSubtractor(name string, minuendPins, subtrahendPins []chargeEmitter) (*NBitSubtractor, error) {

	if len(minuendPins) != len(subtrahendPins) {
		return nil, fmt.Errorf("Mismatched input lengths.  Minuend len: %d, Subtrahend len: %d", len(minuendPins), len(subtrahendPins))
	}

	sub := &NBitSubtractor{}
	sub.comp = NewOnesComplementer(fmt.Sprintf("%s-OnesComplementer", name), subtrahendPins, NewChargeProvider(fmt.Sprintf("%s-signalBattery", name), true))           // the Battery ensures the complimenter is "On"
	sub.adder, _ = NewNBitAdder(fmt.Sprintf("%s-NBitAdder", name), minuendPins, sub.comp.Complements, NewChargeProvider(fmt.Sprintf("%s-carryInBattery", name), true)) // the added Battery is the "+1" to make the "twos compliment"

	// use some better field names for easier external access
	sub.Differences = sub.adder.Sums
	sub.CarryOut = sub.adder.CarryOut

	return sub, nil
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component, to exit
func (s *NBitSubtractor) Shutdown() {
	s.comp.Shutdown()
	s.adder.Shutdown()
}
