package circuit

import "fmt"

//import "fmt"

// AND
// 0 0 0
// 1 0 0
// 0 1 0
// 1 1 1

type ANDGate struct {
	relays []*Relay
	ch     chan bool
	pwrSource
}

func NewANDGate(pins ...pwrEmitter) *ANDGate {
	gate := &ANDGate{}

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewRelay(NewBattery(), pin))
		} else {
			gate.relays = append(gate.relays, NewRelay(&gate.relays[i-1].ClosedOut, pin))
		}
	}

	go func() {
		for {
			select {
			case state := <-gate.ch:
				fmt.Println("Gate transmit")
				gate.Transmit(state)
			}
		}
	}()

	// the last relay in the chain is the final answer for an AND
	gate.relays[len(pins)-1].ClosedOut.WireUp(gate.ch)

	return gate
}

func (g *ANDGate) UpdatePin(andPinNum, relayPinNum int, pin pwrEmitter) {
	if andPinNum < 1 || andPinNum > len(g.relays) {
		panic(fmt.Sprintf("Invalid gate pin number.  Input pin count (%d), requested pin (%d)", len(g.relays), andPinNum))
	}

	g.relays[andPinNum-1].UpdatePin(relayPinNum, pin)
}

/*
func NewSynchronizedANDGate(pins ...pwrEmitter) *ANDGate {
	gate := &ANDGate{}

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewRelay(NewBattery(), pin))
		} else {
			gate.relays = append(gate.relays, NewRelay(&gate.relays[i-1].ClosedOut, pin))
		}

		// every relay needs to be a part of the final answer on a sync'd ANDGate
		gate.relays[i].ClosedOut.WireUp(gate.powerUpdate)
	}

	return gate
}

func (g *ANDGate) powerUpdate(newState bool) {

	allRelaysUpdated := true

	// check see if all relays have had their power state updated
	for _, rel := range g.relays {
		if !rel.updated {
			allRelaysUpdated = false
			break
		}
	}

	if allRelaysUpdated {
		// reset each relay's updated state since we got our answer and are handling it
		for _, rel := range g.relays {
			rel.updated = false
		}

		// as an AND, transmit if the LAST relay in the chain has power at both A and B in
		g.Transmit(g.relays[len(g.relays)-1].aInPowered && g.relays[len(g.relays)-1].bInPowered)
	}
}
*/
/*
// OR
// 0 0 0
// 1 0 1
// 0 1 1
// 1 1 1

type ORGate struct {
	relays []*Relay
	pwrSource
}

func NewORGate(pins ...pwrEmitter) *ORGate {
	gate := &ORGate{}

	for i, pin := range pins {
		gate.relays = append(gate.relays, NewRelay(NewBattery(), pin))

		// every relay can trigger state in s chain of ORs
		gate.relays[i].ClosedOut.WireUp(gate.powerUpdate)
	}

	return gate
}

func (g *ORGate) powerUpdate(newState bool) {
	newState = false

	// check to see if ANY of the relays are closed
	for _, rel := range g.relays {
		if rel.ClosedOut.GetIsPowered() {
			newState = true
			break
		}
	}

	g.Transmit(newState)
}

// NAND
// 0 0 1
// 1 0 1
// 0 1 1
// 1 1 0

type NANDGate struct {
	relays []*Relay
	pwrSource
}

func NewNANDGate(pins ...pwrEmitter) *NANDGate {
	gate := &NANDGate{}

	for i, pin := range pins {
		gate.relays = append(gate.relays, NewRelay(NewBattery(), pin))

		// every relay can trigger state in a chain of NANDs
		gate.relays[i].OpenOut.WireUp(gate.powerUpdate)
	}

	return gate
}

func (g *NANDGate) powerUpdate(newState bool) {
	newState = false

	// check to see if ANY of the relays are open
	for _, rel := range g.relays {
		if rel.OpenOut.GetIsPowered() {
			newState = true
			break
		}
	}

	g.Transmit(newState)
}

// NOR
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 0

type NORGate struct {
	relays []*Relay
	pwrSource
}

func NewNORGate(pins ...pwrEmitter) *NORGate {
	gate := &NORGate{}

	for i, pin := range pins {
		if i == 0 {
			gate.relays = append(gate.relays, NewRelay(NewBattery(), pin))
		} else {
			gate.relays = append(gate.relays, NewRelay(&gate.relays[i-1].OpenOut, pin))
		}
	}

	// the last relay in the chain is the final answer for a NOR
	gate.relays[len(gate.relays)-1].OpenOut.WireUp(gate.Transmit)

	return gate
}

func (g *NORGate) UpdatePin(norPinNum, relayPinNum int, pin pwrEmitter) {
	if norPinNum < 1 || norPinNum > len(g.relays) {
		panic(fmt.Sprintf("Invalid gate pin number.  Input pin count (%d), requested pin (%d)", len(g.relays), norPinNum))
	}

	g.relays[norPinNum-1].UpdatePin(relayPinNum, pin)
}

// XOR
// 0 0 0
// 1 0 1
// 0 1 1
// 1 1 0

type XORGate struct {
	orGate   pwrEmitter
	nandGate pwrEmitter
	andGate  pwrEmitter
	pwrSource
}

func NewXORGate(pin1, pin2 pwrEmitter) *XORGate {
	gate := &XORGate{}

	gate.orGate = NewORGate(pin1, pin2)
	gate.nandGate = NewNANDGate(pin1, pin2)
	gate.andGate = NewANDGate(gate.orGate, gate.nandGate)

	// the state of the shared AND is the answer for an XOR
	gate.andGate.WireUp(gate.Transmit)

	return gate
}

// XNOR (aka equivalence gate) (using Inverter on an XOR gate)
// 0 0 1
// 1 0 0
// 0 1 0
// 1 1 1

type XNORGate struct {
	inverter pwrEmitter
	pwrSource
}

func NewXNORGate(pin1, pin2 pwrEmitter) *XNORGate {
	gate := &XNORGate{}

	gate.inverter = NewInverter(NewXORGate(pin1, pin2))

	// the inverter owns the final answer in this approach to an XNOR
	gate.inverter.WireUp(gate.Transmit)

	return gate
}
*/
