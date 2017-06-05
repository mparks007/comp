package circuit

type relay struct {
	aIn       emitter
	bIn       emitter
	openOut   emitter
	closedOut emitter
}

func newRelay(pin1, pin2 emitter) *relay {
	return &relay{
		pin1,
		pin2,
		newXContact(pin1, pin2),
		newANDContact(pin1, pin2),
	}
}
