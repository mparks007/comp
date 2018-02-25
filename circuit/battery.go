package circuit

// Battery is a low-tech power source to simply store/transmit power state of on or off
type Battery struct {
	pwrSource // battery gains all that is pwrSource too
}

func NewBattery(startState bool) *Battery {
	return NewNamedBattery("?", startState)
}

// NewBattery will return a battery whose initial state is based on the passed in initialization value
func NewNamedBattery(name string, startState bool) *Battery {
	bat := &Battery{}
	bat.Init()
	bat.Name = name

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
