package circuit

type EightBitSubtractor struct {
	adder   *EightBitAdder
	comp    *onesComplimenter
	signBit emitter
}

func NewEightBitSubtractor(byte1, byte2 string) (*EightBitSubtractor, error) {
	s := &EightBitSubtractor{}

	var err error // why did I have to do this?
	s.comp, err = NewOnesComplimenter([]byte(byte2), &Battery{})
	if err != nil {
		return nil, err
	}

	s.adder, err = NewEightBitAdder(byte1, s.comp.Compliment(), &Battery{}) // the added battery is the "+1" to make the "two's compliment"
	if err != nil {
		return nil, err
	}

	s.signBit = s.adder.carryOut

	return s, nil
}

func (s *EightBitSubtractor) String() string {
	return s.adder.String()
}
