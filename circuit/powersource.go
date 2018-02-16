package circuit

import (
	"sync"
)

// pwrSource is the core of non-wire components which can store their own state and transmit that state to other components that have wired up to them
//	Most components embed pwrSource.
type pwrSource struct {
	outChannels []chan bool
	isPowered   bool
	chDone      chan bool
	mu          sync.Mutex // to protect isPowered
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

	// go ahead and transmit to the new subscriber immediately as if just connecting to a potentially hot current
	ch <- p.isPowered
}

// Transmit will push out the power source's new power state (IF state changed) to each wired up component
func (p *pwrSource) Transmit(newState bool) {
	p.mu.Lock()

	if p.isPowered != newState {
		p.isPowered = newState

		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?
		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?
		// WHY DO I NEED TO SYNC THESE CHANNEL PUSHES?

		wg := &sync.WaitGroup{} // will use this to ensure we finish letting all wired up components know of the state change before we move along

		for _, ch := range p.outChannels {
			wg.Add(1)
			go func(ch chan bool) {
				ch <- p.isPowered
				wg.Done()
			}(ch)
		}

		p.mu.Unlock() // wanted to explicitly unlock before the Wait ("block") since we are DONE with the locked fields at this point (is why no defer used)
		wg.Wait()

	} else {
		p.mu.Unlock() // must unlock since we may not have a state change (not using defer unlock due to the Unlock/Wait comment above)
	}
}
