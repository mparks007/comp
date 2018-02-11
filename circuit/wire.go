package circuit

import (
	"fmt"
	"sync"
	"time"
)

// Wire is a component connector, which will transmit with an optional pause to simulate wire length (delay)
type Wire struct {
	length      uint
	Input       chan bool
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
	wire.Input = make(chan bool, 1)

	// spin up the func that will allow the wire's input to be wired up to something, then send to output as necessary
	go func() {
		for {
			state := <-wire.Input
			fmt.Printf("Transmit of %s, %t\n", wire.name, state)
			wire.Transmit(state)
		}
	}()

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

// why do NWireBank if can just make a slice or array of wires (wires wont need any conversion of "000" to states like switchbank helped with)
// why do NWireBank if can just make a slice or array of wires (wires wont need any conversion of "000" to states like switchbank helped with)
// why do NWireBank if can just make a slice or array of wires (wires wont need any conversion of "000" to states like switchbank helped with)

/*
// NWireBank is a convenient way to get allow multiple wires to drop in as an array of emitters
type NWireBank struct {
	Wires []pwrEmitter
}

// NewNWireBank takes a string of 0/1s and creates a variable length list of Switch structs initialized based on their off/on-ness
func NewNSwitchBank(bits string) (*NSwitchBank, error) {

	match, err := regexp.MatchString("^[01]+$", bits)
	if err != nil {
		return nil, err
	}
	if !match {
		err = fmt.Errorf("Input not in binary format: \"%s\"", bits)
		return nil, err
	}

	sb := &NSwitchBank{}

	for _, bit := range bits {
		sb.Switches = append(sb.Switches, NewSwitch(bit == '1'))
	}

	return sb, nil
}
*/
