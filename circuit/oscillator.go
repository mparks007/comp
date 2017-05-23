package circuit

import (
	"sync/atomic"
	"time"
)

type oscillator struct {
	stopCh chan bool
	emit   atomic.Value
}

func newOscillator(init bool) *oscillator {
	o := &oscillator{}

	o.stopCh = make(chan bool)
	o.emit.Store(init)

	return o
}

func (o *oscillator) Oscillate(hertz int) {

	go func() {
		t := time.NewTicker(time.Second * time.Duration(hertz) / 20)
		for {
			select {
			case <-t.C:
				b, _ := o.emit.Load().(bool)
				o.emit.Store(!b)
			case <-o.stopCh:
				t.Stop()
				break
			}
		}
	}()
}

func (o *oscillator) Stop() {
	o.stopCh <- true
}

func (o *oscillator) Emitting() bool {
	b, _ := o.emit.Load().(bool)
	return b
}
