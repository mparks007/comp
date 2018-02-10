package circuit

type pwrEmitter interface {
	WireUp(ch chan bool)
	Transmit(state bool) bool
}

type Logger interface {
	Log(cat, data string) error
}
