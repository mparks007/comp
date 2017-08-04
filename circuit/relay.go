package circuit

import (
	"fmt"
	"sync"
)

type Relay struct {
	aInPowered bool
	bInPowered bool
	OpenOut    pwrSource
	ClosedOut  pwrSource
	chAIn      chan bool
	chBIn      chan bool
	mu         sync.Mutex
}

func NewRelay(pin1, pin2 pwrEmitter) *Relay {
	rel := &Relay{}
	rel.chAIn = make(chan bool, 1)
	rel.chBIn = make(chan bool, 1)

	rel.UpdatePin(1, pin1)
	rel.UpdatePin(2, pin2)

	// doing aIn and bIn go funcs independently since power could be changing either one at the "same" time

	go func() {
		for {
			aState := <-rel.chAIn

			rel.mu.Lock()
			rel.aInPowered = aState
			rel.transmit()
			rel.mu.Unlock()
		}
	}()

	go func() {
		for {
			bState := <-rel.chBIn

			rel.mu.Lock()
			rel.bInPowered = bState
			rel.transmit()
			rel.mu.Unlock()
		}
	}()

	return rel
}

// IS THIS NECESSARY!!!!!!!!!!!!???????? after the wire concept gets added?  Can't wire be the only thing needing an update?
// IS THIS NECESSARY!!!!!!!!!!!!???????? after the wire concept gets added?  Can't wire be the only thing needing an update?
// IS THIS NECESSARY!!!!!!!!!!!!???????? after the wire concept gets added?  Can't wire be the only thing needing an update?
func (r *Relay) UpdatePin(pinNum int, pin pwrEmitter) {
	if pinNum < 1 || pinNum > 2 {
		panic(fmt.Sprintf("Invalid relay pin number.  Relays have two pins and the requested pin was (%d)", pinNum))
	}

	if pinNum == 1 {
		if pin != nil {
			fmt.Printf("Relay pin1 Wireup: %v\n", pin)
			pin.WireUp(r.chAIn)
		}
	} else {
		if pin != nil {
			fmt.Printf("Relay pin2 Wireup: %v\n", pin)
			pin.WireUp(r.chBIn)
		}
	}
}

// transmit will send out the results of the relay to its subscribers
func (r *Relay) transmit() {

	// TODO put these in go funcs to make more parallel?

	r.OpenOut.Transmit(r.aInPowered && !r.bInPowered)
	r.ClosedOut.Transmit(r.aInPowered && r.bInPowered)
}
