package circuit

type OnesComplementer struct {
	xorGates []bitPublisher
}

func NewOnesComplementer2(bits []bitPublisher, signal bitPublisher) *OnesComplementer {

	c := &OnesComplementer{}

	for _, b := range bits {
		c.xorGates = append(c.xorGates, NewXORGate2(signal, b))
	}

	return c
}

func (c *OnesComplementer) AsAnswerString() string {
	s := ""

	for _, x := range c.xorGates {
		if x.(*XORGate2).isPowered {
			s += "1"
		} else {
			s += "0"
		}
	}

	return s
}

func (c *OnesComplementer) AsBitPublishers() []bitPublisher {
	return c.xorGates
}
