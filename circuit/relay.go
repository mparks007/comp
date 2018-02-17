package circuit

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Relay is the core circuit used to contruct logic gates
type Relay struct {
	aInIsPowered atomic.Value
	bInIsPowered atomic.Value
	OpenOut      pwrSource
	ClosedOut    pwrSource
	aInCh        chan bool
	bInCh        chan bool
	chADone      chan bool
	chBDone      chan bool
}

// NewRelay will return a relay, which will be controlled by power state changes of the passed in set of pins
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

	// calling these two receive methods explicitly to ensure the 'answers' for the relay outputs, post WireUp calls above, have settled BEFORE returning and letting things wire up to them
	//receiveA()
	//receiveB()
	// ORRRRR...try to wait for settle by using a and b channels here with close statement at the end?
	chAReady := make(chan bool, 1)
	chBReady := make(chan bool, 1)

	// doing aIn and bIn go funcs independently since power could be changing on either one at the "same" time
	go func() {
		<-chAReady
		for {
			select {
			case <-rel.chADone:
				fmt.Println("Returning from A gofunc inside relay")
				return
			default:
				receiveA()
			}
		}
	}()
	go func() {
		<-chBReady
		for {
			select {
			case <-rel.chBDone:
				fmt.Println("Returning from B gofunc inside relay")
				return
			default:
				receiveB()
			}
		}
	}()

	close(chAReady)
	close(chBReady)

	return rel
}

// Shutdown will allow the go funcs, which are handling listen/transmit, to exit
func (r *Relay) Shutdown() {
	fmt.Println("Shutting down A inside relay")
	r.chADone <- true
				close(r.aInCh)
	fmt.Println("Shutting down B inside relay")
	r.chBDone <- true
				close(r.bInCh)
}
