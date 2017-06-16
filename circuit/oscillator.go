package circuit

import "time"

type Oscillator struct {
	stopCh chan bool
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
		for {
			select {
			case <-tick.C:
				o.Transmit(!o.GetIsPowered())
			case <-o.stopCh:
				tick.Stop()
				break
			}
		}
	}()
}

func (o *Oscillator) Stop() {
	o.stopCh <- true
}
