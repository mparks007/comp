package circuit

type EightBitSubtractor2 struct {
	adder       *EightBitAdder2
	comp        *OnesComplementer
	Differences [8]bitPublisher
	SignBit     bitPublisher
}

func NewEightBitSubtractor2(minuendPins, subtrahendPins [8]bitPublisher) *EightBitSubtractor2 {
	s := &EightBitSubtractor2{}

	s.comp = NewOnesComplementer2(subtrahendPins[:], NewBattery()) // the Battery ensures the compliment occurs since the complimentor can conditional compliment based that parameter

	var complimentBits [8]bitPublisher
	copy(complimentBits[:], s.comp.Compliments)

	s.adder = NewEightBitAdder2(minuendPins, complimentBits, NewBattery()) // the added Battery is the "+1" to make the "two's compliment"

	s.Differences = s.adder.Sums
	s.SignBit = s.adder.CarryOut

	return s
}

func (s *EightBitSubtractor2) AsAnswerString() string {
	return s.adder.AsAnswerString()
}

func (a *EightBitSubtractor2) SignBitAsBool() bool {
	return a.SignBit.(*ORGate2).isPowered
}
