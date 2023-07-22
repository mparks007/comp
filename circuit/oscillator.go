package circuit

import (
	"sync/atomic"
	"time"
)

// Oscillator is a circuit which attempts to simulate a crystal-driven frequency oscillation signal (not very exact unfortunately)
type Oscillator struct {
	isOscillating atomic.Value // to track whether the oscillator is oscillating to avoid setting the stop channel if not needed
	chargeSource
}

// NewOscillator will return a disabled oscillator, whose initial power state, once started, will be based on the passed in init value
func NewOscillator(name string, initState bool) *Oscillator {
	osc := &Oscillator{}
	osc.Init()
	osc.Name = name

	osc.hasCharge.Store(initState)
	osc.isOscillating.Store(false)

	return osc
}

// Oscillate will start the oscillation logic, transmitting power states at a rate based on the passed in hertz
func (o *Oscillator) Oscillate(hertz int) {
	o.isOscillating.Store(true)

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
				o.isOscillating.Store(false)
				return
			}
		}
	}()
}

// Stop will stop the oscillation events of an active Oscillator (if it is active)
func (o *Oscillator) Stop() {
	if o.isOscillating.Load().(bool) {
		o.chStop <- true
	}
}

// Oscillator2 is a circuit which attempts to simulate a crystal-driven frequency oscillation by looping it's own signal
// TODO: Still dabbling in how to do this without a deadlock
// type Oscillator2 struct {
// 	isOscillating atomic.Value // to track whether the oscillator is oscillating to avoid setting the stop channel if not needed
// 	oSwitch       *Switch
// 	loopWire      *Wire
// 	inverter      *Inverter
// 	chargeSource
// }

// // NewOscillator2 will return a disabled oscillator, whose initial power state, once started, will be based on the passed in init value
// func NewOscillator2(name string, initState bool) *Oscillator2 {
// 	osc := &Oscillator2{}
// 	osc.Init()
// 	osc.Name = name

// 	osc.hasCharge.Store(initState)
// 	osc.isOscillating.Store(false)

// 	osc.oSwitch = NewSwitch(fmt.Sprintf("%s-Switch", name), false)
// 	osc.loopWire = NewWire(fmt.Sprintf("%s-Wire", name))
// 	osc.inverter = NewInverter(fmt.Sprintf("%s-Inverter", name), osc.loopWire)

// 	chState := make(chan Charge, 1)
// 	go func() {
// 		for {
// 			select {
// 			case c := <-chState:
// 				Debug(name, fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
// 				osc.Transmit(Charge{state: !osc.hasCharge.Load().(bool)})
// 				c.Done()
// 			case <-osc.chStop:
// 				Debug(name, "Stopped")
// 				return
// 			}
// 		}
// 	}()

// 	osc.inverter.WireUp(osc.loopWire.Input)
// 	osc.oSwitch.WireUp(chState)

// 	return osc
// }

// // Oscillate will start the oscillation logic, transmitting power states at a rate based on the passed in hertz
// func (o *Oscillator2) Oscillate(hertz int) {
// 	o.isOscillating.Store(true)
// 	o.oSwitch.Set(true)

// 	// go func() {
// 	// 	tick := time.NewTicker(time.Second / time.Duration(hertz))
// 	// 	for {
// 	// 		select {
// 	// 		case <-tick.C:
// 	// 			Debug(o.Name, "Tick")
// 	// 			o.Transmit(Charge{state: !o.hasCharge.Load().(bool)})
// 	// 		case <-o.chStop:
// 	// 			Debug(o.Name, "Stopped")
// 	// 			tick.Stop()
// 	// 			o.isOscillating.Store(false)
// 	// 			return
// 	// 		}
// 	// 	}
// 	// }()
// }

// // Stop will stop the oscillation events of an active Oscillator (if it is active)
// func (o *Oscillator2) Stop() {
// 	if o.isOscillating.Load().(bool) {
// 		o.chStop <- true
// 		o.oSwitch.Set(false)
// 	}
// }
