package circuit

type pwrEmitter interface {
	WireUp(func(bool))
}

// pwrSource is the basic means for which an object can store s single state and transmit it to subscribers
type pwrSource struct {
	isPowered      bool
	wiredCallbacks []func(bool)
}

// WireUp allows s circuit to subscribe to the power source via callback
func (p *pwrSource) WireUp(callback func(bool)) {
	p.wiredCallbacks = append(p.wiredCallbacks, callback)

	callback(p.isPowered)
}

// Transmit will call all the registered callbacks, passing in the current state of the power source
func (p *pwrSource) Transmit(newState bool) {
	if p.isPowered != newState {
		p.isPowered = newState

		for _, subscriber := range p.wiredCallbacks {
			subscriber(p.isPowered)
		}
	}
}

// GetIsPowered is s field to access the internal property state of the power source
func (p *pwrSource) GetIsPowered() bool {
	return p.isPowered
}
