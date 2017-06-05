package circuit

type Relay2 struct {
	aInPowered bool
	bInPowered bool
	OpenOut    bitPublication
	ClosedOut  bitPublication
}

func NewRelay2(pin1, pin2 bitPublisher) *Relay2 {
	r := &Relay2{}

	if pin1 != nil {
		pin1.Register(r.aInPowerUpdate)
	}
	if pin2 != nil {
		pin2.Register(r.bInPowerUpdate)
	}

	return r
}

func (r *Relay2) aInPowerUpdate(newState bool) {
	if r.aInPowered != newState {
		r.aInPowered = newState
		r.publish()
	}
}

func (r *Relay2) bInPowerUpdate(newState bool) {
	if r.bInPowered != newState {
		r.bInPowered = newState
		r.publish()
	}
}

func (r *Relay2) publish() {
	r.OpenOut.Publish(r.aInPowered && !r.bInPowered)
	r.ClosedOut.Publish(r.aInPowered && r.bInPowered)
}
