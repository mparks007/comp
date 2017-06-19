package circuit

type OnesComplementer struct {
	Complements []pwrEmitter
}

func NewOnesComplementer(pins []pwrEmitter, signal pwrEmitter) *OnesComplementer {

	comp := &OnesComplementer{}

	for _, pin := range pins {
		comp.Complements = append(comp.Complements, NewXORGate(signal, pin))
	}

	return comp
}

func (c *OnesComplementer) AsComplementString() string {
	s := ""

	for _, compOR := range c.Complements {
		if compOR.(*XORGate).GetIsPowered() {
			s += "1"
		} else {
			s += "0"
		}
	}

	return s
}
