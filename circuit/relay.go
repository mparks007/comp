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
	chAStop      chan bool     // shutdown channel for A channel's listening loop
	chBStop      chan bool     // shutdown channel for B channel's listening loop
}

func NewRelay(pin1, pin2 pwrEmitter) *Relay {
	return NewNamedRelay("?", pin1, pin2)
}

// NewRelay will return a relay, which will be controlled by power state changes of the passed in set of pins
func NewNamedRelay(name string, pin1, pin2 pwrEmitter) *Relay {
	rel := &Relay{}

	rel.aInCh = make(chan Electron, 1)
	rel.bInCh = make(chan Electron, 1)
	rel.chAStop = make(chan bool, 1)
	rel.chBStop = make(chan bool, 1)

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

	// doing aIn and bIn go funcs independently since power could be changing on either one at the "same" time
	go func() {
		for {
			select {
			case e := <-rel.aInCh:
				Debug(fmt.Sprintf("[%s]: aIn Received (%t) from (%s) on (%v)", name, e.powerState, e.Name, rel.aInCh))
				rel.aInIsPowered.Store(e.powerState)
				transmit()
				e.wg.Done()
			case <-rel.chAStop:
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case e := <-rel.bInCh:
				Debug(fmt.Sprintf("[%s]: bIn Received (%t) from (%s) on (%v)", name, e.powerState, e.Name, rel.bInCh))
				rel.bInIsPowered.Store(e.powerState)
				transmit()
				e.wg.Done()
			case <-rel.chBStop:
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
	r.chAStop <- true
	r.chBStop <- true
}
