package circuit

import (
	"fmt"
	"sync"

	uuid "github.com/satori/go.uuid"
)

// Charge will be the the pimary means for indicating current flowing from component to component (and flagging if propogation of state change has ended)
//
//	The reference to "flow" is used loosley as electric current and hole current are of differing directions and that electrons don't flow per se, but jump and bump.
//	It doesn't really matter overall as a closed circuit generally gets "hot" throughout, instantly.
//	For simplicity, I have the concept of a Charge object "moving through the wire", from one end to the other, propogating its charge to each component encountered
type Charge struct {
	sender       string
	state        bool
	lockContexts []uuid.UUID
	wg           *sync.WaitGroup
	mu           *sync.RWMutex
}

// AddContext will allow a component to indicate it will be involved in a lock of itself associated to the Charge flow
func (c *Charge) AddContext(context uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	Debug("Charge", fmt.Sprintf("(%v) adding context (%s)", &c, context))
	c.lockContexts = append(c.lockContexts, context)
}

// HasContext will allow a component to check and see if it isn't already present in the Charge's tracker of locked contexts (i.e. the components)...and therefore would be safe to lock itelf
func (c *Charge) HasContext(context uuid.UUID) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// is this context already in play?
	for _, c := range c.lockContexts {
		if uuid.Equal(c, context) {
			return true
		}
	}
	return false
}

// String will display most of the Charge struct fields
func (c *Charge) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return fmt.Sprintf("self (%p), sender (%s), state (%t), lockContexts (%v)", c, c.sender, c.state, c.lockContexts)
}

// Done will let the internal waitgroup know the processing for the Charge has finished (to allow the parent to 'unwind by one' in order to eventually finish the Transmit calls, sending the current down ALL paths)
func (c *Charge) Done() {
	c.wg.Done()
}
