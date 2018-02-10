package circuit

import (
	"sync"
	"time"
)

// Wire is a component connector, which will transmit with an optional pause to simulate wire length (delay)
type Wire struct {
	length      uint
	outChannels []chan bool
	isPowered   bool
	name        string
	mu          sync.Mutex // for isPowered usage
}

func NewWire(length uint) *Wire {
	return NewNamedWire("", length)
}

func NewNamedWire(name string, length uint) *Wire {
	wire := &Wire{}
	wire.length = length
	wire.name = name
	return wire
}

// WireUp allows a circuit to subscribe to the power source
func (w *Wire) WireUp(ch chan bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.outChannels = append(w.outChannels, ch)

	// go ahead and transmit to the new subscriber
	//fmt.Printf("Wire %s transmitting %t\n", w.name, w.isPowered)
	ch <- w.isPowered
}

// Transmit will push out the state of things (IF state changed) to each subscriber
func (w *Wire) Transmit(newState bool) bool {
	w.mu.Lock()
	var didTransmit = false

	if w.isPowered != newState {
		w.isPowered = newState
		didTransmit = true

		wg := &sync.WaitGroup{} // must use this to ensure we finish blasting bools out to subscribers before we just barrel along in the code

		for _, ch := range w.outChannels {
			wg.Add(1)
			go func(ch chan bool) {
				time.Sleep(time.Millisecond * time.Duration(w.length))
				//fmt.Printf("Wire %s transmitting %t\n", w.name, w.isPowered)
				ch <- w.isPowered
				wg.Done()
			}(ch)
		}

		w.mu.Unlock() // wanted to explicitly unlock before the Wait ("block") since we are DONE with the locked fields at this point (is why no defer used)
		wg.Wait()

	} else {
		w.mu.Unlock() // must unlock since we may not have a state change (not using defer unlock due to the Unlock/Wait comment above)
	}

	return didTransmit
}
