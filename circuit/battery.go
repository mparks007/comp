package circuit

type Battery struct {
	publication
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
