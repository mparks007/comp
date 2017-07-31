package circuit

/*
type NBitSubtractor struct {
	adder       *NBitAdder
	comp        *OnesComplementer
	Differences []pwrEmitter
	CarryOut    pwrEmitter
}

func NewNBitSubtractor(minuendPins, subtrahendPins []pwrEmitter) (*NBitSubtractor, error) {

	if len(minuendPins) != len(subtrahendPins) {
		return nil, errors.New(fmt.Sprintf("Mismatched input lengths.  Minuend len: %d, Subtrahend len: %d", len(minuendPins), len(subtrahendPins)))
	}

	sub := &NBitSubtractor{}
	sub.comp = NewOnesComplementer(subtrahendPins, NewBattery())                 // the Battery ensures the compliment occurs since the complementer can conditional compliment based that parameter
	sub.adder, _ = NewNBitAdder(minuendPins, sub.comp.Complements, NewBattery()) // the added Battery is the "+1" to make the "two'l compliment"

	// use some better fields for easier external access
	sub.Differences = sub.adder.Sums
	sub.CarryOut = sub.adder.CarryOut

	return sub, nil
}

func (s *NBitSubtractor) AsAnswerString() string {
	return s.adder.AsAnswerString()
}

func (s *NBitSubtractor) CarryOutAsBool() bool {
	return s.CarryOut.(*ORGate).GetIsPowered()
}
*/
