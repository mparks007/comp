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
}

func NewWire(length uint) *Wire {
	wire := &Wire{}
	wire.length = length
	return wire
}

// WireUp allows a circuit to subscribe to the power source
func (w *Wire) WireUp(ch chan bool) {
	w.outChannels = append(w.outChannels, ch)

	// go ahead and transmit to the new subscriber
	ch <- w.isPowered
}

// Transmit will push out the state of things (IF state changed) to each subscriber
func (w *Wire) Transmit(newState bool) {
	if w.isPowered != newState {
		w.isPowered = newState

		wg := &sync.WaitGroup{}

		for _, ch := range w.outChannels {
			wg.Add(1)
			go func() {
				time.Sleep(time.Millisecond * time.Duration(w.length))
				ch <- newState
				wg.Done()
			}()
		}

		wg.Wait()
	}
}
