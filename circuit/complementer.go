package circuit

type OnesComplementer struct {
	Complements []pwrEmitter
}

func NewOnesComplementer(bits []pwrEmitter, signal pwrEmitter) *OnesComplementer {

	c := &OnesComplementer{}

	for _, b := range bits {
		c.Complements = append(c.Complements, NewXORGate(signal, b))
	}

	return c
}

func (c *OnesComplementer) AsComplementString() string {
	s := ""

	for _, x := range c.Complements {
		if x.(*XORGate).GetIsPowered() {
			s += "1"
		} else {
			s += "0"
		}
	}

	return s
}
