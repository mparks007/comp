package circuit

import "sync"

type Relay struct {
	mu         sync.Mutex
	aInPowered bool
	bInPowered bool
	OpenOut    pwrSource
	ClosedOut  pwrSource
}

func NewRelay(pin1, pin2 pwrEmitter) *Relay {
	r := &Relay{}

	r.UpdatePins(pin1, pin2)

	return r
}

func (r *Relay) UpdatePins(pin1, pin2 pwrEmitter) {
	if pin1 != nil {
		pin1.WireUp(r.aInPowerUpdate)
	}
	if pin2 != nil {
		pin2.WireUp(r.bInPowerUpdate)
	}
}

func (r *Relay) aInPowerUpdate(newState bool) {
	if r.aInPowered != newState {
		r.aInPowered = newState
		r.transmit()
	}
}

func (r *Relay) bInPowerUpdate(newState bool) {
	if r.bInPowered != newState {
		r.bInPowered = newState
		r.transmit()
	}
}

func (r *Relay) transmit() {
	r.OpenOut.Transmit(r.aInPowered && !r.bInPowered)
	r.ClosedOut.Transmit(r.aInPowered && r.bInPowered)
}
