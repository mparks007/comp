package circuit

type andContact struct {
	sources []emitter
}

func newANDContact(emitters ...emitter) *andContact {
	c := &andContact{}

	for _, e := range emitters {
		c.sources = append(c.sources, e)
	}

	return c
}

// Emitting returns true if ALL of its tracked power sources are emitting power
func (c *andContact) Emitting() bool {
	if len(c.sources) == 0 {
		return false
	}

	for _, s := range c.sources {
		if s == nil || !s.Emitting() {
			return false
		}
	}

	return true
}

type xContact struct {
	pwrSourceA emitter
	pwrSourceB emitter
}

func newXContact(a, b emitter) *xContact {
	return &xContact{a, b}
}

// Emitting returns true if ONLY source A is emitting power
func (x *xContact) Emitting() bool {
	return x.pwrSourceA != nil && x.pwrSourceA.Emitting() &&
		(x.pwrSourceB == nil ||
			(x.pwrSourceB != nil && !x.pwrSourceB.Emitting()))
}
