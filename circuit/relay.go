package circuit

import (
	"fmt"
	"sync/atomic"
)

// Relay is the core circuit used to contruct logic gates
type Relay struct {
	aInIsPowered atomic.Value  // core state flag to track the relay arm path input's current state
	bInIsPowered atomic.Value  // core state flag to track the electromagnet path input's current state
	OpenOut      pwrSource     // external access point to inactive/disengaged relay
	ClosedOut    pwrSource     // external access point to active/engaged relay
	aInCh        chan Electron // channel to track the relay arm path input
	bInCh        chan Electron // channel to track the electromagnet path inpupt
	chStop       chan bool     // shutdown channel for listening loop
}

// NewRelay will return a relay, which will be controlled by power state changes of the passed in set of pins
func NewRelay(name string, pin1, pin2 pwrEmitter) *Relay {
	rel := &Relay{}

	rel.aInCh = make(chan Electron, 1)
	rel.bInCh = make(chan Electron, 1)
	rel.chStop = make(chan bool, 1)

	// default to false (as a boolean defaults)
	rel.aInIsPowered.Store(false)
	rel.bInIsPowered.Store(false)

	// Init these pwrSource types (need to ensure isPowered is defaulting to false)
	rel.OpenOut.Init()
	rel.ClosedOut.Init()

	rel.OpenOut.Name = fmt.Sprintf("%s-OpenOut", name)
	rel.ClosedOut.Name = fmt.Sprintf("%s-ClosedOut", name)

	transmit := func() {
		aInIsPowered := rel.aInIsPowered.Load().(bool)
		bInIsPowered := rel.bInIsPowered.Load().(bool)

		rel.OpenOut.Transmit(aInIsPowered && !bInIsPowered)
		rel.ClosedOut.Transmit(aInIsPowered && bInIsPowered)
	}

	go func() {
		for {
			select {
			case e := <-rel.aInCh:
				Debug(name, fmt.Sprintf("(aIn) Received (%t) from (%s) on (%v)", e.powerState, e.Name, rel.aInCh))
				rel.aInIsPowered.Store(e.powerState)
				transmit()
				e.wg.Done()
			case e := <-rel.bInCh:
				Debug(name, fmt.Sprintf("(bIn) Received (%t) from (%s) on (%v)", e.powerState, e.Name, rel.bInCh))
				rel.bInIsPowered.Store(e.powerState)
				transmit()
				e.wg.Done()
			case <-rel.chStop:
				return
			}
		}
	}()

	pin1.WireUp(rel.aInCh)
	pin2.WireUp(rel.bInCh)

	return rel
}

// Shutdown will allow the go funcs, which are handling listen/transmit, to exit
func (r *Relay) Shutdown() {
	r.chStop <- true
}
