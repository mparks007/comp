package circuit

type battery struct {
}

// Emitting on a battery is always considered true (battery never drains)
func (b *battery) Emitting() bool {
	return true
}
