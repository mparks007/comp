package circuit

/*
// OnesComplementer is a circuit which will invert the power of all input pins when given a signal
type OnesComplementer struct {
	Complements []pwrEmitter
}

// NewOnesComplementer will return a OnesComplementer component which will convert the power state of every input IF the signal pin is given power
func NewOnesComplementer(pins []pwrEmitter, signal pwrEmitter) *OnesComplementer {

	comp := &OnesComplementer{}

	for _, pin := range pins {
		comp.Complements = append(comp.Complements, NewXORGate(signal, pin))
	}

	return comp
}

// Shutdown will allow the go funcs, which are handling listen/transmit on the sub-gates, to exit
func (o *OnesComplementer) Shutdown() {
	for i := range o.Complements {
		o.Complements[i].(*XORGate).Shutdown()
	}
}

*/
