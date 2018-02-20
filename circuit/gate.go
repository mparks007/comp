package circuit

// ANDGate is a standard AND logic gate
//	Wired like a NOR gate, but wired up to the CLOSED outs of each relay.
//
// Truth Table
// in in out
// 0  0   0
// 1  0   0
// 0  1   0
// 1  1   1
type ANDGate struct {
	relays    []*Relay // two or more relays to control the final gate state answer
	pwrSource          // switch gains all that is pwrSource too
}

// NewANDGate will return an AND gate whose inputs are set by the passed in pins
func NewANDGate(pins ...pwrEmitter) *ANDGate {
	gate := &ANDGate{}
	gate.Init()

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewRelay(NewBattery(true), pin))
		} else {
			gate.relays = append(gate.relays, NewRelay(&gate.relays[i-1].ClosedOut, pin))
		}
	}

	chState := make(chan Electron, 1)
	go func() {
		for {
			select {
			case e := <-chState:
				gate.Transmit(e.powerState)
				e.wg.Done()
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
	relays  []*Relay
	chStops []chan bool
	pwrSource
}

// NewORGate will return an OR gate whose inputs are set by the passed in pins
func NewORGate(pins ...pwrEmitter) *ORGate {
	gate := &ORGate{}
	gate.Init()

	// build a relay and associated listen/transmit func to deal with each input pin
	var chStates []chan Electron
	for i, pin := range pins {
		gate.relays = append(gate.relays, NewRelay(NewBattery(true), pin))

		chStates = append(chStates, make(chan Electron, 1))
		gate.chStops = append(gate.chStops, make(chan bool, 1))
		go func(index int) {
			for {
				select {
				case e := <-chStates[index]:
					gate.Transmit(e.powerState)
					e.wg.Done()
				case <-gate.chStops[index]:
					return
				}
			}
		}(i)

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
	relays  []*Relay
	chStops []chan bool
	pwrSource
}

// NewNANDGate will return a NAND gate whose inputs are set by the passed in pins
func NewNANDGate(pins ...pwrEmitter) *NANDGate {
	gate := &NANDGate{}
	gate.Init()

	// build a relay and associated listen/transmit func to deal with each input pin
	var chStates []chan Electron
	for i, pin := range pins {
		gate.relays = append(gate.relays, NewRelay(NewBattery(true), pin))

		chStates = append(chStates, make(chan Electron, 1))
		gate.chStops = append(gate.chStops, make(chan bool, 1))
		go func(index int) {
			for {
				select {
				case e := <-chStates[index]:
					gate.Transmit(e.powerState)
					e.wg.Done()
				case <-gate.chStops[index]:
					return
				}
			}
		}(i)

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

/*
// NORGate is a standard NOR (Not-OR) logic gate.
//	Wired like an AND gate, but wired up to the OPEN outs of each relay.
//
// Truth Table
// in in out
// 0  0   1
// 1  0   0
// 0  1   0
// 1  1   0
type NORGate struct {
	relays []*Relay
	ch     chan bool
	pwrSource
}

// NewNORGate will return a NOR gate whose inputs are set by the passed in pins
func NewNORGate(pins ...pwrEmitter) *NORGate {
	gate := &NORGate{}
	gate.Init()

	gate.ch = make(chan bool, 1)

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewRelay(NewBattery(true), pin))
		} else {
			gate.relays = append(gate.relays, NewRelay(&gate.relays[i-1].OpenOut, pin))
		}
	}

	// for a NOR, the last relay in the chain is the final answer (from OPEN out)
	gate.relays[len(pins)-1].OpenOut.WireUp(gate.ch)

	transmit := func() {
		gate.Transmit(<-gate.ch)
	}

	// calling transmit explicitly to ensure the 'answer' for the output, post WireUp above, has settled BEFORE returning and letting things wire up to it
	transmit()

	go func() {
		for {
			select {
			case <-gate.chStop:
				return
			default:
				transmit()
			}
		}
	}()

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
	orGate   *ORGate
	nandGate *NANDGate
	andGate  *ANDGate
	ch       chan bool
	pwrSource
}

// NewXORGate will return an XOR gate whose inputs are set by the passed in pins
func NewXORGate(pin1, pin2 pwrEmitter) *XORGate {
	gate := &XORGate{}
	gate.Init()

	gate.ch = make(chan bool, 1)

	gate.orGate = NewORGate(pin1, pin2)
	gate.nandGate = NewNANDGate(pin1, pin2)
	gate.andGate = NewANDGate(gate.orGate, gate.nandGate)

	// for an XOR, the state of the shared AND Gate is the answer
	gate.andGate.WireUp(gate.ch)

	transmit := func() {
		gate.Transmit(<-gate.ch)
	}

	// calling transmit explicitly to ensure the 'answer' for the output, post WireUp above, has settled BEFORE returning and letting things wire up to it
	transmit()

	go func() {
		for {
			select {
			case <-gate.chStop:
				return
			default:
				transmit()
			}
		}
	}()

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
	inverter *Inverter
	xorGate  *XORGate
	ch       chan bool
	pwrSource
}

// NewXNORGate will return an XNOR gate whose inputs are set by the passed in pins
func NewXNORGate(pin1, pin2 pwrEmitter) *XNORGate {
	gate := &XNORGate{}
	gate.Init()

	gate.ch = make(chan bool, 1)

	gate.xorGate = NewXORGate(pin1, pin2) // having to make one as a named object so it can be Shutdown later (vs. just feeding NewXORGate(pin1, pin2) into NewInverter())
	gate.inverter = NewInverter(gate.xorGate)

	// in this approach to an XNOR (vs. building it with a combination of other gates), the Inverter owns the final answer
	gate.inverter.WireUp(gate.ch)

	transmit := func() {
		gate.Transmit(<-gate.ch)
	}

	// calling transmit explicitly to ensure the 'answer' for the output, post WireUp above, has settled BEFORE returning and letting things wire up to it
	transmit()

	go func() {
		for {
			select {
			case <-gate.chStop:
				return
			default:
				transmit()
			}
		}
	}()

	return gate
}

// Shutdown will allow the go funcs, which are handling listen/transmit on each sub-component and the gate itself, to exit
func (g *XNORGate) Shutdown() {
	g.xorGate.Shutdown()
	g.inverter.Shutdown()
	g.chStop <- true
}
*/
