package circuit

type Battery struct {
	bitPublication
}

func NewBattery() *Battery {
	b := &Battery{}
	b.isPowered = true
	return b
}
func (b *Battery) Charge() {
	b.Publish(true)
}

func (b *Battery) Discharge() {
	b.Publish(false)
}

// OLD STUFF
// OLD STUFF
// OLD STUFF
// OLD STUFF
// OLD STUFF
// OLD STUFF
type emitter interface {
	Emitting() bool
}

func (b *Battery) Emitting() bool {
	return true
}
