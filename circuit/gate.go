package circuit

import (
	"fmt"
	"sync/atomic"
)

// ANDGate is a standard AND logic gate
//	Wired like a NOR gate, but each relay is chained via the CLOSED out
//
// Truth Table
// in in out
// 0  0   0
// 1  0   0
// 0  1   0
// 1  1   1
type ANDGate struct {
	relays    []*Relay // two or more relays to control the final gate state answer
	pwrSource          // gate gains all that is pwrSource too
}

func NewANDGate(pins ...pwrEmitter) *ANDGate {
	return NewNamedANDGate("?", pins...)
}

// NewANDGate will return an AND gate whose inputs are set by the passed in pins
func NewNamedANDGate(name string, pins ...pwrEmitter) *ANDGate {
	gate := &ANDGate{}
	gate.Init()
	gate.Name = name

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewNamedRelay(fmt.Sprintf("%s-Relay%d", name, i), NewBattery(true), pin))
		} else {
			gate.relays = append(gate.relays, NewNamedRelay(fmt.Sprintf("%s-Relay%d", name, i), &gate.relays[i-1].ClosedOut, pin))
		}
	}

	chState := make(chan Electron, 1)
	go func() {
		for {
			select {
			case e := <-chState:
				// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, would not be blocked by the select/case
				go func() {
					Debug(fmt.Sprintf("[%s]: Received (%t) from (%s) on (%v)", name, e.powerState, e.Name, chState))
					gate.Transmit(e.powerState)
					e.wg.Done()
				}()
			case <-gate.chStop:
				return
			}
		}
	}()

	// for an AND, the last relay in the chain is the final answer (from CLOSED out)
	gate.relays[len(pins)-1].ClosedOut.WireUp(chState)

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each relay and the gate itself, to exit
func (g *ANDGate) Shutdown() {
	for i := range g.relays {
		g.relays[i].Shutdown()
	}
	g.chStop <- true
}

// ORGate is a standard OR logic gate.
//	Wired like a NAND gate, but wired up to the CLOSED outs of each relay.
//
// Truth Table
// in in out
// 0  0   0
// 1  0   1
// 0  1   1
// 1  1   1
type ORGate struct {
	relays    []*Relay    // two or more relays to control the final gate state answer
	chStops   []chan bool // need to have a go func per relay and a Stop (from this slice) for each of those go func for loops
	pwrSource             // gate gains all that is pwrSource too
}

func NewORGate(pins ...pwrEmitter) *ORGate {
	return NewNamedORGate("?", pins...)
}

// NewORGate will return an OR gate whose inputs are set by the passed in pins
func NewNamedORGate(name string, pins ...pwrEmitter) *ORGate {
	gate := &ORGate{}
	gate.Init()
	gate.Name = name

	// build a relay and associated listen/transmit func to deal with each input pin
	gots := make([]atomic.Value, len(pins))
	var chStates []chan Electron
	for i, pin := range pins {
		gate.relays = append(gate.relays, NewNamedRelay(fmt.Sprintf("%s-Relay%d", name, i), NewBattery(true), pin))

		chStates = append(chStates, make(chan Electron, 1))
		gate.chStops = append(gate.chStops, make(chan bool, 1))
		go func(chState chan Electron, chStop chan bool, index int) {
			for {
				select {
				case e := <-chState:
					Debug(fmt.Sprintf("[%s]: Relay[%d] received (%t) from (%s) on (%v)", name, index, e.powerState, e.Name, chState))
					gots[index].Store(e.powerState)

					var answer bool
					// if already found a true, no need to check the other relays
					if e.powerState {
						answer = true
					} else {
						answer = false
						for g := range gots {
							// if found ANY relay as powered at ClosedOut (see WireUp later), flag and bail, the OR gate is powered (see truth table)
							if gots[g].Load() != nil && gots[g].Load().(bool) {
								answer = true
								break
							}
						}
					}
					Debug(fmt.Sprintf("[%s]: Final answer to transmit (%t)", name, answer))
					// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, would not be blocked by the select/case
					go func() {
						gate.Transmit(answer)
						e.wg.Done()
					}()
				case <-chStop:
					return
				}
			}
		}(chStates[i], gate.chStops[i], i)

		// for an OR, every relay can trigger state (from CLOSED out)
		gate.relays[i].ClosedOut.WireUp(chStates[i])
	}

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each relay and the gate itself, to exit
func (g *ORGate) Shutdown() {
	for i := range g.relays {
		g.relays[i].Shutdown()
		g.chStops[i] <- true
	}
}

// NANDGate is a standard NAND (Not-AND) logic gate.
//	Wired like an OR gate, but wired up to the OPEN outs of each relay.
//
// Truth Table
// in in out
// 0  0   1
// 1  0   1
// 0  1   1
// 1  1   0
type NANDGate struct {
	relays    []*Relay    // two or more relays to control the final gate state answer
	chStops   []chan bool // need to have a go func per relay and a Stop (from this slice) for each of those go func for loops
	pwrSource             // gate gains all that is pwrSource too
}

func NewNANDGate(pins ...pwrEmitter) *NANDGate {
	return NewNamedNANDGate("?", pins...)
}

// NewNANDGate will return a NAND gate whose inputs are set by the passed in pins
func NewNamedNANDGate(name string, pins ...pwrEmitter) *NANDGate {
	gate := &NANDGate{}
	gate.Init()
	gate.Name = name

	// build a relay and associated listen/transmit func to deal with each input pin
	gots := make([]atomic.Value, len(pins))
	var chStates []chan Electron
	for i, pin := range pins {
		gate.relays = append(gate.relays, NewNamedRelay(fmt.Sprintf("%s-Relay%d", name, i), NewBattery(true), pin))

		chStates = append(chStates, make(chan Electron, 1))
		gate.chStops = append(gate.chStops, make(chan bool, 1))
		go func(chState chan Electron, chStop chan bool, index int) {
			for {
				select {
				case e := <-chState:
					Debug(fmt.Sprintf("[%s]: Relay[%d] received (%t) from (%s) on (%v)", name, index, e.powerState, e.Name, chState))
					gots[index].Store(e.powerState)

					var answer bool
					// if already found a true, no need to check the other relays
					if e.powerState {
						answer = true
					} else {
						answer = false
						for g := range gots {
							// if found ANY relay as powered at OpenOut (see WireUp later), flag and bail, the NAND gate is powered (see truth table)
							if gots[g].Load() != nil && gots[g].Load().(bool) {
								answer = true
								break
							}
						}
					}
					Debug(fmt.Sprintf("[%s]: Final answer to transmit (%t)", name, answer))
					// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, would not be blocked by the select/case
					go func() {
						gate.Transmit(answer)
						e.wg.Done()
					}()
				case <-chStop:
					return
				}
			}
		}(chStates[i], gate.chStops[i], i)

		// for a NAND, every relay can trigger state (from OPEN out)
		gate.relays[i].OpenOut.WireUp(chStates[i])
	}

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each relay and the gate itself, to exit
func (g *NANDGate) Shutdown() {
	for i := range g.relays {
		g.relays[i].Shutdown()
		g.chStops[i] <- true
	}
}

// NORGate is a standard NOR (Not-OR) logic gate.
//	Wired like an AND gate, but each relay is chained via the OPEN out
//
// Truth Table
// in in out
// 0  0   1
// 1  0   0
// 0  1   0
// 1  1   0
type NORGate struct {
	relays    []*Relay // two or more relays to control the final gate state answer
	pwrSource          // gate gains all that is pwrSource too
}

// NewNORGate will return a NOR gate whose inputs are set by the passed in pins
func NewNORGate(pins ...pwrEmitter) *NORGate {
	gate := &NORGate{}
	gate.Init()

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewRelay(NewBattery(true), pin))
		} else {
			gate.relays = append(gate.relays, NewRelay(&gate.relays[i-1].OpenOut, pin))
		}
	}

	chState := make(chan Electron, 1)
	go func() {
		for {
			select {
			case e := <-chState:
				// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, would not be blocked by the select/case
				go func() {
					gate.Transmit(e.powerState)
					e.wg.Done()
				}()
			case <-gate.chStop:
				return
			}
		}
	}()

	// for a NOR, the last relay in the chain is the final answer (from OPEN out)
	gate.relays[len(pins)-1].OpenOut.WireUp(chState)

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each relay and the gate itself, to exit
func (g *NORGate) Shutdown() {
	for i := range g.relays {
		g.relays[i].Shutdown()
	}
	g.chStop <- true
}

// XORGate is a standard XOR (Exclusive-OR) logic gate.
//
// Truth Table
// in in out
// 0  0   0
// 1  0   1
// 0  1   1
// 1  1   0
type XORGate struct {
	orGate    *ORGate   // standard OR Gate used to build a basic XOR Gate
	nandGate  *NANDGate // standard NAND Gate used to build a basic XOR Gate
	andGate   *ANDGate  // standard AND Gate used to build a basic XOR Gate
	pwrSource           // gate gains all that is pwrSource too
}

func NewXORGate(pin1, pin2 pwrEmitter) *XORGate {
	return NewNamedXORGate("?", pin1, pin2)
}

// NewXORGate will return an XOR gate whose inputs are set by the passed in pins
func NewNamedXORGate(name string, pin1, pin2 pwrEmitter) *XORGate {
	gate := &XORGate{}
	gate.Init()
	gate.Name = name

	gate.orGate = NewNamedORGate(fmt.Sprintf("%s-ORGate", name), pin1, pin2)
	gate.nandGate = NewNamedNANDGate(fmt.Sprintf("%s-NANDGate", name), pin1, pin2)
	gate.andGate = NewNamedANDGate(fmt.Sprintf("%s-ANDGate", name), gate.orGate, gate.nandGate)

	chState := make(chan Electron, 1)
	go func() {
		for {
			select {
			case e := <-chState:
				// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, would not be blocked by the select/case
				go func() {
					Debug(fmt.Sprintf("[%s]: Received (%t) from (%s) on (%v)", name, e.powerState, e.Name, chState))
					gate.Transmit(e.powerState)
					e.wg.Done()
				}()
			case <-gate.chStop:
				return
			}
		}
	}()

	// for an XOR, the state of the shared AND Gate is the answer
	gate.andGate.WireUp(chState)

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-gate and the gate itself, to exit
func (g *XORGate) Shutdown() {
	g.andGate.Shutdown()
	g.nandGate.Shutdown()
	g.orGate.Shutdown()
	g.chStop <- true
}

// XNORGate is a standard XNOR (Exclusive-Not-OR) logic gate (aka equivalence gate).
// 	The approach to the circuit is simplified by just using an Inverter on an XOR gate.
//
// Truth Table
// in in out
// 0  0   1
// 1  0   0
// 0  1   0
// 1  1   1
type XNORGate struct {
	inverter  *Inverter // will use this to invert a basic XOR Gate to get an XNOR answer
	xorGate   *XORGate  // will start with this basic XOR Gate and invert it to make the XNOR result
	pwrSource           // gate gains all that is pwrSource too
}

// NewXNORGate will return an XNOR gate whose inputs are set by the passed in pins
func NewXNORGate(pin1, pin2 pwrEmitter) *XNORGate {
	gate := &XNORGate{}
	gate.Init()

	gate.xorGate = NewXORGate(pin1, pin2) // having to make one as a named object so it can be Shutdown later (vs. just feeding NewXORGate(pin1, pin2) into NewInverter())
	gate.inverter = NewInverter(gate.xorGate)

	chState := make(chan Electron, 1)
	go func() {
		for {
			select {
			case e := <-chState:
				// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, would not be blocked by the select/case
				go func() {
					gate.Transmit(e.powerState)
					e.wg.Done()
				}()
			case <-gate.chStop:
				return
			}
		}
	}()

	// in this approach to an XNOR (vs. building it with a combination of other gates), the Inverter owns the final answer
	gate.inverter.WireUp(chState)

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component and the gate itself, to exit
func (g *XNORGate) Shutdown() {
	g.xorGate.Shutdown()
	g.inverter.Shutdown()
	g.chStop <- true
}
