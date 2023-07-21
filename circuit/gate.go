package circuit

import (
	"fmt"
	"sync"
	"sync/atomic"

	uuid "github.com/satori/go.uuid"
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
	relays []*Relay // two or more relays to control the gates's output charge state
	chargeSource
}

// NewANDGate will return an AND gate whose input charge states are set by the passed in pins
func NewANDGate(name string, pins ...chargeEmitter) *ANDGate {
	gate := &ANDGate{}
	gate.Init()
	gate.Name = name

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewRelay(fmt.Sprintf("%s-Relays[%d]", name, i), NewChargeProvider(fmt.Sprintf("%s-Relays[%d]-pin1ChargeProvider", name, i), true), pin))
		} else {
			gate.relays = append(gate.relays, NewRelay(fmt.Sprintf("%s-Relays[%d]", name, i), &gate.relays[i-1].ClosedOut, pin))
		}
	}

	chState := make(chan Charge, 1)
	go func() {
		for {
			select {
			case c := <-chState:
				Debug(name, fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
				// putting this in a new go func() will allow any loopbacks triggered by the transmit, that need to end up feeding back into THIS gate, to not be blocked by the select/case
				go func(c Charge) {
					gate.Transmit(c)
					c.Done()
				}(c)
			case <-gate.chStop:
				Debug(name, "Stopped")
				return
			}
		}
	}()

	// for an AND, the last relay in the chain is the final output charge state (from CLOSED out)
	gate.relays[len(pins)-1].ClosedOut.WireUp(chState)

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each relay and the gate itself, to exit
func (g *ANDGate) Shutdown() {
	for _, r := range g.relays {
		r.Shutdown()
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
	relays  []*Relay    // two or more relays to control the final gate's output charge state
	chStops []chan bool // need to have a go func per relay and a Stop (from this slice) for each of those go func for loops
	chargeSource
}

// NewORGate will return an OR gate whose input charge states are set by the passed in pins
func NewORGate(name string, pins ...chargeEmitter) *ORGate {
	gate := &ORGate{}
	gate.Init()
	gate.Name = name

	mu := &sync.Mutex{} // to ensure order of inner-relay transmits since they are transmitting via go funcs (the order of relay charges states must go out in that order!)

	// to track if loopback, to avoid deadlock on mu
	var lockedContext atomic.Value
	lockedContext.Store(uuid.Must(uuid.NewV4()))

	// build a relay and associated listen/transmit funcs to deal with each input pin
	inputStates := make([]atomic.Value, len(pins))
	var chStates []chan Charge
	for i, pin := range pins {
		gate.relays = append(gate.relays, NewRelay(fmt.Sprintf("%s-Relays[%d]", name, i), NewChargeProvider(fmt.Sprintf("%s-Relays[%d]-pin1ChargeProvider", name, i), true), pin))

		chStates = append(chStates, make(chan Charge, 1))
		gate.chStops = append(gate.chStops, make(chan bool, 1))
		go func(chState chan Charge, chStop chan bool, index int) {
			for {
				select {
				case c := <-chState:
					Debug(name, fmt.Sprintf("(Relays[%d]) Received on Channel (%v), Charge {%s}", index, chState, c.String()))

					// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, to not be blocked by the select/case
					go func(c Charge) {
						// need to track if later must skip unlock() call
						ownsLock := false

						// if got here NOT due to a loopback situation, safe to lock
						if !c.HasContext(lockedContext.Load().(uuid.UUID)) {
							Debug(name, fmt.Sprintf("(Relays[%d]) Locking", index))
							mu.Lock()
							ownsLock = true

							// must register a unique lockId value on the Charge to check for potential case where the charge flow loops back through here
							lockId := uuid.Must(uuid.NewV4())
							lockedContext.Store(lockId)
							Debug(name, fmt.Sprintf("(Relays[%d]) Registering new lockContext (%v)", index, lockId))
							c.AddContext(lockId)
						} else {
							Debug(name, fmt.Sprintf("(Relays[%d]) Loopback (bypassing lock)", index))
						}

						inputStates[index].Store(c.state)

						var outputCharge bool
						// if already found a true, no need to check the other relays
						if c.state {
							outputCharge = true
						} else {
							Debug(name, fmt.Sprintf("Since (Relays[%d]) is (false), checking the other relays within the gate", index))

							outputCharge = false
							for i = 0; i < len(inputStates); i++ {
								// if found ANY relay as charged at ClosedOut (per ClosedOut WireUps at end of func), flag and bail, the OR gate output is charged (see truth table)
								if inputStates[i].Load() != nil && inputStates[i].Load().(bool) {
									Debug(name, fmt.Sprintf("(Relays[%d]) was found to be (true) so changing the gate's output charge state", i))
									outputCharge = true
									break
								} else {
									Debug(name, fmt.Sprintf("(Relays[%d]) was found to be (false) as well", i))
								}
							}
						}
						Debug(name, fmt.Sprintf("Final gate charge to transmit (%t)", outputCharge))

						c.state = outputCharge
						gate.Transmit(c)

						// finally got to transmit so done with locking stuff
						if ownsLock {
							Debug(name, fmt.Sprintf("(Relays[%d]) Unlocking", index))
							mu.Unlock()
						}
						c.Done()
					}(c)
				case <-chStop:
					Debug(name, fmt.Sprintf("(Relays[%d]) Stopped", index))
					return
				}
			}
		}(chStates[i], gate.chStops[i], i)

		// for an OR, every relay can trigger output charge state (from CLOSED out)
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
	relays  []*Relay    // two or more relays to control the final gate's output charge state
	chStops []chan bool // need to have a go func per relay and a Stop (from this slice) for each of those go func for loops
	chargeSource
}

// NewNANDGate will return a NAND gate whose input charge states are set by the passed in pins
func NewNANDGate(name string, pins ...chargeEmitter) *NANDGate {
	gate := &NANDGate{}
	gate.Init()
	gate.Name = name

	mu := &sync.Mutex{} // to ensure order of inner-relay transmits since they are transmitting via go funcs (the order of relay charge states must go out in that order!)

	// to track if loopback, to avoid deadlock on mu
	var lockedContext atomic.Value
	lockedContext.Store(uuid.Must(uuid.NewV4()))

	// build a relay and associated listen/transmit funcs to deal with each input pin
	inputStates := make([]atomic.Value, len(pins))
	var chStates []chan Charge
	for i, pin := range pins {
		gate.relays = append(gate.relays, NewRelay(fmt.Sprintf("%s-Relays[%d]", name, i), NewChargeProvider(fmt.Sprintf("%s-Relays[%d]-pin1ChargeProvider", name, i), true), pin))

		chStates = append(chStates, make(chan Charge, 1))
		gate.chStops = append(gate.chStops, make(chan bool, 1))
		go func(chState chan Charge, chStop chan bool, index int) {
			for {
				select {
				case c := <-chState:
					Debug(name, fmt.Sprintf("(Relays[%d]) Received on Channel (%v), Charge {%s}", index, chState, c.String()))

					// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, to not be blocked by the select/case
					go func(e Charge) {
						// need to track if later must skip unlock() call
						ownsLock := false

						// if got here NOT due to a loopback situation, safe to lock
						if !e.HasContext(lockedContext.Load().(uuid.UUID)) {
							Debug(name, fmt.Sprintf("(Relays[%d]) Locking", index))
							mu.Lock()
							ownsLock = true

							// must register a unique lockId value on the Charge to check for potential case where the charge flow loops back through here
							lockId := uuid.Must(uuid.NewV4())
							lockedContext.Store(lockId)
							Debug(name, fmt.Sprintf("(Relays[%d]) Registering new lockContext (%v)", index, lockId))
							e.AddContext(lockId)
						} else {
							Debug(name, fmt.Sprintf("(Relays[%d]) Loopback (bypassing lock)", index))
						}

						inputStates[index].Store(e.state)

						var outputCharge bool
						// if already found a true, no need to check the other relays
						if e.state {
							outputCharge = true
						} else {
							Debug(name, fmt.Sprintf("Since (Relays[%d]) is (false), checking the other relays within the gate", index))

							outputCharge = false
							for i = 0; i < len(inputStates); i++ {
								// if found ANY relay as charged at OpenOut (per OpenOut WireUps at end of func), flag and bail, the NAND gate is charged (see truth table)
								if inputStates[i].Load() != nil && inputStates[i].Load().(bool) {
									Debug(name, fmt.Sprintf("(Relays[%d]) was found to be (true) so changing the gate's output charge state", i))
									outputCharge = true
									break
								} else {
									Debug(name, fmt.Sprintf("(Relays[%d]) was found to be (false) as well", i))
								}
							}
						}
						Debug(name, fmt.Sprintf("Final gate charge to transmit (%t)", outputCharge))

						e.state = outputCharge
						gate.Transmit(e)

						// finally got to transmit so done with locking stuff
						if ownsLock {
							Debug(name, fmt.Sprintf("(Relays[%d]) Unlocking", index))
							mu.Unlock()
						}
						e.Done()
					}(c)
				case <-chStop:
					Debug(name, fmt.Sprintf("(Relays[%d]) Stopped", index))
					return
				}
			}
		}(chStates[i], gate.chStops[i], i)

		// for a NAND, every relay can trigger output charge state (from OPEN out)
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
	relays []*Relay // two or more relays to control the final gate's output charge state
	chargeSource
}

// NewNORGate will return a NOR gate whose inputs charge states are set by the passed in pins
func NewNORGate(name string, pins ...chargeEmitter) *NORGate {
	gate := &NORGate{}
	gate.Init()
	gate.Name = name

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewRelay(fmt.Sprintf("%s-Relays[%d]", name, i), NewChargeProvider(fmt.Sprintf("%s-Relays[%d]-pin1ChargeProvider", name, i), true), pin))
		} else {
			gate.relays = append(gate.relays, NewRelay(fmt.Sprintf("%s-Relays[%d]", name, i), &gate.relays[i-1].OpenOut, pin))
		}
	}

	chState := make(chan Charge, 1)
	go func() {
		for {
			select {
			case c := <-chState:
				Debug(name, fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
				// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, to not be blocked by the select/case
				go func(e Charge) {
					gate.Transmit(e)
					e.Done()
				}(c)
			case <-gate.chStop:
				Debug(name, "Stopped")
				return
			}
		}
	}()

	// for a NOR, the last relay in the chain is the output charge state (from OPEN out)
	gate.relays[len(pins)-1].OpenOut.WireUp(chState)

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each relay and the gate itself, to exit
func (g *NORGate) Shutdown() {
	for _, r := range g.relays {
		r.Shutdown()
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
	orGate   *ORGate   // standard OR Gate used to build a basic XOR Gate
	nandGate *NANDGate // standard NAND Gate used to build a basic XOR Gate
	andGate  *ANDGate  // standard AND Gate used to build a basic XOR Gate
	chargeSource
}

// NewXORGate will return an XOR gate whose input charge states are set by the passed in pins
func NewXORGate(name string, pin1, pin2 chargeEmitter) *XORGate {
	gate := &XORGate{}
	gate.Init()
	gate.Name = name

	gate.orGate = NewORGate(fmt.Sprintf("%s-ORGate", name), pin1, pin2)
	gate.nandGate = NewNANDGate(fmt.Sprintf("%s-NANDGate", name), pin1, pin2)
	gate.andGate = NewANDGate(fmt.Sprintf("%s-ANDGate", name), gate.orGate, gate.nandGate)

	chState := make(chan Charge, 1)
	go func() {
		for {
			select {
			case c := <-chState:
				Debug(name, fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
				// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, to not be blocked by the select/case
				go func(e Charge) {
					gate.Transmit(e)
					e.Done()
				}(c)
			case <-gate.chStop:
				Debug(name, "Stopped")
				return
			}
		}
	}()

	// for an XOR, the state of the shared AND Gate is the output charge state
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
// 	The approach to the circuit is simplified by just using an Inverter on an XOR gate (yeah, I cheated).
//
// Truth Table
// in in out
// 0  0   1
// 1  0   0
// 0  1   0
// 1  1   1
type XNORGate struct {
	inverter *Inverter // will use this to invert a basic XOR Gate to get an XNOR charge state
	xorGate  *XORGate  // will start with this basic XOR Gate and invert it to make the XNOR charge state
	chargeSource
}

// NewXNORGate will return an XNOR gate whose input charge states are set by the passed in pins
func NewXNORGate(name string, pin1, pin2 chargeEmitter) *XNORGate {
	gate := &XNORGate{}
	gate.Init()
	gate.Name = name

	gate.xorGate = NewXORGate(fmt.Sprintf("%s-XORGate", name), pin1, pin2) // having to make one as a named object so it can be Shutdown later (vs. just feeding NewXORGate(pin1, pin2) into NewInverter())
	gate.inverter = NewInverter(fmt.Sprintf("%s-Inverter", name), gate.xorGate)

	chState := make(chan Charge, 1)
	go func() {
		for {
			select {
			case c := <-chState:
				Debug(name, fmt.Sprintf("Received on Channel (%v), Charge {%s}", chState, c.String()))
				// putting this in a new go func() will allow any loopbacks triggered by the transmit, that end up feeding back into THIS gate, to not be blocked by the select/case
				go func(e Charge) {
					gate.Transmit(e)
					e.Done()
				}(c)
			case <-gate.chStop:
				Debug(name, "Stopped")
				return
			}
		}
	}()

	// in this approach to an XNOR (vs. building it with a combination of other gates), the Inverter owns the final output charge state
	gate.inverter.WireUp(chState)

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component and the gate itself, to exit
func (g *XNORGate) Shutdown() {
	g.xorGate.Shutdown()
	g.inverter.Shutdown()
	g.chStop <- true
}
