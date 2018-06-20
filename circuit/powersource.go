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
	isPowered   atomic.Value    // core state flag to know of the components current state (allows avoiding having to constantly push the power states around)
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
	Debug(p.Name, fmt.Sprintf("Transmitting (%t) to Channel (%v) due to WireUp", p.isPowered.Load().(bool), ch))
	ch <- Electron{sender: p.Name, powerState: p.isPowered.Load().(bool), wg: wg}
	wg.Wait()
}

// Transmit will push out the power source's new power state (IF state changed) to each wired up component
func (p *pwrSource) Transmit(e Electron) {

	Debug(p.Name, fmt.Sprintf("Transmit (%t)...maybe", e.powerState))

	if p.isPowered.Load().(bool) == e.powerState {
		Debug(p.Name, "Skipping Transmit (no state change)")
		return
	}

	Debug(p.Name, fmt.Sprintf("Transmit (%t)...better chance since state did change", e.powerState))

	p.isPowered.Store(e.powerState)

	if len(p.outChannels) == 0 {
		Debug(p.Name, "Skipping Transmit (nothing wired up)")
		return
	}

	// take over the passed in Electron to use as a fresh waitgroup for transmitting to listeners
	e.wg = &sync.WaitGroup{}
	e.sender = p.Name

	for i, ch := range p.outChannels {
		e.wg.Add(1)
		go func(i int, ch chan Electron) {
			Debug(p.Name, fmt.Sprintf("Transmitting (%t) to outChannels[%d]: (%v)", e.powerState, i, ch))
			ch <- e
		}(i, ch)
	}

	e.wg.Wait() // all immediate listeners must finish their OWN transmits before returning from this one
}
