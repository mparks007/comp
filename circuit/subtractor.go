package circuit

type EightBitSubtractor struct {
	adder   *EightBitAdder
	comp    *onesComplementer
	signBit emitter
}

func NewEightBitSubtractor(byte1, byte2 string) (*EightBitSubtractor, error) {
	s := &EightBitSubtractor{}

	var err error                                                // why did I have to do this?
	s.comp, err = NewOnesComplementer([]byte(byte2), &Battery{}) // the battery ensures the compliment occurs since the complimentor can conditional compliment based on that emit
	if err != nil {
		return nil, err
	}

	s.adder, err = NewEightBitAdder(byte1, s.comp.Complement(), &Battery{}) // the added battery is the "+1" to make the "two's compliment"
	if err != nil {
		return nil, err
	}

	s.signBit = s.adder.carryOut

	return s, nil
}

func (s *EightBitSubtractor) String() string {
	return s.adder.String()
}
