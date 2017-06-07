package circuit

import "time"

type Oscillator struct {
	stopCh chan bool
	bitPublication
}

func NewOscillator(init bool) *Oscillator {
	o := &Oscillator{}

	o.stopCh = make(chan bool)
	o.isPowered = init

	return o
}

func (o *Oscillator) Oscillate(hertz int) {

	go func() {
		t := time.NewTicker(time.Second / time.Duration(hertz))
		for {
			select {
			case <-t.C:
				o.Publish(!o.GetIsPowered())
			case <-o.stopCh:
				t.Stop()
				break
			}
		}
	}()
}

func (o *Oscillator) Stop() {
	o.stopCh <- true
}
