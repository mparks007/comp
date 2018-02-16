package circuit

// Battery is a low-tech power source to simply store/transmit power state of on or off
type Battery struct {
	pwrSource
}

// NewBattery will return a battery which defaults to charged (on)
func NewBattery() *Battery {
	bat := &Battery{}
	bat.isPowered = true
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
