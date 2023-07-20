package circuit

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// chargeSource is the core of non-wire components which can store their own state and transmit that state to other components that have been wired up to them
//	Most components embed chargeSource
type chargeSource struct {
	outChannels []chan Charge // hold list of other component's input channels that are wired up to this one to recieve charge state changes
	hasCharge   atomic.Value  // core state flag to know the component (which has embeded this object) charge state (allows avoiding having to constantly push the charge states around)
	chStop      chan bool     // listen/transmit loop shutdown channel
	Name        string        // name of owning component for debug purposes
}

// Init will do initialization code for all chargetSource-based objects
func (cs *chargeSource) Init() {
	cs.hasCharge.Store(false)
	cs.chStop = make(chan bool, 1)
}

// WireUp allows another component to subscribe to the charge source (via the passed in channel) in order to be told of charge state changes
func (cs *chargeSource) WireUp(ch chan Charge) {
	cs.outChannels = append(cs.outChannels, ch)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	Debug(cs.Name, fmt.Sprintf("Transmitting (%t) to Channel (%v) due to WireUp", cs.hasCharge.Load().(bool), ch))
	ch <- Charge{sender: cs.Name, state: cs.hasCharge.Load().(bool), wg: wg, mu: &sync.RWMutex{}}
	wg.Wait()
}

// Transmit will push out the charge source's new charge state (IF state changed) to each wired up component
func (cs *chargeSource) Transmit(c Charge) {

	Debug(cs.Name, fmt.Sprintf("Transmit (%t)...maybe", c.state))

	if cs.hasCharge.Load().(bool) == c.state {
		Debug(cs.Name, "Skipping Transmit (no state change)")
		return
	}

	Debug(cs.Name, fmt.Sprintf("Transmit (%t)...better chance since state did change", c.state))

	cs.hasCharge.Store(c.state)

	if len(cs.outChannels) == 0 {
		Debug(cs.Name, "Skipping Transmit (nothing wired up)")
		return
	}

	// if someone passed in a fresh Charge, must init the mutex that protects lockContexts
	if c.mu == nil {
		c.mu = &sync.RWMutex{}
	}

	// take over the passed in Charge to use as a fresh waitgroup for transmitting to listeners (the passed in 'c' was only needed for setting hasCharge just above)
	c.wg = &sync.WaitGroup{}
	c.sender = cs.Name

	for i, ch := range cs.outChannels {
		c.wg.Add(1)
		go func(i int, ch chan Charge) {
			Debug(cs.Name, fmt.Sprintf("Transmitting (%t) to outChannels[%d]: (%v)", c.state, i, ch))
			ch <- c
		}(i, ch)
	}

	c.wg.Wait() // all immediate listeners must finish their OWN transmits before returning from this one
}
