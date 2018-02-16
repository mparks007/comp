package circuit

// pwrEmitter allows a circuit component to take part in the power subscription/transmission process
type pwrEmitter interface {
	WireUp(ch chan bool)
	Transmit(state bool) //bool
}

// Logger allows a circuit component to take part in various logging techniques, specifying log category text and log details/data
type Logger interface {
	Log(cat, data string) error
}
