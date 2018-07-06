package circuit

import (
	"fmt"
	"time"
)

// C:\gospace\src\github.com\mparks007\comp> go install .\circuit\
// C:\gospace\src\github.com\mparks007\comp> godoc -http=":6060" // shows docs for all installed packages

// Debugging is the master flag to control very verbose logging to the console
var Debugging = false

// wireUpper allows a circuit component to wire up to a transmitter component in order to be told of the transmitter's power state
type wireUpper interface {
	WireUp(ch chan Electron)
}

// transmitter allows a circuit component to transmit it's power state to a wired up listener component
type transmitter interface {
	Transmit(e Electron)
}

// pwrEmitter allows a circuit component to take part in the power subscription/transmission process (what is a better name for this??)
type pwrEmitter interface {
	wireUpper
	transmitter
}

// Logger allows a circuit component to take part in various logging techniques, specifying log category text and log details/data
type Logger interface {
	Log(cat, data string) error
}

// Debug will write verbose state to the console (csv format: date/time,name,text)
func Debug(name, text string) {
	if Debugging {
		fmt.Printf("%v,%s,\"%s\"\n", time.Now().Format("01-02-2006 15:04:05.9999999"), name, text)
	}
}
