package circuit

type OnesComplementer struct {
	Compliments []bitPublisher
}

func NewOnesComplementer2(bits []bitPublisher, signal bitPublisher) *OnesComplementer {

	c := &OnesComplementer{}

	for _, b := range bits {
		c.Compliments = append(c.Compliments, NewXORGate2(signal, b))
	}

	return c
}

func (c *OnesComplementer) AsComplimentString() string {
	s := ""

	for _, x := range c.Compliments {
		if x.(*XORGate2).isPowered {
			s += "1"
		} else {
			s += "0"
		}
	}

	return s
}
