package circuit

import (
	"fmt"
	"sync"
	"time"
)

// Wire is a component connector, which will transmit between source and listeners (with an optional pause to simulate wire length)
//	Most loop-back based compoound components will use a wire for the looping aspect.
type Wire struct {
	length        uint        // will pause for this many milliseconds to simulate resistance of wire due to length
	Input         chan bool   // will be used to allow the wire to WireUp to an outside component, and therefore await power states from it to transmit to whatever is wired up to the wire
	outChannels   []chan bool // hold list of other components that are wired up to this one
	isPowered     bool        // core state flag to track the components current state
	chStop        chan bool   // listen/transmit loop shutdown channel
	chTransmitted chan bool   // to track when the transmit loop has finished sending state to all subscribers
	mu            sync.Mutex  // to protect isPowered and outChannels
}

// NewWire creates a wire of a specified length (delay)
func NewWire(length uint) *Wire {
	wire := &Wire{}

	wire.length = length
	wire.Input = make(chan bool, 1)
	wire.chStop = make(chan bool, 1)
	wire.chTransmitted = make(chan bool, 1)

	// spin up the func that will allow the wire's input to be sent as output
	go func() {
		for {
			select {
			case state := <-wire.Input:
				wire.Transmit(state)
			case <-wire.chStop:
				fmt.Println("DEBUG: Bailing from Wire go func loop")
				return
			}
		}
	}()

	return wire
}

// Shutdown will allow the go func, which is handling listen/transmit, to exit
func (w *Wire) Shutdown() {
	w.chStop <- true
}

// WireUp allows another component to subscribe to the wire (via the passed in channel) in order to be told of power state changes
func (w *Wire) WireUp(ch chan bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.outChannels = append(w.outChannels, ch)

	// go ahead and transmit to the new subscriber immediately as if something just connected to the wire's potentially hot current
	ch <- w.isPowered
}

// Transmit will push out the wire's new power state (IF state changed) to each wired up component
func (w *Wire) Transmit(newState bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isPowered != newState {
		w.isPowered = newState

		wg := &sync.WaitGroup{} // will use this to ensure we finish firing off the state change to all wired up components (unknown how concurrent this will actually be, but trying a bit)

		for _, ch := range w.outChannels {
			wg.Add(1)
			go func(ch chan bool) {
				if w.length > 0 {
					time.Sleep(time.Millisecond * time.Duration(w.length)) // simulate resistance due to "length" of wire
				}
				ch <- w.isPowered
				wg.Done()
			}(ch)
		}

		wg.Wait()
	}
	fmt.Println("w.chTransmitted <- true")
	w.chTransmitted <- true
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
	fmt.Println("DEBUG: About to shutdown wires in ribbon cable")
	for i := range r.Wires {
		r.Wires[i].(*Wire).Shutdown()
	}
}

// SetInputs allows each wire to listen to an independent component that can transmit power
func (r *RibbonCable) SetInputs(pwrInputs ...pwrEmitter) {
	for i, pwr := range pwrInputs {
		pwr.WireUp((r.Wires[i]).(*Wire).Input)

		// wait for each wire to transmit to any subscribers (or skip due to no state change)
		<-(r.Wires[i]).(*Wire).chTransmitted
	}
}
