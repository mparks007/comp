package circuit

import "sync"

type pwrEmitter interface {
	WireUp(func(bool))
}

// pwrSource is the basic means for which an object can store a single state and transmit it to subscribers
type pwrSource struct {
	mu             sync.Mutex
	isPowered      bool
	wiredCallbacks []func(bool)
}

// WireUp allows an object subscribe to the publication via callback
func (p *pwrSource) WireUp(callback func(bool)) {
	p.wiredCallbacks = append(p.wiredCallbacks, callback)

	callback(p.isPowered)
}

// Transmit will call all the registered callbacks, passing in the current state
func (p *pwrSource) Transmit(newState bool) {
	if p.isPowered != newState {
		p.isPowered = newState

		for _, subscriber := range p.wiredCallbacks {
			subscriber(p.isPowered)
		}
	}
}

func (p *pwrSource) GetIsPowered() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.isPowered
}
