package circuit

import (
	"fmt"
)

type Relay struct {
	aInPowered bool
	bInPowered bool
	OpenOut    pwrSource
	ClosedOut  pwrSource
	ready bool
}

func NewRelay(pin1, pin2 pwrEmitter) *Relay {
	rel := &Relay{}

	rel.UpdatePin(1, pin1)
	rel.UpdatePin(2, pin2)

	return rel
}

func (r *Relay) UpdatePin(pinNum int, pin pwrEmitter) {
	if pinNum < 1 || pinNum > 2 {
		panic(fmt.Sprintf("Invalid relay pin number.  Relays have two pins and the requested pin was (%d)", pinNum))
	}

	if pinNum == 1 {
		if pin != nil {
			pin.WireUp(r.aInPowerUpdate)
		}
	} else {
		if pin != nil {
			pin.WireUp(r.bInPowerUpdate)
		}
	}
}

func (r *Relay) aInPowerUpdate(newState bool) {
	r.ready = true
	if r.aInPowered != newState {
		r.aInPowered = newState
		r.transmit()
	}
}

func (r *Relay) bInPowerUpdate(newState bool) {
	r.ready = true
	if r.bInPowered != newState {
		r.bInPowered = newState
		r.transmit()
	}
}

func (r *Relay) transmit() {
	r.OpenOut.Transmit(r.aInPowered && !r.bInPowered)
	r.ClosedOut.Transmit(r.aInPowered && r.bInPowered)
}
