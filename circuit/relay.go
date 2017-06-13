package circuit

import (
	"fmt"
	"sync"
)

type Relay struct {
	mu         sync.Mutex
	aInPowered bool
	bInPowered bool
	OpenOut    pwrSource
	ClosedOut  pwrSource
}

func NewRelay(pin1, pin2 pwrEmitter) *Relay {
	r := &Relay{}

	r.UpdatePin(1, pin1)
	r.UpdatePin(2, pin2)

	return r
}

func (r *Relay) UpdatePin(pinNum int, pin pwrEmitter) {
	if pinNum > 2 {
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
