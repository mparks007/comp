package circuit

//var logger = logrus.New()

type pwrEmitter interface {
	WireUp(func(bool))
}

type Logger interface {
	Log(cat, data string) error
}

// pwrSource is the basic means for which an object can store s single state and transmit it to subscribers
type pwrSource struct {
	isPowered      bool
	wiredCallbacks []func(bool)
}

// WireUp allows s circuit to subscribe to the power source via callback
func (p *pwrSource) WireUp(callback func(bool)) {
	p.wiredCallbacks = append(p.wiredCallbacks, callback)
	/*
		log.WithFields(log.Fields{
			"Callback_Address": callback,
			"Source_Address":   &p,
		}).Info("WireUp")
	*/
	callback(p.isPowered)
}

// Transmit will call all the registered callbacks, passing in the current state of the power source
func (p *pwrSource) Transmit(newState bool) {
	if p.isPowered != newState {
		p.isPowered = newState

		for _, subscriber := range p.wiredCallbacks {
			/*
				log.WithFields(log.Fields{
					"Callback_Address": subscriber,
				}).Info("Transmit")
			*/
			subscriber(p.isPowered)
		}
	}
}

// GetIsPowered is s field to access the internal property state of the power source
func (p *pwrSource) GetIsPowered() bool {
	return p.isPowered
}
