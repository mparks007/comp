package circuit

import "sync"

type Relay struct {
	mu         sync.Mutex
	aInPowered bool
	bInPowered bool
	OpenOut    bitPublication
	ClosedOut  bitPublication
}

func NewRelay(pin1, pin2 bitPublisher) *Relay {
	r := &Relay{}

	if pin1 != nil {
		pin1.Register(r.aInPowerUpdate)
	}
	if pin2 != nil {
		pin2.Register(r.bInPowerUpdate)
	}

	return r
}

func (r *Relay) aInPowerUpdate(newState bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.aInPowered != newState {
		r.aInPowered = newState
		r.publish()
	}
}

func (r *Relay) bInPowerUpdate(newState bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.bInPowered != newState {
		r.bInPowered = newState
		r.publish()
	}
}

func (r *Relay) publish() {
	r.OpenOut.Publish(r.aInPowered && !r.bInPowered)
	r.ClosedOut.Publish(r.aInPowered && r.bInPowered)
}
