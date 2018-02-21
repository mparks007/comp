package circuit

// Battery is a low-tech power source to simply store/transmit power state of on or off
type Battery struct {
	pwrSource // battery gains all that is pwrSource too
}

// NewBattery will return a battery whose initial state is based on the passed in initialization value
func NewBattery(startState bool) *Battery {
	bat := &Battery{}
	bat.Init()
	bat.isPowered.Store(startState)
	return bat
}

// Charge will simulate a live battery by simply transmitting power as on
func (b *Battery) Charge() {
	b.Transmit(true)
}

// Discharge will simulate a dead battery by simply transmitting power as off
func (b *Battery) Discharge() {
	b.Transmit(false)
}
