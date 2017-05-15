package circuit

type battery struct {
}

func (b *battery) Emitting() bool {
	return true
}
