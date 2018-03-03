package circuit

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// pwrSource is the core of non-wire components which can store their own state and transmit that state to other components that have wired up to them
//	Most components embed pwrSource.
type pwrSource struct {
	outChannels []chan Electron // hold list of other components that are wired up to this one to recieve power state changes
	isPowered   atomic.Value    // core state flag to know of the components current state
	chStop      chan bool       // listen/transmit loop shutdown channel
	Name        string          // name of component for debug purposes
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
	Debug(p.Name, fmt.Sprintf("Transmitting (%t) to (%v) due to WireUp", p.isPowered.Load().(bool), ch))
	ch <- Electron{Name: p.Name, powerState: p.isPowered.Load().(bool), wg: wg}
	wg.Wait()
}

// Transmit will push out the power source's new power state (IF state changed) to each wired up component
func (p *pwrSource) Transmit(newPowerState bool) {

	Debug(p.Name, fmt.Sprintf("Transmit (%t)...maybe", newPowerState))

	if p.isPowered.Load().(bool) != newPowerState {
		p.isPowered.Store(newPowerState)
		Debug(p.Name, fmt.Sprintf("Transmit (better chance of transmitting (%t) since state did change)", newPowerState))

		if len(p.outChannels) == 0 {
			Debug(p.Name, "No Transmit, nothing wired up")
		} else {
			wg := &sync.WaitGroup{}                          // will use this to ensure we finish firing off the state change to all wired up components
			e := Electron{Name: p.Name, powerState: newPowerState, wg: wg} // for now, will share the same electron object across all immediate listeners (each listener's channel receipt must call their own Done)

			for i, ch := range p.outChannels {
				wg.Add(1)
				go func(i int, ch chan Electron) {
					Debug(p.Name, fmt.Sprintf("Transmitting (%t) to outChannels[%d]: (%v)", newPowerState, i, ch))
					ch <- e
				}(i, ch)
			}

			wg.Wait()
		}
	} else {
		Debug(p.Name, "Skipping Transmit (no state change)")
	}
}
