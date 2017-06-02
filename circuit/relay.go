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

type Relay2 struct {
	aInPowered bool
	bInPowered bool
	OpenOut    publication
	ClosedOut  publication
}

func NewRelay2(pin1, pin2 publisher) *Relay2 {
	r := &Relay2{}

	if pin1 != nil {
		pin1.Register(r.aInPowerChange)
	}
	if pin2 != nil {
		pin2.Register(r.bInPowerChange)
	}

	// force on if batteries
	if _, ok := pin1.(*Battery); ok {
		r.aInPowerChange(true)
	}
	if _, ok := pin2.(*Battery); ok {
		r.bInPowerChange(true)
	}

	return r
}

func (r *Relay2) aInPowerChange(state bool) {
	if r.aInPowered != state {
		r.aInPowered = state

		r.OpenOut.state = r.aInPowered && !r.bInPowered
		r.OpenOut.Publish()

		r.ClosedOut.state = r.aInPowered && r.bInPowered
		r.ClosedOut.Publish()
	}
}

func (r *Relay2) bInPowerChange(state bool) {
	if r.bInPowered != state {
		r.bInPowered = state

		r.OpenOut.state = r.aInPowered && !r.bInPowered
		r.OpenOut.Publish()

		r.ClosedOut.state = r.aInPowered && r.bInPowered
		r.ClosedOut.Publish()
	}
}
