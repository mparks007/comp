package circuit

type Battery struct {
	publication
}

func (b *Battery) Charge() {
	b.state = true
	b.Publish()
}

func (b *Battery) Discharge() {
	b.state = false
	b.Publish()
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
