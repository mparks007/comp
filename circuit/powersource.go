package circuit

import (
	"sync"
	"sync/atomic"
)

// pwrSource is the core of non-wire components which can store their own state and transmit that state to other components that have wired up to them
//	Most components embed pwrSource.
type pwrSource struct {
	outChannels []chan Electron // hold list of other components that are wired up to this one to recieve power state changes
	isPowered   atomic.Value    // core state flag to know of the components current state
	chStop      chan bool       // listen/transmit loop shutdown channel
}

// Init will do initialization code for all pwrSource-based objects
func (p *pwrSource) Init() {
	p.isPowered.Store(false)
	p.chStop = make(chan bool, 1)
}

// WireUp allows another component to subscribe to the power source (via the passed in channel) in order to be told of power state changes
func (p *pwrSource) WireUp(ch chan Electron) {

	p.outChannels = append(p.outChannels, ch)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	ch <- Electron{powerState: p.isPowered.Load().(bool), wg: wg}
	wg.Wait()
}

// Transmit will push out the power source's new power state (IF state changed) to each wired up component
func (p *pwrSource) Transmit(newPowerState bool) {

	if p.isPowered.Load().(bool) != newPowerState {
		p.isPowered.Store(newPowerState)

		wg := &sync.WaitGroup{} // will use this to ensure we finish firing off the state change to all wired up components (unknown how concurrent this will actually be, but trying a bit)

		e := Electron{powerState: newPowerState, wg: wg} // for now, will share the same electron object across all listeners (though the wg.Add(1) will still allow each listener to call their own Done)

		for _, ch := range p.outChannels {
			wg.Add(1)
			go func(ch chan Electron) {
				ch <- e
			}(ch)
		}

		wg.Wait()
	}
}
