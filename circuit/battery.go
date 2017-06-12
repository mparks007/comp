package circuit

type Battery struct {
	pwrSource
}

func NewBattery() *Battery {
	b := &Battery{}
	b.isPowered = true
	return b
}
func (b *Battery) Charge() {
	b.Transmit(true)
}

func (b *Battery) Discharge() {
	b.Transmit(false)
}
