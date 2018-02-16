package circuit

import (
	"sync"
	"time"
)

// Wire is a component connector, which will transmit between source and listeners (with an optional pause to simulate wire length)
//	Most loop-back based compoound components will use a wire for the looping aspect.
type Wire struct {
	length      uint
	Input       chan bool
	outChannels []chan bool
	isPowered   bool
	chDone      chan bool
	mu          sync.Mutex // to protect isPowered
}

// NewWire creates a wire of a specified length (delay)
func NewWire(length uint) *Wire {
	wire := &Wire{}
	wire.length = length
	wire.Input = make(chan bool, 1)
	wire.chDone = make(chan bool, 1)

	// spin up the func that will allow the wire's input to be sent as output
	go func() {
		for {
			select {
			case state := <-wire.Input:
				wire.Transmit(state)
			case <-wire.chDone:
				return
			}
		}
	}()

	return wire
}

// Shutdown will allow the go func, which is handling listen/transmit, to exit
func (w *Wire) Shutdown() {
	w.chDone <- true
}

// WireUp allows another component to subscribe to the wire (via the passed in channel) in order to be told of power state changes
func (w *Wire) WireUp(ch chan bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.outChannels = append(w.outChannels, ch)

	// go ahead and transmit to the new subscriber immediately as if just connecting to a potentially hot current
	ch <- w.isPowered
}

// Transmit will push out the wire's new power state (IF state changed) to each wired up component
func (w *Wire) Transmit(newState bool) {
	w.mu.Lock()

	if w.isPowered != newState {
		w.isPowered = newState

		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?
		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?
		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?

		wg := &sync.WaitGroup{} // will use this to ensure we finish letting all wired up components know of the state change before we move along

		for _, ch := range w.outChannels {
			wg.Add(1)
			go func(ch chan bool) {
				time.Sleep(time.Millisecond * time.Duration(w.length))
				ch <- w.isPowered
				wg.Done()
			}(ch)
		}

		w.mu.Unlock() // wanted to explicitly unlock before the Wait ("block") since we are DONE with the locked fields at this point (is why no defer used)
		wg.Wait()

	} else {
		w.mu.Unlock() // must unlock since we may not have a state change (not using defer unlock due to the Unlock/Wait comment above)
	}
}

// RibbonCable is a convenient way to allow multiple wires to drop in as slice of pwrEmitters (for receiving/transmitting power)
type RibbonCable struct {
	Wires []pwrEmitter
}

// NewRibbonCable creates a slice of wires of the designated width (number of wires) and length (delay applied to each wire)
func NewRibbonCable(width, len uint) *RibbonCable {

	rib := &RibbonCable{}

	for i := 0; uint(i) < width; i++ {
		rib.Wires = append(rib.Wires, NewWire(len))
	}

	return rib
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each wire, to exit
func (r *RibbonCable) Shutdown() {
	for i, _ := range r.Wires {
		r.Wires[i].(*Wire).Shutdown()
	}
}

// SetInputs allows each wire to listen to an independent component that can transmit power
func (r *RibbonCable) SetInputs(wires []pwrEmitter) {
	for i, pwr := range wires {
		pwr.WireUp((r.Wires[i]).(*Wire).Input)
	}
}
