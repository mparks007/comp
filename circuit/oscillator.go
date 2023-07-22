package circuit

import (
	"sync/atomic"
	"time"
)

// Oscillator is a circuit which attempts to simulate a crystal-driven frequency oscillation signal (not very exact unfortunately)
type Oscillator struct {
	active atomic.Value // to track whether the oscillator is oscillating to avoid setting the stop channel if not needed
	chargeSource
}

// NewOscillator will return a disabled oscillator, whose initial power state, once started, will be based on the passed in init value
func NewOscillator(name string, initState bool) *Oscillator {
	osc := &Oscillator{}
	osc.Init()
	osc.Name = name

	osc.hasCharge.Store(initState)
	osc.active.Store(false)

	return osc
}

// Oscillate will start the oscillation logic, transmitting power states at a rate based on the passed in hertz
func (o *Oscillator) Oscillate(hertz int) {
	o.active.Store(true)

	go func() {
		tick := time.NewTicker(time.Second / time.Duration(hertz))
		for {
			select {
			case <-tick.C:
				Debug(o.Name, "Tick")
				o.Transmit(Charge{state: !o.hasCharge.Load().(bool)})
			case <-o.chStop:
				Debug(o.Name, "Stopped")
				tick.Stop()
				o.active.Store(false)
				return
			}
		}
	}()
}

// Stop will stop the oscillation events of an active Oscillator (if it is active)
func (o *Oscillator) Stop() {
	if o.active.Load().(bool) {
		o.chStop <- true
	}
}
