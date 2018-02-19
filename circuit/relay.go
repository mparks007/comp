package circuit

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Relay is the core circuit used to contruct logic gates
type Relay struct {
	aInIsPowered   atomic.Value // core state flag to track the relay arm path input's current state
	bInIsPowered   atomic.Value // core state flag to track the electromagnet path input's current state
	OpenOut        pwrSource    // external access point to inactive/disengaged relay
	ClosedOut      pwrSource    // external access point to active/engaged relay
	aInCh          chan bool    // channel to track the relay arm path input
	bInCh          chan bool    // channel to track the electromagnet path inpupt
	chAStop        chan bool    // shutdown channel for A channel's listening loop
	chBStop        chan bool    // shutdown channel for B channel's listening loop
	chATransmitted chan bool
	chBTransmitted chan bool
}

// NewRelay will return a relay, which will be controlled by power state changes of the passed in set of pins
func NewRelay(pin1, pin2 pwrEmitter) *Relay {
	rel := &Relay{}
	mu := &sync.Mutex{}

	rel.aInCh = make(chan bool, 1)
	rel.bInCh = make(chan bool, 1)
	rel.chAStop = make(chan bool, 1)
	rel.chBStop = make(chan bool, 1)
	rel.chATransmitted = make(chan bool, 1)
	rel.chBTransmitted = make(chan bool, 1)

	// default to false (as a boolean defaults)
	rel.aInIsPowered.Store(false)
	rel.bInIsPowered.Store(false)

	transmit := func() {
		mu.Lock() // must lock since receiveA and receiveB might be called concurrently, per their go funcs below, which call this transmit
		defer mu.Unlock()

		aInIsPowered := rel.aInIsPowered.Load().(bool)
		bInIsPowered := rel.bInIsPowered.Load().(bool)

		rel.OpenOut.Transmit(aInIsPowered && !bInIsPowered)
		fmt.Println("Before <-rel.OpenOut.chTransmitted")
		<-rel.OpenOut.chTransmitted
		fmt.Println("After <-rel.OpenOut.chTransmitted")
		rel.ClosedOut.Transmit(aInIsPowered && bInIsPowered)
		fmt.Println("Before <-rel.ClosedOut.chTransmitted")
		<-rel.ClosedOut.chTransmitted
		fmt.Println("After <-rel.ClosedOut.chTransmitted")
	}

	// doing aIn and bIn go funcs independently since power could be changing on either one at the "same" time
	go func() {
		for {
			select {
			case aState := <-rel.aInCh:
				rel.aInIsPowered.Store(aState)
				transmit()
			case <-rel.chAStop:
				fmt.Println("DEBUG: Bailing from A Relay go func loop")
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case bState := <-rel.bInCh:
				rel.bInIsPowered.Store(bState)
				transmit()
			case <-rel.chBStop:
				fmt.Println("DEBUG: Bailing from B Relay go func loop")
				return
			}
		}
	}()

	pin1.WireUp(rel.aInCh)
	pin2.WireUp(rel.bInCh)

	//pin1.WaitForTransmit() // add this to interface?  let this method do the <-chTransmitted stuff vs raw code?
	//<-pin2.chTransmitted

	// receiveA := func() {
	//rel.aInIsPowered.Store(<-rel.aInCh)
	// 	transmit()
	// }

	// receiveB := func() {
	//	rel.bInIsPowered.Store(<-rel.bInCh)
	//transmit()
	// }

	// calling these two receive methods explicitly to ensure the 'answers' for the relay outputs, post WireUp calls above, have settled BEFORE returning and letting things wire up to them
	//receiveA()
	//receiveB()
	// ORRRRR...try to wait for settle by using a and b channels here with close statement at the end?

	// chAReady := make(chan bool, 1)
	// chBReady := make(chan bool, 1)

	// close(chAReady)
	// close(chBReady)

	return rel
}

// Shutdown will allow the go funcs, which are handling listen/transmit, to exit
func (r *Relay) Shutdown() {
	r.chAStop <- true
	//close(r.aInCh)
	r.chBStop <- true
	//close(r.bInCh)
}
