package circuit

import (
	"fmt"
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

	go func() {
		tick := time.NewTicker(time.Second / time.Duration(hertz))
		o.mu.Lock()
		o.active = true
		o.mu.Unlock()
		for {
			select {
			case <-tick.C:
				fmt.Println("case <-tick.C")
				o.Transmit(!o.GetIsPowered())
			case <-o.stopCh:
				fmt.Println("case <-stopCh")
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
		fmt.Println("o.stopCh <- true")
		o.stopCh <- true
	}
}
