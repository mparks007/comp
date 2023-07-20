package circuit

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Wire is a component connector, which will transmit between source and listeners
//	Most loop-back based compoound components will use a wire for the looping aspect.
type Wire struct {
	Input       chan Charge   // will be used to allow the wire to WireUp to an outside component, and therefore await power states from it to transmit to whatever is wired up to the wire
	outChannels []chan Charge // hold list of other components that are wired up to this one to recieve charge state changes
	hasCharge   atomic.Value  // core state flag to know of the component's charge state (allows avoiding having to constantly push the power states around)
	chStop      chan bool     // listen/transmit loop shutdown channel
	name        string        // name of component for debug purposes
}

// NewWire creates a wire
func NewWire(name string) *Wire {
	w := &Wire{}

	w.name = name
	w.Input = make(chan Charge, 1)
	w.hasCharge.Store(false)
	w.chStop = make(chan bool, 1)

	// spin up the func that will allow the wire's input to be sent as output
	go func() {
		for {
			select {
			case c := <-w.Input:
				Debug(w.name, fmt.Sprintf("Received on Channel (%v), Charge {%s}", w.Input, c.String()))
				go func(c Charge) {
					w.Transmit(c)
					c.Done()
				}(c)
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

// WireUp allows another component to subscribe to the wire (via the passed in channel) in order to be told of charge state changes
func (w *Wire) WireUp(ch chan Charge) {
	w.outChannels = append(w.outChannels, ch)

	// go ahead and transmit to the new subscriber immediately as if something just connected to the wire's potentially hot current
	wg := &sync.WaitGroup{}
	wg.Add(1)
	Debug(w.name, fmt.Sprintf("Transmitting (%t) to Channel (%v) due to WireUp", w.hasCharge.Load().(bool), ch))
	ch <- Charge{sender: w.name, state: w.hasCharge.Load().(bool), wg: wg, mu: &sync.RWMutex{}}
	wg.Wait()
}

// Transmit will push out the wire's new charge state (IF state changed) to each wired up component
func (w *Wire) Transmit(c Charge) {

	Debug(w.name, fmt.Sprintf("Transmit (%t)...maybe", c.state))

	if w.hasCharge.Load().(bool) == c.state {
		Debug(w.name, "Skipping Transmit (no state change)")
		return
	}

	Debug(w.name, fmt.Sprintf("Transmit (%t)...better chance since state did change", c.state))

	w.hasCharge.Store(c.state)

	if len(w.outChannels) == 0 {
		Debug(w.name, "Skipping Transmit (nothing wired up)")
		return
	}

	// if someone passed in a fresh Charge, must init the mutex that protects lockContexts
	if c.mu == nil {
		c.mu = &sync.RWMutex{}
	}

	// take over the passed in Charge to use as a fresh waitgroup for transmitting to listeners (the passed in 'c' was only needed for setting hasCharge just above)
	c.wg = &sync.WaitGroup{}
	c.sender = w.name

	for i, ch := range w.outChannels {
		c.wg.Add(1)
		go func(i int, ch chan Charge) {
			Debug(w.name, fmt.Sprintf("Transmitting (%t) to outChannels[%d]: (%v)", c.state, i, ch))
			ch <- c
		}(i, ch)
	}

	c.wg.Wait() // all immediate listeners must finish their OWN transmits before returning from this one
}

// RibbonCable is a convenient way to allow multiple wires to drop in as slice of chargeEmitters (for receiving/transmitting a charge)
type RibbonCable struct {
	Wires []chargeEmitter
}

// NewRibbonCable creates a slice of wires of the designated width (number of wires)
func NewRibbonCable(name string, width uint) *RibbonCable {

	rib := &RibbonCable{}

	for i := 0; uint(i) < width; i++ {
		rib.Wires = append(rib.Wires, NewWire(fmt.Sprintf("%s-Wires[%d]", name, i)))
	}

	return rib
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each wire, to exit
func (r *RibbonCable) Shutdown() {
	for _, w := range r.Wires {
		w.(*Wire).Shutdown()
	}
}

// SetInputs allows each wire to listen to an independent component that can transmit a charge
func (r *RibbonCable) SetInputs(pwrInputs ...chargeEmitter) {
	for i, pwr := range pwrInputs {
		pwr.WireUp((r.Wires[i]).(*Wire).Input)
	}
}
