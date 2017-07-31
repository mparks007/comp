package circuit

import (
	"sync"
	"time"
)

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

func (w *Wire) WireUp(ch chan bool) {
	w.outChannels = append(w.outChannels, ch)

	state := w.isPowered

	go func() {
		ch <- state
	}()
}

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
