package circuit

import (
	"fmt"
	"sync"
)

var Debugging = false

// pwrEmitter allows a circuit component to take part in the power subscription/transmission process
type pwrEmitter interface {
	WireUp(ch chan Electron)
	Transmit(powerState bool)
}

// Logger allows a circuit component to take part in various logging techniques, specifying log category text and log details/data
type Logger interface {
	Log(cat, data string) error
}

// Electron will be the the pimary means for indicating power flowing from component to component (and flagging if propogation of state change has ended)
type Electron struct {
	powerState bool
	wg         *sync.WaitGroup
	Name       string
}

func Debug(text string) {
	if Debugging {
		fmt.Println(text)
	}
}
