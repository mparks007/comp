package circuit

import (
	"fmt"
	"sync"

	"github.com/satori/go.uuid"
)

// Electron will be the the pimary means for indicating power flowing from component to component (and flagging if propogation of state change has ended)
type Electron struct {
	sender       string
	powerState   bool
	lockContexts []uuid.UUID
	//	lockContexts atomic.Value
	wg *sync.WaitGroup
	sync.RWMutex
}

// why can't I get an Electron owned field mutext to avoid the race?  had to use global...
// why can't I get an Electron owned field mutext to avoid the race?  had to use global...
// why can't I get an Electron owned field mutext to avoid the race?  had to use global...
// why can't I get an Electron owned field mutext to avoid the race?  had to use global...
// why can't I get an Electron owned field mutext to avoid the race?  had to use global...
var conLock sync.RWMutex

// AddContext will allow a component to indicate it will be involved in a lock of itself associated to the electron flow
func (e *Electron) AddContext(context uuid.UUID) {
	conLock.Lock()
	defer conLock.Unlock()
	e.lockContexts = append(e.lockContexts, context)
}

// HasContext will allow a component to check and see if it isn't already present in the electron's tracker of locked contexts (i.e. the components)...and therefore would be safe to lock itelf
func (e *Electron) HasContext(context uuid.UUID) bool {
	conLock.Lock()
	defer conLock.Unlock()
	for _, c := range e.lockContexts {
		if uuid.Equal(c, context) {
			return true
		}
	}
	return false
}

// String will display most of the Electron fields
func (e *Electron) String() string {
	conLock.Lock()
	defer conLock.Unlock()
	return fmt.Sprintf("sender (%s), powerState (%t), lockContexts (%v)", e.sender, e.powerState, e.lockContexts)
}

// Done will let the internal waitgroup know the processing for the Electron has finished (to allow the parent to 'unwind by one' in order to eventually finish the Transmit calls)
func (e *Electron) Done() {
	e.wg.Done()
}