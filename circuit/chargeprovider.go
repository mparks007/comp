package circuit

// ChargeProvider is a low-tech charge giver to simply control/store/transmit charge state of on or off
type ChargeProvider struct {
	chargeSource
}

// NewChargeProvider will return a generic charge provider whose initial state is based on the passed in initialization value
func NewChargeProvider(name string, startState bool) *ChargeProvider {
	cp := &ChargeProvider{}
	cp.Init()
	cp.Name = name

	cp.hasCharge.Store(startState)
	return cp
}

// Charge will simulate a live charge by simply transmitting charge as on (and tracking a unique sequence number for all downstream results of THIS state change)
func (cp *ChargeProvider) Charge() {
	cp.Transmit(Charge{state: true})
}

// Discharge will simulate an inactive charge by simply transmitting charge as off (and tracking a unique sequence number for all downstream results of THIS state change)
func (cp *ChargeProvider) Discharge() {
	cp.Transmit(Charge{state: false})
}
