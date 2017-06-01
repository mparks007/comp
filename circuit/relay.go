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

type pin struct {
	publication
}

type Relay2 struct {
	aInPowered bool
	bInPowered bool
	OpenOut    powerPublisher
	ClosedOut  powerPublisher
}

func NewRelay2(pin1, pin2 powerPublisher) *Relay2 {
	r := &Relay2{}

	r.OpenOut = &pin{}
	r.ClosedOut = &pin{}

	pin1.Subscribe(r.aInPowerChange)
	pin2.Subscribe(r.bInPowerChange)

	return r
}

func (r *Relay2) aInPowerChange(state bool) {
	if r.aInPowered != state {
		r.aInPowered = state
		r.OpenOut.Publish(state && !r.bInPowered)
	}
}

func (r *Relay2) bInPowerChange(state bool) {
	if r.bInPowered != state {
		r.bInPowered = state
		r.ClosedOut.Publish(state && r.aInPowered)
		r.OpenOut.Publish(state && !r.bInPowered)
	}
}
