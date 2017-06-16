package circuit

type OnesComplementer struct {
	Complements []pwrEmitter
}

func NewOnesComplementer(bits []pwrEmitter, signal pwrEmitter) *OnesComplementer {

	comp := &OnesComplementer{}

	for _, b := range bits {
		comp.Complements = append(comp.Complements, NewXORGate(signal, b))
	}

	return comp
}

func (c *OnesComplementer) AsComplementString() string {
	s := ""

	for _, compOr := range c.Complements {
		if compOr.(*XORGate).GetIsPowered() {
			s += "1"
		} else {
			s += "0"
		}
	}

	return s
}
