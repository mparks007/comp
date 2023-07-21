package circuit

import (
	"fmt"
	"sync/atomic"
)

// Relay is the core circuit used to contruct (non transistor based) logic gates
type Relay struct {
	aInHasCharge atomic.Value // core state flag to track the 'relay arm path' input's charge state
	bInHasCharge atomic.Value // core state flag to track the 'electromagnet path' input's charge state

	OpenOut   chargeSource // external access point of an inactive/disengaged relay
	ClosedOut chargeSource // external access point of an active/engaged relay

	aInCh   chan Charge // channel to track the relay arm path input
	bInCh   chan Charge // channel to track the electromagnet path inpupt
	chAStop chan bool   // shutdown channel for listening loop
	chBStop chan bool   // shutdown channel for listening loop
}

// NewRelay will return a relay, which will be controlled by charge state changes of the passed in set of pins
func NewRelay(name string, pin1, pin2 chargeEmitter) *Relay {
	rel := &Relay{}

	rel.aInCh = make(chan Charge, 1)
	rel.bInCh = make(chan Charge, 1)
	rel.chAStop = make(chan bool, 1)
	rel.chBStop = make(chan bool, 1)

	// default to false (as a boolean defaults)
	rel.aInHasCharge.Store(false)
	rel.bInHasCharge.Store(false)

	// Init these chargeSoures to ensure hasCharge is defaulting to false
	rel.OpenOut.Init()
	rel.ClosedOut.Init()

	rel.OpenOut.Name = fmt.Sprintf("%s-OpenOut", name)
	rel.ClosedOut.Name = fmt.Sprintf("%s-ClosedOut", name)

	transmit := func(c Charge) {
		// using variables to avoid having the private field charge values change inbetween the Open and Closed transmits just below
		aInHasCharge := rel.aInHasCharge.Load().(bool)
		bInHasCharge := rel.bInHasCharge.Load().(bool)

		c.state = aInHasCharge && !bInHasCharge
		rel.OpenOut.Transmit(c)

		c.state = aInHasCharge && bInHasCharge
		rel.ClosedOut.Transmit(c)
	}

	// must do separate go funcs since loopback-based circuits may send aIns processing back around to the relay and we don't want to lock out the bIn case (and vice versa)
	go func() {
		for {
			select {
			case c := <-rel.aInCh:
				Debug(name, fmt.Sprintf("(aIn) Received on Channel (%v), Charge {%s}", rel.aInCh, c.String()))
				rel.aInHasCharge.Store(c.state)
				transmit(c)
				c.Done()
			case <-rel.chAStop:
				Debug(name, "Stopped")
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case c := <-rel.bInCh:
				Debug(name, fmt.Sprintf("(bIn) Received on Channel (%v), Charge {%s}", rel.bInCh, c.String()))
				rel.bInHasCharge.Store(c.state)
				transmit(c)
				c.Done()
			case <-rel.chBStop:
				Debug(name, "Stopped")
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
