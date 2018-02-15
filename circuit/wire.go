package circuit

import (
	"fmt"
	"sync"
	"time"
)

// Wire is a component connector, which will transmit with an optional pause to simulate wire length (delay)
type Wire struct {
	length      uint
	Input       chan bool
	outChannels []chan bool
	isPowered   bool
	name        string
	chDone      chan bool
	mu          sync.Mutex // for isPowered usage
}

func NewWire(length uint) *Wire {
	return NewNamedWire("", length)
}

func NewNamedWire(name string, length uint) *Wire {
	wire := &Wire{}
	wire.length = length
	wire.name = name
	wire.Input = make(chan bool, 1)
	wire.chDone = make(chan bool, 1)

	// spin up the func that will allow the wire's input to be wired up to something, then send to output as necessary
	go func() {
		for {
			select {
			case state := <-wire.Input:
				fmt.Printf("Transmit of %s, %t\n", wire.name, state)
				wire.Transmit(state)
			case <-wire.chDone:
				return
			}
		}
	}()

	return wire
}

// Quit allows any 'for' loops inside go funcs to exit
func (w *Wire) Quit() {
	w.chDone <- true
}

// WireUp allows a circuit to subscribe to the power source
func (w *Wire) WireUp(ch chan bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.outChannels = append(w.outChannels, ch)

	// go ahead and transmit to the new subscriber
	//fmt.Printf("Wire %s transmitting %t\n", w.name, w.isPowered)
	ch <- w.isPowered
}

// Transmit will push out the state of things (IF state changed) to each subscriber
func (w *Wire) Transmit(newState bool) bool {
	w.mu.Lock()
	var didTransmit = false

	if w.isPowered != newState {
		w.isPowered = newState
		didTransmit = true

		wg := &sync.WaitGroup{} // must use this to ensure we finish blasting bools out to subscribers before we just barrel along in the code

		for _, ch := range w.outChannels {
			wg.Add(1)
			go func(ch chan bool) {
				time.Sleep(time.Millisecond * time.Duration(w.length))
				//fmt.Printf("Wire %s transmitting %t\n", w.name, w.isPowered)
				ch <- w.isPowered
				wg.Done()
			}(ch)
		}

		w.mu.Unlock() // wanted to explicitly unlock before the Wait ("block") since we are DONE with the locked fields at this point (is why no defer used)
		wg.Wait()

	} else {
		w.mu.Unlock() // must unlock since we may not have a state change (not using defer unlock due to the Unlock/Wait comment above)
	}

	return didTransmit
}

// RibbonCable is a convenient way to allow multiple wires to drop in as slice of pwrEmitters (for receiving/emitting)
type RibbonCable struct {
	Wires []pwrEmitter
}

// RibbonCable creates a slice of wires of the designated width (number of wires) and length (applied to each wire)
func NewRibbonCable(width, len uint) *RibbonCable {

	rib := &RibbonCable{}

	for i := 0; uint(i) < width; i++ {
		rib.Wires = append(rib.Wires, NewWire(len))
	}

	return rib
}

// Quit allows the quitting of all the wires in the ribbon cable
func (r *RibbonCable) Quit() {
	for i, _ := range r.Wires {
		r.Wires[i].(*Wire).Quit()
	}
}

func (r *RibbonCable) SetInputs(wires []pwrEmitter) {
	for i, pwr := range wires {
		pwr.WireUp((r.Wires[i]).(*Wire).Input)
	}
}
