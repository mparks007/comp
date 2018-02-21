package circuit

import (
	"sync"
	"time"
)

// IF I END UP NOT NEEDING A PAUSE CONCEPT, MAYBE MAKE WIRE EMBED PWRSOURCE AND LOSS THE DUPLICATION
// IF I END UP NOT NEEDING A PAUSE CONCEPT, MAYBE MAKE WIRE EMBED PWRSOURCE AND LOSS THE DUPLICATION
// IF I END UP NOT NEEDING A PAUSE CONCEPT, MAYBE MAKE WIRE EMBED PWRSOURCE AND LOSS THE DUPLICATION
// IF I END UP NOT NEEDING A PAUSE CONCEPT, MAYBE MAKE WIRE EMBED PWRSOURCE AND LOSS THE DUPLICATION
// IF I END UP NOT NEEDING A PAUSE CONCEPT, MAYBE MAKE WIRE EMBED PWRSOURCE AND LOSS THE DUPLICATION

// Wire is a component connector, which will transmit between source and listeners (with an optional pause to simulate wire length)
//	Most loop-back based compoound components will use a wire for the looping aspect.
type Wire struct {
	length      uint            // will pause for this many milliseconds to simulate resistance of wire due to length
	Input       chan Electron   // will be used to allow the wire to WireUp to an outside component, and therefore await power states from it to transmit to whatever is wired up to the wire
	outChannels []chan Electron // hold list of other components that are wired up to this one to recieve power state changes
	isPowered   bool            // core state flag to track the components current state
	chStop      chan bool       // listen/transmit loop shutdown channel
	mu          sync.Mutex      // to protect isPowered and outChannels
}

// NewWire creates a wire of a specified length (delay)
func NewWire(length uint) *Wire {
	wire := &Wire{}

	wire.length = length
	wire.Input = make(chan Electron, 1)
	wire.chStop = make(chan bool, 1)

	// spin up the func that will allow the wire's input to be sent as output
	go func() {
		for {
			select {
			case e := <-wire.Input:
				wire.Transmit(e.powerState)
				e.wg.Done()
			case <-wire.chStop:
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
func (w *Wire) WireUp(ch chan Electron) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.outChannels = append(w.outChannels, ch)

	// go ahead and transmit to the new subscriber immediately as if something just connected to the wire's potentially hot current
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ch <- Electron{powerState: w.isPowered, wg: wg}
	wg.Wait()
}

// Transmit will push out the wire's new power state (IF state changed) to each wired up component
func (w *Wire) Transmit(newPowerState bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isPowered != newPowerState {
		w.isPowered = newPowerState

		wg := &sync.WaitGroup{} // will use this to ensure we finish firing off the state change to all wired up components (unknown how concurrent this will actually be, but trying a bit)

		e := Electron{powerState: newPowerState, wg: wg} // for now, will share the same electron object across all listeners (though the wg.Add(1) will still allow each listener to call their own Done)

		for _, ch := range w.outChannels {
			wg.Add(1)
			go func(ch chan Electron) {
				if w.length > 0 {
					time.Sleep(time.Millisecond * time.Duration(w.length)) // simulate resistance due to "length" of wire
				}
				ch <- e
			}(ch)
		}

		wg.Wait()
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
	for i := range r.Wires {
		r.Wires[i].(*Wire).Shutdown()
	}
}

// SetInputs allows each wire to listen to an independent component that can transmit power
func (r *RibbonCable) SetInputs(pwrInputs ...pwrEmitter) {
	for i, pwr := range pwrInputs {
		pwr.WireUp((r.Wires[i]).(*Wire).Input)
	}
}
