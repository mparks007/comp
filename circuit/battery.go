package circuit

type Battery struct {
	pwrSource
}

func NewBattery() *Battery {
	bat := &Battery{}
	bat.isPowered = true
	return bat
}
func (b *Battery) Charge() {
	b.Transmit(true)
}

func (b *Battery) Discharge() {
	b.Transmit(false)
}
