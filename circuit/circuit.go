package circuit

type emitter interface {
	Emitting() bool
}

type Battery struct {
}

// Emitting on a Battery is always considered true (Battery never drains)
func (b *Battery) Emitting() bool {
	return true
}
