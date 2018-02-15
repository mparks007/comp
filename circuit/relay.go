package circuit

import (
	"sync"
	"sync/atomic"
)

type Relay struct {
	aInIsPowered atomic.Value
	bInIsPowered atomic.Value
	OpenOut      pwrSource
	ClosedOut    pwrSource
	aInCh        chan bool
	bInCh        chan bool
	name         string
	chADone      chan bool
	chBDone      chan bool
}

func NewRelay(pin1, pin2 pwrEmitter) *Relay {
	rel := &Relay{}
	m := &sync.Mutex{}

	rel.aInCh = make(chan bool, 1)
	rel.bInCh = make(chan bool, 1)
	rel.chADone = make(chan bool, 1)
	rel.chBDone = make(chan bool, 1)

	// default to false, as if boolean defaults
	rel.aInIsPowered.Store(false)
	rel.bInIsPowered.Store(false)

	if pin1 != nil {
		pin1.WireUp(rel.aInCh)
	}
	if pin2 != nil {
		pin2.WireUp(rel.bInCh)
	}

	transmit := func() {
		m.Lock() // must lock since receiveA and receiveB might be called concurrently, per their go funcs below, which call this transmit
		defer m.Unlock()

		aInIsPowered := rel.aInIsPowered.Load().(bool)
		bInIsPowered := rel.bInIsPowered.Load().(bool)

		rel.OpenOut.Transmit(aInIsPowered && !bInIsPowered)
		rel.ClosedOut.Transmit(aInIsPowered && bInIsPowered)
	}

	receiveA := func() {
		rel.aInIsPowered.Store(<-rel.aInCh)
		transmit()
	}

	receiveB := func() {
		rel.bInIsPowered.Store(<-rel.bInCh)
		transmit()
	}

	// calling these two receive methods explicitly to ensure the 'answers' for the relay outputs, post UpdatePin calls above, have settled BEFORE returning and letting things wire up to them
	receiveA()
	receiveB()

	// doing aIn and bIn go funcs independently since power could be changing on either one at the "same" time
	go func() {
		for {
			select {
			case <-rel.chADone:
				return
			default:
				receiveA()
			}
		}
	}()
	go func() {
		for {
			select {
			case <-rel.chBDone:
				return
			default:
				receiveB()
			}
		}
	}()

	return rel
}

// Quit allows any 'for' loops inside go funcs to exit
func (r *Relay) Quit() {
	r.chADone <- true
	r.chBDone <- true
}
