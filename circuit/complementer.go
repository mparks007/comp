package circuit

import "fmt"

// OnesComplementer is a circuit which will invert the charge of all input pins when given a signal
type OnesComplementer struct {
	Complements []chargeEmitter
}

// NewOnesComplementer will return a OnesComplementer component which will convert the charge state of every input IF the signal pin is given a charge
func NewOnesComplementer(name string, pins []chargeEmitter, signal chargeEmitter) *OnesComplementer {

	comp := &OnesComplementer{}

	for _, pin := range pins {
		comp.Complements = append(comp.Complements, NewXORGate(fmt.Sprintf("%s-XORGate", name), signal, pin))
	}

	return comp
}

// Shutdown will allow the go funcs, which are handling listen/transmit on the sub-gates, to exit
func (o *OnesComplementer) Shutdown() {
	for _, c := range o.Complements {
		c.(*XORGate).Shutdown()
	}
}
