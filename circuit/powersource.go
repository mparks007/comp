package circuit

import (
	"sync"
	//"time"
)

// pwrSource is the core of non-wire components which can store their own state and transmit that state to other components that have wired up to them
//	Most components embed pwrSource.
type pwrSource struct {
	outChannels []chan bool // hold list of other components that are wired up to this one
	isPowered   bool        // core state flag to know of the components current state
	chDone      chan bool   // shutdown channel
	mu          sync.Mutex  // to protect isPowered and outChannels
}

// Init will do initialization code for all pwrSource-based objects
func (p *pwrSource) Init() {
	p.chDone = make(chan bool, 1)
}

// WireUp allows another component to subscribe to the power source (via the passed in channel) in order to be told of power state changes
func (p *pwrSource) WireUp(ch chan bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.outChannels = append(p.outChannels, ch)

	// go ahead and transmit to the new subscriber immediately as if something just connected to the pwrsource's potentially hot current
	ch <- p.isPowered
//	time.Sleep(time.Millisecond * 10)
}

// Transmit will push out the power source's new power state (IF state changed) to each wired up component
func (p *pwrSource) Transmit(newState bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isPowered != newState {
		p.isPowered = newState

		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?
		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?
		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?

		wg := &sync.WaitGroup{} // will use this to ensure we finish firing off the state change to all wired up components (unknown how concurrent this will actually be, but trying a bit)

		for _, ch := range p.outChannels {
			wg.Add(1)
			go func(ch chan bool) {
				ch <- p.isPowered
				wg.Done()
			}(ch)
		}

		wg.Wait()
	}
}
