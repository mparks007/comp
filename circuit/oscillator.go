package circuit

import (
	"sync"
	"time"
)

type Oscillator struct {
	stopCh chan bool
	mu     sync.Mutex
	active bool
	pwrSource
}

func NewOscillator(init bool) *Oscillator {
	osc := &Oscillator{}

	osc.stopCh = make(chan bool)
	osc.isPowered = init

	return osc
}

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
				break
			}
		}
	}()
}

func (o *Oscillator) Stop() {
	o.mu.Lock()
	active := o.active
	o.mu.Unlock()

	if active {
		o.stopCh <- true
	}
}
