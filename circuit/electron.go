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
}

// AddContext will allow a component to indicate it will be involved in a lock of itself associated to the electron flow
func (e *Electron) AddContext(context uuid.UUID) {
	e.lockContexts = append(e.lockContexts, context)
}

// HasContext will allow a component to check and see if it isn't already present in the electron's tracker of locked contexts (i.e. the components)...and therefore would be safe to lock itelf
func (e *Electron) HasContext(context uuid.UUID) bool {
	for _, c := range e.lockContexts {
		if uuid.Equal(c, context) {
			return true
		}
	}
	return false
}

// String will display most of the Electron fields
func (e *Electron) String() string {
	return fmt.Sprintf("sender (%s), powerState (%t), lockContexts (%v)", e.sender, e.powerState, e.lockContexts)
}

// Done will let the internal waitgroup know the processing for the Electron has finished (to allow the parent to 'unwind by one' in order to eventually finish the Transmit calls)
func (e *Electron) Done() {
	e.wg.Done()
}

/*
// HasContext will allow a component to check and see if it isn't already present in the electron's tracker of locked contexts (i.e. the components)...and therefore would be safe to lock itelf
func (e *Electron) HasContext(context uuid.UUID) bool {
	if _, ok := e.lockContexts.Load().([]uuid.UUID); !ok {
		e.lockContexts.Store(uuid.Must(uuid.NewV4()))
	}

	for i := 0; i < len(e.lockContexts.Load().([]uuid.UUID)); i++ {

		if uuid.Equal(e.lockContexts.Load().([]uuid.UUID)[i], context) {
			return true
		}
	}
	return false
}

// AddContext will allow a component to indicate it will be involved in a lock of itself associated to the electron flow
func (e *Electron) AddContext(context uuid.UUID) {
	if _, ok := e.lockContexts.Load().([]uuid.UUID); !ok {
		e.lockContexts.Store(uuid.Must(uuid.NewV4()))
	}
	e.lockContexts.Store(append(e.lockContexts.Load().([]uuid.UUID), context))
}

// String will display most of the Electron fields
func (e *Electron) String() string {
	locks, ok := e.lockContexts.Load().([]uuid.UUID)
	if !ok {
		e.lockContexts.Store(uuid.Must(uuid.NewV4()))
	}
	return fmt.Sprintf("sender (%s), powerState (%t), lockContexts (%v)", e.sender, e.powerState, locks)
}
*/
