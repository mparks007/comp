package circuit

import (
	"sync"
	"time"
)

// Oscillator is a circuit which attempts to simulate a crystal-driven frequency oscillation signal (not exact unfortunately)
type Oscillator struct {
	stopCh chan bool
	mu     sync.Mutex
	active bool
	pwrSource
}

// NewOscillator will return a disabled oscillator, whose initial power state, once started, will be based on the passed in init value
func NewOscillator(init bool) *Oscillator {
	osc := &Oscillator{}

	osc.stopCh = make(chan bool)
	osc.isPowered = init

	return osc
}

// Oscillate will start the oscillation logic, transmitting power states at a rate based on the passed in hertz
func (o *Oscillator) Oscillate(hertz int) {

	o.mu.Lock()
	o.active = true
	o.mu.Unlock()

	go func() {
		tick := time.NewTicker(time.Second / time.Duration(hertz))
		for {
			select {
			case <-tick.C:
				o.Transmit(!o.isPowered)
			case <-o.stopCh:
				tick.Stop()
				o.mu.Lock()
				o.active = false
				o.mu.Unlock()
				return
			}
		}
	}()
}

// Stop will stop the oscillation events of an active Oscillator
func (o *Oscillator) Stop() {
	o.mu.Lock()
	active := o.active
	o.mu.Unlock()

	if active {
		o.stopCh <- true
	}
}
