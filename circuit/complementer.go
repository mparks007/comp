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

func (o *OnesComplementer) Shutdown() {
	// for i, _ := range o.Complements {
	// 	o.Complements[i].(*XORGate).Shutdown()
	// }
}