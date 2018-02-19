package circuit

import (
	"fmt"
	"sync"
)

type Electron struct {
	state bool
	wg    *sync.WaitGroup
}

// pwrSource is the core of non-wire components which can store their own state and transmit that state to other components that have wired up to them
//	Most components embed pwrSource.
type pwrSource struct {
	outChannels   []chan bool     // hold list of other components that are wired up to this one
	outChannels2  []chan Electron // hold list of other components that are wired up to this one
	isPowered     bool            // core state flag to know of the components current state
	chStop        chan bool       // listen/transmit loop shutdown channel
	chTransmitted chan bool       // to track when the transmit loop has finished sending state to all subscribers
	mu            sync.Mutex      // to protect isPowered and outChannels
}

// Init will do initialization code for all pwrSource-based objects
func (p *pwrSource) Init() {
	p.chStop = make(chan bool, 1)
	p.chTransmitted = make(chan bool, 1)
}

// WireUp allows another component to subscribe to the power source (via the passed in channel) in order to be told of power state changes
func (p *pwrSource) WireUp(ch chan bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.outChannels = append(p.outChannels, ch)

	// go ahead and transmit to the new subscriber immediately as if something just connected to the pwrsource's potentially hot current
	ch <- p.isPowered
}

// Transmit will push out the power source's new power state (IF state changed) to each wired up component
func (p *pwrSource) Transmit(newState bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isPowered != newState {
		p.isPowered = newState

		wg := &sync.WaitGroup{} // will use this to ensure we finish firing off the state change to all wired up components (unknown how concurrent this will actually be, but trying a bit)
		e := Electron{state: newState, wg: wg}

		for _, ch := range p.outChannels2 {
			wg.Add(1)
			go func(ch chan bool) {
				ch <- e
				//wg.Done()  // move to deepest point where the ch is know to no longer cause anything
			}(ch)
		}

		wg.Wait()
	}
	fmt.Println("p.chTransmitted <- true")
	p.chTransmitted <- true
}
