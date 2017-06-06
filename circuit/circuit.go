package circuit

import "sync"

type bitPublisher interface {
	Register(func(bool))
	//DeRegister(func(bool))
}

type bitPublication struct {
	isPowered           bool
	subscriberCallbacks []func(bool)
}

func (p *bitPublication) Register(callback func(bool)) {
	p.subscriberCallbacks = append(p.subscriberCallbacks, callback)

	// ensure newly registered callback immediately gets current state
	callback(p.isPowered)
}

/*
func (p *bitPublication) DeRegister(callback func(bool)) {

	for i := 0; i < len(p.subscriberCallbacks); i++ {
		if &p.subscriberCallbacks[i] == &callback {
			p.subscriberCallbacks[i](false)
			p.subscriberCallbacks = append(p.subscriberCallbacks[:i], p.subscriberCallbacks[i+1:]...)
			i -= 1 // -1 as the slice just got shorter
		}
	}
}
*/
func (p *bitPublication) Publish(newState bool) {
	if p.isPowered != newState {
		p.isPowered = newState

		wg := &sync.WaitGroup{}
		for _, subscriber := range p.subscriberCallbacks {
			wg.Add(1)
			go func(subFunc func(bool)) {
				subFunc(p.isPowered)
				wg.Done()
			}(subscriber)
		}
		wg.Wait()
	}
}
