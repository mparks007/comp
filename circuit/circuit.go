package circuit

import "sync"

type bitPublisher interface {
	Register(func(bool))
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

func (p *bitPublication) Publish(newState bool) {
	if p.isPowered != newState {
		p.isPowered = newState

		wg := &sync.WaitGroup{}
		for _, subscriber := range p.subscriberCallbacks {
			wg.Add(1)
			go func() {
				subscriber(p.isPowered)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

//func (p *bitPublication) GetState() bool {
	//return p.isPowered
//}

/*
type eightBitPublisher interface {
	Register(func([8]bool))
}

type eightBitPublication struct {
	isPowered           [8]bool
	subscriberCallbacks []func([8]bool)
}

func (p *eightBitPublication) Register(callback func([8]bool)) {
	p.subscriberCallbacks = append(p.subscriberCallbacks, callback)

	// ensure newly registered callback immediately gets current isPowered
	callback(p.isPowered)
}

func (p *eightBitPublication) Publish(newState bool) {
	if p.isPowered != newState {
		p.isPowered = newState

		for _, subscriber := range p.subscriberCallbacks {
			subscriber(p.isPowered)
		}
	}
}
*/
