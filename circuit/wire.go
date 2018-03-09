package circuit

import (
	"fmt"
	"sync"
	"sync/atomic"
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
	isPowered   atomic.Value    // core state flag to track the components current state
	seqNum int64
	chStop      chan bool       // listen/transmit loop shutdown channel
	name        string          // name of component for debug purposes
}

// NewWire creates a wire of a specified length (delay)
func NewWire(name string, length uint) *Wire {
	w := &Wire{}

	w.name = name
	w.length = length
	w.Input = make(chan Electron, 1)
	w.isPowered.Store(false)
	w.seqNum = -1
	w.chStop = make(chan bool, 1)

	// spin up the func that will allow the wire's input to be sent as output
	go func() {
		for {
			select {
			case e := <-w.Input:
				Debug(w.name, fmt.Sprintf("Received (%t) from (%s) on (%v)", e.powerState, e.name, w.Input))
				go func(e Electron) {
					Debug(name, fmt.Sprintf("e.seqNum (%v)", e.seqNum))
					w.Transmit(e.powerState, e.seqNum)
					e.Done()
				}(e)
			case <-w.chStop:
				Debug(w.name, "Stopped")
				return
			}
		}
	}()

	return w
}

// Shutdown will allow the go func, which is handling listen/transmit, to exit
func (w *Wire) Shutdown() {
	w.chStop <- true
}

// WireUp allows another component to subscribe to the wire (via the passed in channel) in order to be told of power state changes
func (w *Wire) WireUp(ch chan Electron) {
	w.outChannels = append(w.outChannels, ch)

	// go ahead and transmit to the new subscriber immediately as if something just connected to the wire's potentially hot current
	wg := &sync.WaitGroup{}
	wg.Add(1)
	Debug(w.name, fmt.Sprintf("Transmitting (%t) to (%v) due to WireUp", w.isPowered.Load().(bool), ch))
	ch <- Electron{name: w.name, powerState: w.isPowered.Load().(bool), seqNum: w.seqNum, wg: wg}
	wg.Wait()
}

// Transmit will push out the wire's new power state (IF state changed) to each wired up component
func (w *Wire) Transmit(newPowerState bool, seqNum int64) {

	Debug(w.name, fmt.Sprintf("Transmit (%t)...maybe", newPowerState))

	if w.isPowered.Load().(bool) == newPowerState {
		Debug(w.name, "Skipping Transmit (no state change)")
		return
	}

	Debug(w.name, fmt.Sprintf("Transmit (%t)...better chance since state did change", newPowerState))

	w.isPowered.Store(newPowerState)

	if len(w.outChannels) == 0 {
		Debug(w.name, "No Transmit, nothing wired up")
		return
	}

	wg := &sync.WaitGroup{}                                                        // will use this to ensure all immediate listeners finish their OWN transmits before returning from this one
	e := Electron{name: w.name, powerState: newPowerState, seqNum: seqNum, wg: wg} // use common electron object for all immediate listeners (each listener's channel select must call their own Done)

	for i, ch := range w.outChannels {
		if w.length > 0 {
			time.Sleep(time.Millisecond * time.Duration(w.length)) // simulate resistance due to "length" of wire
		}

		wg.Add(1)
		go func(i int, ch chan Electron) {
			Debug(w.name, fmt.Sprintf("Transmitting (%t) to outChannels[%d]: (%v)", newPowerState, i, ch))
			ch <- e
		}(i, ch)
	}

	wg.Wait()
}

// RibbonCable is a convenient way to allow multiple wires to drop in as slice of pwrEmitters (for receiving/transmitting power)
type RibbonCable struct {
	Wires []pwrEmitter
}

// NewRibbonCable creates a slice of wires of the designated width (number of wires) and length (delay applied to each wire)
func NewRibbonCable(name string, width, len uint) *RibbonCable {

	rib := &RibbonCable{}

	for i := 0; uint(i) < width; i++ {
		rib.Wires = append(rib.Wires, NewWire(fmt.Sprintf("%s-Wires[%d]", name, i), len))
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
