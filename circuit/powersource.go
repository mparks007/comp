package circuit

import "sync"

// pwrSource is the basic means for which a component can store its own state and transmit that state to its subscribers
type pwrSource struct {
	outChannels []chan bool
	isPowered   bool
}

// WireUp allows a circuit to subscribe to the power source
func (p *pwrSource) WireUp(ch chan bool) {
	p.outChannels = append(p.outChannels, ch)

	ch <- p.isPowered
}

func (p *pwrSource) Transmit(newState bool) {
	if p.isPowered != newState {
		p.isPowered = newState

		wg := &sync.WaitGroup{}

		for _, ch := range p.outChannels {
			wg.Add(1)
			go func(ch chan bool) {
				ch <- newState
				wg.Done()
			}(ch)
		}

		wg.Wait()
	}
}

// GetIsPowered is s field to access the internal property state of the power source
//func (p *pwrSource) GetIsPowered() bool {
//	return p.isPowered
//}
