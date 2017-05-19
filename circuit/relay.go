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

/*
func (r *relay) String() string {

	s := fmt.Sprintf("Relay dump:\n")
	s += fmt.Sprintf("\tInput A emitting: %t\n", r.in.Emitting())
	s += fmt.Sprintf("\tInput B emitting: %t\n", r.in.Emitting())
	s += fmt.Sprintf("\tOpen Output emitting: %t\n", r.openOut.Emitting())
	s += fmt.Sprintf("\tClosed Output emitting: %t\n", r.closedOut.Emitting())

	return s
}
*/
