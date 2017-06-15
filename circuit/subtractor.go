package circuit

type EightBitSubtracter struct {
	adder       *EightBitAdder
	comp        *OnesComplementer
	Differences [8]pwrEmitter
	CarryOut    pwrEmitter
}

func NewEightBitSubtracter(minuendPins, subtrahendPins [8]pwrEmitter) *EightBitSubtracter {
	s := &EightBitSubtracter{}

	s.comp = NewOnesComplementer(subtrahendPins[:], NewBattery()) // the Battery ensures the compliment occurs since the complementer can conditional compliment based that parameter

	var complimentBits [8]pwrEmitter
	copy(complimentBits[:], s.comp.Complements)

	s.adder = NewEightBitAdder(minuendPins, complimentBits, NewBattery()) // the added Battery is the "+1" to make the "two'l compliment"

	// use some better fields for easier, external access
	s.Differences = s.adder.Sums
	s.CarryOut = s.adder.CarryOut

	return s
}

func (s *EightBitSubtracter) AsAnswerString() string {
	return s.adder.AsAnswerString()
}

func (a *EightBitSubtracter) CarryOutAsBool() bool {
	return a.CarryOut.(*ORGate).GetIsPowered()
}
