package circuit

type bitPublisher interface {
	Register(func(bool))
}

// bitPublication is the basic means for which an object can store a single state and publish it to subscribers
type bitPublication struct {
	//mu                  sync.Mutex
	isPowered           bool
	subscriberCallbacks []func(bool)
}

// Register allows an object subscribe to the publication via callback
func (p *bitPublication) Register(callback func(bool)) {
	//p.mu.Lock()
	//defer p.mu.Unlock()

	p.subscriberCallbacks = append(p.subscriberCallbacks, callback)

	// ensure newly registered callback immediately gets current state
	callback(p.isPowered)
}

// Publish will call all the registered callbacks, passing in the current state
func (p *bitPublication) Publish(newState bool) {
	//p.mu.Lock()
	//defer p.mu.Unlock()

	if p.isPowered != newState {
		p.isPowered = newState

		//wg := &sync.WaitGroup{}
		for _, subscriber := range p.subscriberCallbacks {
			//wg.Add(1)
			//go func(state bool, subFunc func(bool)) {
			//	subFunc(state)
			//	wg.Done()
			subscriber(p.isPowered)
			//	}(p.isPowered, subscriber)
		}
		//wg.Wait()
	}
}

func (p *bitPublication) GetIsPowered() bool {
	//p.mu.Lock()
	//defer p.mu.Unlock()

	return p.isPowered
}
