package circuit

type inverter struct {
	in      emitter
	openOut emitter
}

func newInverter(pin emitter) *inverter {
	return &inverter{
		pin,
		newXContact(&battery{}, pin),
	}
}

func (i *inverter) Emitting() bool {
	return i.openOut.Emitting()
}
