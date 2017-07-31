package circuit

type pwrEmitter interface {
	WireUp(ch chan bool)
}

type Logger interface {
	Log(cat, data string) error
}

