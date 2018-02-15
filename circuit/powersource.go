package circuit

import (
	"sync"
)

// pwrSource is the basic means for which a component can store its own state and transmit that state to its subscribers
type pwrSource struct {
	outChannels []chan bool
	isPowered   bool
	name        string
	chDone      chan bool
	mu          sync.Mutex // for isPowered usage
}

// Init will do initialization code for all pwrSource-based objects
func (p *pwrSource) Init() {
	p.chDone = make(chan bool, 1)
}

// WireUp allows a circuit to subscribe to the power source
func (p *pwrSource) WireUp(ch chan bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.outChannels = append(p.outChannels, ch)

	// go ahead and transmit to the new subscriber
	ch <- p.isPowered
}

// Transmit will push out the state of things (IF state changed) to each subscriber
func (p *pwrSource) Transmit(newState bool) bool {
	p.mu.Lock()
	var didTransmit = false

	if p.isPowered != newState {
		p.isPowered = newState
		didTransmit = true

		wg := &sync.WaitGroup{} // must use this to ensure we finish blasting bools out to subscribers before we just barrel along in the code

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

	return didTransmit
}

// Quit allows for looped go funcs to exit
func (p *pwrSource) Quit() {
	p.chDone <- true
}
